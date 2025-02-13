package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var migration90 = migrate.Migration{
	Version:     90,
	Description: "Refactor tags structure in a separeted collection.",
	Up: migrate.MigrationFunc(func(ctx context.Context, db *mongo.Database) error {
		logrus.WithFields(logrus.Fields{
			"component": "migration",
			"version":   90,
			"action":    "Up",
		}).Info("Applying migration up")

		session, err := db.Client().StartSession()
		if err != nil {
			return err
		}
		defer session.EndSession(ctx)

		_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
			cursor, err := db.Collection("devices").Find(sessCtx, bson.M{"tags": bson.M{"$ne": []string{}}})
			if err != nil {
				return nil, err
			}
			defer cursor.Close(sessCtx)

			type device struct {
				UID      string   `bson:"uid"`
				TenantID string   `bson:"tenant_id"`
				Tags     []string `bson:"tags"`
			}

			tagMapping := make(map[string]map[string]string) // tenant_id -> tag_name -> tag_id
			for cursor.Next(sessCtx) {
				d := new(device)
				if err := cursor.Decode(d); err != nil {
					return nil, err
				}

				if _, ok := tagMapping[d.TenantID]; !ok {
					tagMapping[d.TenantID] = make(map[string]string)
				}

				tagIDs := make([]string, 0)
				for _, tagName := range d.Tags {
					id := primitive.NewObjectID()
					data := bson.M{
						"$setOnInsert": bson.M{"_id": id, "created_at": clock.Now(), "updated_at": clock.Now()},
						"$set":         bson.M{"name": tagName, "tenant_id": d.TenantID},
					}

					_, err := db.
						Collection("tags").
						UpdateOne(sessCtx, bson.M{"tenant_id": d.TenantID, "name": tagName}, data, options.Update().SetUpsert(true))
					if err != nil {
						return nil, err
					}

					tagIDs = append(tagIDs, id.String())
				}

				if _, err = db.Collection("devices").UpdateOne(sessCtx, bson.M{"uid": d.UID}, bson.M{"$set": bson.M{"tags": tagIDs}}); err != nil {
					return nil, err
				}
			}

			if err := cursor.Err(); err != nil {
				return nil, err
			}

			for tenantID, tagNameToID := range tagMapping {
				for tagName, tagID := range tagNameToID {
					for _, collection := range []string{"public_keys", "firewall_ruless"} {
						_, err := db.Collection(collection).UpdateMany(
							sessCtx,
							bson.M{
								"tenant_id":    tenantID,
								"filters.tags": tagName,
							},
							bson.M{
								"$set": bson.M{
									"filters.tags.$[elem]": tagID,
								},
							},
							&options.UpdateOptions{
								ArrayFilters: &options.ArrayFilters{
									Filters: []interface{}{
										bson.M{"elem": tagName},
									},
								},
							},
						)
						if err != nil {
							return nil, err
						}
					}
				}
			}

			return nil, nil
		})

		return err
	}),
	Down: migrate.MigrationFunc(func(_ context.Context, _ *mongo.Database) error {
		return nil
	}),
}
