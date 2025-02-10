package migrations

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/pkg/envs"
	envmock "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	migrate "github.com/xakep666/mongo-migrate"
	"go.mongodb.org/mongo-driver/bson"
)

func TestMigration91Up(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
	}{
		{
			description: "Success to apply up on migration 91",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, map[string]interface{}{
						"events": map[string]interface{}{
							"types": []string{},
							"items": []map[string]interface{}{},
							"seats": []int{},
						},
					})

				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			assert.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[89])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("sessions").
				FindOne(context.TODO(), bson.M{})

			session := make(map[string]interface{})
			require.NoError(tt, query.Decode(&session))

			events, ok := session["events"].(map[string]interface{})
			require.True(tt, ok)
			require.Contains(tt, events, "seats")
		})
	}
}

func TestMigration91Down(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	cases := []struct {
		description string
		setup       func() error
	}{
		{
			description: "Success to revert migration 91",
			setup: func() error {
				_, err := c.
					Database("test").
					Collection("sessions").
					InsertOne(ctx, models.Session{
						Events: models.SessionEvents{
							Items: []models.SessionEvent{
								{
									Seat: 0,
								},
							},
							Seats: []int{0},
						},
					})

				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			assert.NoError(tt, tc.setup())

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[90])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))
			require.NoError(tt, migrates.Down(context.Background(), migrate.AllAvailable))

			query := c.
				Database("test").
				Collection("sessions").
				FindOne(context.TODO(), bson.M{})

			session := make(map[string]interface{})
			require.NoError(tt, query.Decode(&session))

			events, ok := session["events"].(map[string]interface{})
			require.True(tt, ok)
			require.NotContains(tt, events, "seats")
		})
	}
}
