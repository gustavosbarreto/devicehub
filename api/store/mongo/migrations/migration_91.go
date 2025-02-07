package migrations

import (
	"context"

	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration91 = migrate.Migration{
	Version:     91,
	Description: "Adding seat and seats to sessions and event's session",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   91,
			"action":    "Up",
		}).Info("Applying migration")

		filter := bson.M{}

		update := bson.M{
			"$set": bson.M{
				"events.items.$[].seat": 0,
				"events.seats":          bson.A{0},
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
			"version":   91,
			"action":    "Down",
		}).Info("Reverting migration")

		filter := bson.M{}

		update := bson.M{
			"$unset": bson.M{
				"events.items.$[].seat": "",
				"events.seats":          "",
			},
		}

		_, err := db.
			Collection("sessions").
			UpdateMany(ctx, filter, update)

		return err
	}),
}
