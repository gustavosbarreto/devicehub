package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration90 = migrate.Migration{
	Version:     90,
	Description: "Add events field on sessions",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   90,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{
			"events":       bson.M{"$exists": false},
			"events.types": bson.M{"$exists": false},
			"events.items": bson.M{"$exists": false},
		}

		update := bson.M{
			"$set": bson.M{
				"events": bson.M{
					"types": bson.A{},
					"items": bson.A{},
					"seats": bson.A{0},
				},
			},
		}

		_, err := db.
			Collection("sessions").
			UpdateMany(ctx, filter, update)

		return err
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   90,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{}

		update := bson.M{
			"$unset": bson.M{
				"events": "",
			},
		}

		_, err := db.
			Collection("sessions").
			UpdateMany(ctx, filter, update)

		return err
	}),
}
