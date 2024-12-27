package migrations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration57(t *testing.T) {
	t.Cleanup(func() {
		assert.NoError(t, srv.Reset())
	})

	type PaymentFailed struct {
		Date    time.Time `json:"date" bson:"date,omitempty"`
		Details string    `json:"details" bson:"details,omitempty"`
		Amount  float64   `json:"amount" bson:"amount,omitempty"`
		Status  bool      `json:"status" bson:"status,omitempty"`
	}

	type Billing struct {
		CurrentPeriodEnd time.Time      `json:"current_period_end" bson:"current_period_end,omitempty"`
		PaymentFailed    *PaymentFailed `json:"payment_failed" bson:"payment_failed,omitempty"`
		SubscriptionID   string         `json:"subscription_id" bson:"subscription_id,omitempty"`
		PriceID          string         `json:"price_id" bson:"price_id,omitempty"`
		CustomerID       string         `json:"customer_id" bson:"customer_id,omitempty"`
		PaymentMethodID  string         `json:"payment_method_id" bson:"payment_method_id,omitempty"`
		State            string         `json:"state" bson:"state,omitempty"`
		SubItem          string         `json:"sub_item_id" bson:"sub_item_id,omitempty"`
		Active           bool           `json:"active" bson:"active,omitempty"`
	}

	type Namespace struct {
		Billing  *Billing `json:"billing" bson:"billing,omitempty"`
		TenantID string   `json:"tenant_id" bson:"tenant_id"`
	}

	cases := []struct {
		description string
		setup       func() error
		run         func() error
		check       func() (string, error)
		expected    string
	}{
		{
			description: "Success to apply up on migration 57 when namespace has billing",
			setup: func() error {
				_, err := c.Database("test").Collection("namespaces").InsertOne(context.TODO(), Namespace{
					TenantID: "00000000-0000-0000-0000-000000000001",
					Billing: &Billing{
						State: "processed",
					},
				})

				return err
			},
			run: func() error {
				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[56:57]...)

				return migrates.Up(context.Background(), migrate.AllAvailable)
			},
			check: func() (string, error) {
				namespace := new(models.Namespace)
				err := c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{
					"tenant_id": "00000000-0000-0000-0000-000000000001",
				}).Decode(&namespace)
				if err != nil {
					return "", err
				}

				return string(namespace.Billing.Status), nil
			},
			expected: "active",
		},
		{
			description: "Success to apply up on migration 57 when namespace has no billing",
			setup: func() error {
				_, err := c.Database("test").Collection("namespaces").InsertOne(context.TODO(), Namespace{
					TenantID: "00000000-0000-0000-0000-000000000002",
				})

				return err
			},
			run: func() error {
				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[56:57]...)

				return migrates.Up(context.Background(), migrate.AllAvailable)
			},
			check: func() (string, error) {
				namespace := new(models.Namespace)
				if err := c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{"tenant_id": "00000000-0000-0000-0000-000000000002"}).Decode(&namespace); err != nil {
					return "", err
				}

				if namespace.Billing != nil {
					return "", errors.New("billing should be nil")
				}

				return "", nil
			},
			expected: "",
		},
		{
			description: "Success to apply down on migration 57 when namespace has billing",
			setup: func() error {
				_, err := c.Database("test").Collection("namespaces").InsertOne(context.TODO(), &models.Namespace{
					TenantID: "00000000-0000-0000-0000-000000000003",
					Billing: &models.Billing{
						Status: "active",
					},
				})

				return err
			},
			run: func() error {
				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[56:57]...)

				return migrates.Down(context.Background(), migrate.AllAvailable)
			},
			check: func() (string, error) {
				namespace := new(Namespace)
				err := c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{
					"tenant_id": "00000000-0000-0000-0000-000000000003",
				}).Decode(&namespace)
				if err != nil {
					return "", err
				}

				return namespace.Billing.State, nil
			},
			expected: "processed",
		},
		{
			description: "Success to apply down on migration 57 when namespace has no billing",
			setup: func() error {
				_, err := c.Database("test").Collection("namespaces").InsertOne(context.TODO(), &models.Namespace{
					TenantID: "00000000-0000-0000-0000-000000000004",
				})

				return err
			},
			run: func() error {
				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[56:57]...)

				return migrates.Down(context.Background(), migrate.AllAvailable)
			},
			check: func() (string, error) {
				namespace := new(Namespace)
				err := c.Database("test").Collection("namespaces").FindOne(context.TODO(), bson.M{
					"tenant_id": "00000000-0000-0000-0000-000000000004",
				}).Decode(&namespace)
				if err != nil {
					return "", err
				}

				if namespace.Billing != nil {
					return "", errors.New("billing should be nil")
				}

				return "", nil
			},
			expected: "",
		},
	}

	for _, test := range cases {
		tc := test
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, tc.setup())
			assert.NoError(t, tc.run())

			result, err := tc.check()
			assert.Equal(t, tc.expected, result)
			assert.NoError(t, err)
		})
	}
}
