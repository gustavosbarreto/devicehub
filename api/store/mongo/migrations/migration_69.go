package migrations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/shellhub-io/shellhub/pkg/models"
	log "github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration69 = migrate.Migration{
	Version:     69,
	Description: "Hash API key ID. It will delete the old document and create a new one with the hashed ID.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   69,
				"action":    "Up",
			}).
			Info("Applying migration")

		pipeline := []bson.M{
			{
				"$match": bson.M{
					"_id": bson.M{
						"$regex": "^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-4[a-fA-F0-9]{3}-[8|9|aA|bB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$",
					},
				},
			},
		}

		cursor, err := db.Collection("api_keys").Aggregate(ctx, pipeline)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)

		updateModels := make([]mongo.WriteModel, 0)
		deleteModels := make([]mongo.WriteModel, 0)

		for cursor.Next(ctx) {
			apiKey := new(models.APIKey)
			if err := cursor.Decode(apiKey); err != nil {
				return err
			}

			idSum := sha256.Sum256([]byte(apiKey.ID))
			hashedID := hex.EncodeToString(idSum[:])

			doc := &models.APIKey{
				ID:        hashedID,
				Name:      apiKey.Name,
				CreatedBy: apiKey.CreatedBy,
				TenantID:  apiKey.TenantID,
				Role:      apiKey.Role,
				CreatedAt: apiKey.CreatedAt,
				UpdatedAt: apiKey.UpdatedAt,
				ExpiresIn: apiKey.ExpiresIn,
			}

			deleteModels = append(deleteModels, mongo.NewDeleteOneModel().SetFilter(bson.M{"_id": apiKey.ID}))
			updateModels = append(updateModels, mongo.NewInsertOneModel().SetDocument(doc))
		}

		if len(updateModels) > 0 || len(deleteModels) > 0 {
			mongoSession, err := db.Client().StartSession()
			if err != nil {
				return err
			}
			defer mongoSession.EndSession(ctx)

			_, err = mongoSession.WithTransaction(ctx, func(_ mongo.SessionContext) (interface{}, error) {
				if _, err := db.Collection("api_keys").BulkWrite(ctx, updateModels); err != nil {
					return nil, err
				}

				_, err := db.Collection("api_keys").BulkWrite(ctx, deleteModels)

				return nil, err
			})
			if err != nil {
				return err
			}
		}

		return nil
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		log.
			WithFields(log.Fields{
				"component": "migration",
				"version":   69,
				"action":    "Down",
			}).
			Info("Applying migration")

		log.Info("Unable to undo the api key hash")

		return nil
	}),
}
