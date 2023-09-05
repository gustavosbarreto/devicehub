package workers

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
)

func StartHeartBeat(ctx context.Context) error {
	envs, err := getEnvs()
	if err != nil {
		return fmt.Errorf("failed to get the envs: %w", err)
	}

	connStr, err := connstring.ParseAndValidate(envs.MongoURI)
	if err != nil {
		log.WithError(err).Fatal("Invalid Mongo URI format")
	}

	clientOptions := options.Client().ApplyURI(envs.MongoURI)
	client, err := mongodriver.Connect(ctx, clientOptions)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to MongoDB")
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.WithError(err).Fatal("Failed to ping MongoDB")
	}

	store := mongo.NewStore(client.Database(connStr.Database), cache.NewNullCache())

	addr, err := asynq.ParseRedisURI(envs.RedisURI)
	if err != nil {
		return fmt.Errorf("failed to parse redis uri: %w", err)
	}

	aggregate := func(group string, tasks []*asynq.Task) *asynq.Task {
		var b strings.Builder

		for _, task := range tasks {
			b.Write(task.Payload())
			b.WriteString("\n")
		}

		return asynq.NewTask("api:heartbeat", []byte(b.String()))
	}

	srv := asynq.NewServer(
		addr,
		asynq.Config{
			GroupAggregator:  asynq.GroupAggregatorFunc(aggregate),
			GroupMaxDelay:    time.Duration(envs.AsynqGroupMaxDelay) * time.Second,
			GroupGracePeriod: time.Duration(envs.AsynqGroupGracePeriod) * time.Second,
			GroupMaxSize:     envs.AsynqGroupMaxSize,
			Queues:           map[string]int{"api": 6},
		},
	)

	mux := asynq.NewServeMux()

	mux.HandleFunc("api:heartbeat", func(ctx context.Context, task *asynq.Task) error {
		scanner := bufio.NewScanner(bytes.NewReader(task.Payload()))
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			store.DeviceSetOnline(ctx, models.UID(scanner.Text()), true) //nolint:errcheck
		}

		return nil
	})

	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}

	return nil
}
