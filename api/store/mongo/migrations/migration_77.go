package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/envs"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration77 = migrate.Migration{
	Version:     77,
	Description: "Adding VPN settings to namespace",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   77,
			"action":    "Up",
		}).Info("Applying migration")

		if envs.IsEnterprise() {
			update := bson.M{
				"$set": bson.M{
					"vpn": bson.M{
						"enable":  false,
						"address": bson.A{10, 0, 0, 0},
						"mask":    16,
					},
				},
			}

			_, err := db.
				Collection("namespaces").
				UpdateMany(ctx, bson.M{}, update)

			return err
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   77,
			"action":    "Down",
		}).Info("Reverting migration")

		if envs.IsEnterprise() {
			update := bson.M{
				"$unset": bson.M{"vpn": ""},
			}

			_, err := db.
				Collection("namespaces").
				UpdateMany(ctx, bson.M{}, update)

			return err
		}

		return nil
	}),
}
