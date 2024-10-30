package main

import (
	"context"

	"github.com/shellhub-io/shellhub/adm/cmd"
	"github.com/shellhub-io/shellhub/adm/services"
	"github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/shellhub-io/shellhub/pkg/loglevel"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type config struct {
	MongoURI string `env:"MONGO_URI,default=mongodb://mongo:27017/main"`
	RedisURI string `env:"REDIS_URI,default=redis://redis:6379"`
}

func init() {
	loglevel.SetLogLevel()
}

func main() {
	ctx := context.Background()

	cfg, err := envs.ParseWithPrefix[config]("CLI_")
	if err != nil {
		log.Error(err.Error())
	}

	log.Info("Connecting to Redis")

	cache, err := cache.NewRedisCache(cfg.RedisURI, 0)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Connected to Redis")

	log.Trace("Connecting to MongoDB")

	_, db, err := mongo.Connect(ctx, cfg.MongoURI)
	if err != nil {
		log.
			WithError(err).
			Fatal("unable to connect to MongoDB")
	}

	store, err := mongo.NewStore(ctx, db, cache)
	if err != nil {
		log.
			WithError(err).
			Fatal("failed to create the store")
	}

	service := services.NewService(store)

	rootCmd := &cobra.Command{Use: "cli"}
	rootCmd.AddCommand(cmd.UserCommands(service))
	rootCmd.AddCommand(cmd.NamespaceCommands(service))
	// WARN: this is deprecated and will be removed soon
	cmd.DeprecatedCommands(rootCmd, service)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
	}
}
