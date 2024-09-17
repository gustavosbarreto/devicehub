package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/getsentry/sentry-go"
	"github.com/shellhub-io/shellhub/api/routes"
	"github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/api/store/mongo/options"
	"github.com/shellhub-io/shellhub/pkg/api/internalclient"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/geoip"
	"github.com/shellhub-io/shellhub/pkg/worker/asynq"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use: "server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(cmd.Context())

		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		cfg, ok := ctx.Value("cfg").(*config)
		if !ok {
			log.Fatal("Failed to retrieve environment config from context")
		}

		log.Trace("Connecting to Redis")

		cache, err := storecache.NewRedisCache(cfg.RedisURI, cfg.RedisCachePoolSize)
		if err != nil {
			log.WithError(err).Error("Failed to configure redis store cache")
		}

		log.Info("Connected to Redis")

		log.Trace("Connecting to MongoDB")

		_, db, err := mongo.Connect(ctx, cfg.MongoURI)
		if err != nil {
			log.
				WithError(err).
				Fatal("unable to connect to MongoDB")
		}

		store, err := mongo.NewStore(ctx, db, cache, options.RunMigatrions)
		if err != nil {
			log.
				WithError(err).
				Fatal("failed to create the store")
		}

		log.Info("Connected to MongoDB")

		go func() {
			sig := <-sigs

			log.WithFields(log.Fields{
				"signal": sig,
			}).Info("signal received to terminate API")

			cancel()
		}()

		return startServer(ctx, cfg, store, cache)
	},
}

// Provides the configuration for the API service.
// The values are load from the system environment variables.
type config struct {
	// MongoDB connection string (URI format)
	MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`
	// Redis connection string (URI format)
	RedisURI string `env:"REDIS_URI,default=redis://redis:6379"`
	// RedisCachePoolSize is the pool size of connections available for Redis cache.
	RedisCachePoolSize int `env:"REDIS_CACHE_POOL_SIZE,default=0"`
	// Enable GeoIP feature.
	//
	// GeoIP features enable the ability to get the logitude and latitude of the client from the IP address.
	// The feature is disabled by default. To enable it, it is required to have a `MAXMIND` database license and feed it
	// to `SHELLHUB_MAXMIND_LICENSE` with it, and `SHELLHUB_GEOIP=true`.
	GeoIP               bool   `env:"GEOIP,default=false"`
	GeoIPMaxMindLicense string `env:"MAXMIND_LICENSE,default="`
	// Session record cleanup worker schedule
	// Sentry DSN.
	SentryDSN string `env:"SENTRY_DSN,default="`
	// AsynqGroupMaxDelay is the maximum duration to wait before processing a group of tasks.
	//
	// Its time unit is second.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupMaxDelay int `env:"ASYNQ_GROUP_MAX_DELAY,default=1"`
	// AsynqGroupGracePeriod is the grace period has configurable upper bound: you can set a maximum aggregation delay, after which Asynq server
	// will aggregate the tasks regardless of the remaining grace period.
	///
	// Its time unit is second.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupGracePeriod int64 `env:"ASYNQ_GROUP_GRACE_PERIOD,default=2"`
	// AsynqGroupMaxSize is the maximum number of tasks that can be aggregated together. If that number is reached, Asynq
	// server will aggregate the tasks immediately.
	//
	// Check [https://github.com/hibiken/asynq/wiki/Task-aggregation] for more information.
	AsynqGroupMaxSize int `env:"ASYNQ_GROUP_MAX_SIZE,default=1000"`
}

// startSentry initializes the Sentry client.
//
// The Sentry client is used to report errors to the Sentry server, and is initialized only if the `SHELLHUB_SENTRY_DSN`
// environment variable is set. Else, the function returns a error with a not initialized Sentry client.
func startSentry(dsn string) (*sentry.Client, error) {
	if dsn != "" {
		var err error
		reporter, err := sentry.NewClient(sentry.ClientOptions{ //nolint:exhaustruct
			Dsn:              dsn,
			Release:          os.Getenv("SHELLHUB_VERSION"),
			EnableTracing:    true,
			TracesSampleRate: 1,
		})
		if err != nil {
			log.WithError(err).Error("Failed to create Sentry client")

			return nil, err
		}
		log.Info("Sentry client started")

		return reporter, nil
	}

	return nil, errors.New("sentry DSN not provided")
}

func startServer(ctx context.Context, cfg *config, store store.Store, cache storecache.Cache) error {
	log.Info("Starting API server")

	apiClient, err := internalclient.NewClient(internalclient.WithAsynqWorker(cfg.RedisURI))
	if err != nil {
		log.WithError(err).
			Fatal("failed to create the internalclient")
	}

	servicesOptions := []services.Option{}

	if cfg.GeoIP {
		log.Info("GeoIP feature is enable")

		locator, err := geoip.NewGeoLite2(cfg.GeoIPMaxMindLicense)
		if err != nil {
			log.WithError(err).Fatal("Failed to init GeoIP")
		}

		servicesOptions = append(servicesOptions, services.WithLocator(locator))
	}

	service := services.NewService(store, nil, nil, cache, apiClient, servicesOptions...)

	routerOptions := []routes.Option{}

	if cfg.SentryDSN != "" {
		log.Info("Sentry report is enabled")

		reporter, err := startSentry(cfg.SentryDSN)
		if err != nil {
			log.WithField("DSN", cfg.SentryDSN).WithError(err).Warn("Failed to start Sentry")
		} else {
			log.Info("Sentry client started")
		}

		routerOptions = append(routerOptions, routes.WithReporter(reporter))
	}

	worker := asynq.NewServer(cfg.RedisURI, asynq.BatchConfig(cfg.AsynqGroupMaxSize, cfg.AsynqGroupMaxDelay, int(cfg.AsynqGroupGracePeriod)))
	worker.HandleTask(services.TaskDevicesHeartbeat, service.DevicesHeartbeat(), asynq.BatchTask())

	if err := worker.Start(); err != nil {
		log.WithError(err).
			Fatal("failed to start the worker")
	}

	router := routes.NewRouter(service, routerOptions...)

	go func() {
		<-ctx.Done()

		log.Debug("Closing HTTP server due context cancellation")

		worker.Shutdown()
		router.Close()
	}()

	err = router.Start(":8080") //nolint:errcheck
	log.WithError(err).Info("HTTP server closed")

	return nil
}
