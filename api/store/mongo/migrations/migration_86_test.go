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

func TestMigration86(t *testing.T) {
	ctx := context.Background()

	mock := &envmock.Backend{}
	envs.DefaultBackend = mock

	tests := []struct {
		description string
		setup       func(t *testing.T)
		run         func(t *testing.T)
	}{
		{
			description: "Apply up on migration 83 when there is at least one user",
			setup: func(t *testing.T) {
				_, err := c.Database("test").Collection("tags").InsertOne(ctx, models.Tags{
					Name:   "red",
					Color:  "#ff0000",
					Tenant: "00000000-0000-4000-0000-000000000000",
				})
				require.NoError(t, err)
			},
			run: func(t *testing.T) {
				result := c.Database("test").Collection("tags").FindOne(ctx, bson.M{})
				require.NoError(t, result.Err())

				var tags models.Tags

				err := result.Decode(&tags)
				require.NoError(t, err)

				assert.Equal(t, "#ff0000", tags.Color)
				assert.Equal(t, "red", tags.Name)
				assert.Equal(t, "00000000-0000-4000-0000-000000000000", tags.Tenant)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(tt *testing.T) {
			tt.Cleanup(func() {
				assert.NoError(tt, srv.Reset())
			})

			migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[86-1])
			require.NoError(tt, migrates.Up(context.Background(), migrate.AllAvailable))

			test.setup(tt)

			test.run(tt)
		})
	}
}
