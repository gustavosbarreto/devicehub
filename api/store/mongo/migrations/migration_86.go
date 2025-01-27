package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var migration86 = migrate.Migration{
	Version:     86,
	Description: "Creating Tag collection.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   83,
			"action":    "Up",
		}).Info("Applying migration")

		if err := db.CreateCollection(ctx, "tags"); err != nil {
			return err
		}

		indexName := mongo.IndexModel{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetName("name").SetUnique(false),
		}

		_, err := db.Collection("tags",
			options.Collection().SetWriteConcern(writeconcern.Majority()),
		).Indexes().CreateOne(ctx, indexName)
		if err != nil {
			return err
		}

		indexTenant := mongo.IndexModel{
			Keys:    bson.D{{Key: "tenant_id", Value: 1}},
			Options: options.Index().SetName("tenant_id").SetUnique(false),
		}

		_, err2 := db.Collection("tags",
			options.Collection().SetWriteConcern(writeconcern.Majority()),
		).Indexes().CreateOne(ctx, indexTenant)
		if err2 != nil {
			return err2
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   86,
			"action":    "Down",
		}).Info("Reverting migration")

		_, err := db.Collection("tags",
			options.Collection().SetWriteConcern(writeconcern.Majority()),
		).Indexes().DropOne(ctx, "names")
		if err != nil {
			return err
		}

		_, err2 := db.Collection("tags",
			options.Collection().SetWriteConcern(writeconcern.Majority()),
		).Indexes().DropOne(ctx, "tenant_id")

		if err2 != nil {
			return err
		}

		return db.Collection("tags",
			options.Collection().SetWriteConcern(writeconcern.Majority()),
		).Drop(ctx)
	}),
}
