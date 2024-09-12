package asynq

import (
	"runtime"
	"time"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/pkg/uuid"
	"github.com/shellhub-io/shellhub/pkg/worker"
)

type ServerOption func(s *server) error

// BatchConfig sets the batch configuration of the server. It's required when
// setting a task with [BatchTask] option.
//
// maxSize is the maximum number of tasks that a batch task can handle before
// processing.
//
// maxDelay is the maximum amount of time that a batch task can wait before
// processing.
//
// gracePeriod is the amount of time that the server will wait before aggregating
// batch tasks.
func BatchConfig(maxSize, maxDelay, gracePeriod int) ServerOption {
	return func(s *server) error {
		s.batchConfig.maxSize = maxSize
		s.batchConfig.maxDelay = time.Second * time.Duration(maxDelay)
		s.batchConfig.gracePeriod = time.Second * time.Duration(gracePeriod)

		return nil
	}
}

type server struct {
	redisURI    string
	asynqSrv    *asynq.Server
	asynqMux    *asynq.ServeMux
	asynqSch    *asynq.Scheduler
	batchConfig *batchConfig

	queues   queues
	tasks    []worker.Task
	cronjobs []worker.Cronjob
}

func NewServer(redisURI string, opts ...ServerOption) worker.Server {
	s := &server{
		redisURI:    redisURI,
		queues:      queues{cronQueue: 1},
		tasks:       []worker.Task{},
		cronjobs:    []worker.Cronjob{},
		batchConfig: &batchConfig{},
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil // NOTE: currently all opts returns nil
		}
	}

	return s
}

func (s *server) HandleTask(pattern worker.TaskPattern, handler worker.TaskHandler, opts ...worker.TaskOption) {
	pattern.MustValidate()

	if _, ok := s.queues[pattern.Queue()]; !ok {
		s.queues[pattern.Queue()] = 1
	}

	task := worker.Task{Pattern: pattern, Handler: handler}
	for _, opt := range opts {
		opt(&task)
	}

	s.tasks = append(s.tasks, task)
}

func (s *server) HandleCron(spec worker.CronSpec, handler worker.CronHandler) {
	spec.MustValidate()

	cronjob := worker.Cronjob{
		Identifier: uuid.Generate(),
		Spec:       spec,
		Handler:    handler,
	}

	s.cronjobs = append(s.cronjobs, cronjob)
}

func (s *server) Start() error {
	fmt.Println("WORKER:: starting worker")
	if err := s.setupAsynq(); err != nil {
		fmt.Println("fail at setup : worker")
		return err
	}

	if err := s.asynqSrv.Start(s.asynqMux); err != nil {
		fmt.Println("WORKER:: fail at start async mux : worker")
		return worker.ErrServerStartFailed
	}

	if err := s.asynqSch.Start(); err != nil {
		fmt.Println("WORKER:: fail at start : worker")
		return worker.ErrServerStartFailed
	}

	return nil
}

func (s *server) Shutdown() {
	s.asynqSrv.Shutdown()
	s.asynqSch.Shutdown()
}

func (s *server) setupAsynq() error {
	fmt.Println("WORKER:: start Setup Asynq")
	addr, err := asynq.ParseRedisURI(s.redisURI)
	if err != nil {
		fmt.Println("WORKER:: ERROR at redis URI")
		return err
	}

	s.asynqSch = asynq.NewScheduler(addr, nil)
	s.asynqMux = asynq.NewServeMux()
	s.asynqSrv = asynq.NewServer(
		addr,
		asynq.Config{ //nolint:exhaustruct
			Concurrency:      runtime.NumCPU(),
			Queues:           s.queues,
			GroupAggregator:  asynq.GroupAggregatorFunc(aggregate),
			GroupMaxSize:     s.batchConfig.maxSize,
			GroupMaxDelay:    s.batchConfig.maxDelay,
			GroupGracePeriod: s.batchConfig.gracePeriod,
		},
	)

	fmt.Println("WORKER:: run tasks")
	for _, t := range s.tasks {
		s.asynqMux.HandleFunc(t.Pattern.String(), taskToAsynq(t.Handler))
	}

	fmt.Println("WORKER:: run cronjobs")
	for _, c := range s.cronjobs {
		s.asynqMux.HandleFunc(c.Identifier, cronToAsynq(c.Handler))
		task := asynq.NewTask(c.Identifier, nil, asynq.Queue(cronQueue))
		if _, err := s.asynqSch.Register(c.Spec.String(), task); err != nil {
			fmt.Println("WORKER:: ERROR : register new task")
			return worker.ErrHandleCronFailed
		}
	}
	fmt.Println("WORKER:: ended setup asyncq")

	return nil
}
