package migrations

import (
	"context"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/sirupsen/logrus"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var migration20 = migrate.Migration{
	Version: 20,
	Up: func(db *mongo.Database) error {
		logrus.Info("Applying migration 20 - Up")
		cursor, err := db.Collection("firewall_rules").Find(context.TODO(), bson.D{})
		if err != nil {
			return err
		}

		type firewallRule struct {
			ID                        primitive.ObjectID `json:"id" bson:"_id"`
			TenantID                  string             `json:"tenant_id" bson:"tenant_id"`
			models.FirewallRuleFields `bson:",inline"`
		}

		defer cursor.Close(context.TODO())
		for cursor.Next(context.TODO()) {
			firewall := new(models.FirewallRule)
			err := cursor.Decode(&firewall)
			if err != nil {
				return err
			}
			objID, err := primitive.ObjectIDFromHex(firewall.ID)
			replacedRule := firewallRule{
				TenantID:           firewall.TenantID,
				ID:                 objID,
				FirewallRuleFields: firewall.FirewallRuleFields,
			}

			if err == nil {
				if errDelete := db.Collection("firewall_rules").FindOneAndDelete(context.TODO(), bson.M{"_id": firewall.ID}); errDelete.Err() != nil {
					continue
				}

				if _, err := db.Collection("firewall_rules").InsertOne(context.TODO(), replacedRule); err != nil {
					return err
				}
			}
		}

		return nil
	},

	Down: func(db *mongo.Database) error {
		logrus.Info("Applying migration 20 - Down")

		return nil
	},
}
