package migrations

// import (
// 	"context"
// 	"sort"
// 	"testing"
//
// 	"github.com/shellhub-io/shellhub/pkg/models"
// 	"github.com/stretchr/testify/assert"
// 	migrate "github.com/xakep666/mongo-migrate"
// 	"go.mongodb.org/mongo-driver/bson"
// )
//
// func TestMigration46(t *testing.T) {
// 	cases := []struct {
// 		description string
// 		Test        func(t *testing.T)
// 	}{
// 		{
// 			"Success to apply up on migration 46",
// 			func(t *testing.T) {
// 				t.Helper()
//
// 				keyUsernameEmpty := &models.PublicKey{
// 					Fingerprint: "fingerprint",
// 					TenantID:    "tenant",
// 					PublicKeyFields: models.PublicKeyFields{
// 						Name:     "key",
// 						Username: "",
// 						Filter: models.PublicKeyFilter{
// 							Tags: []string{"tag1", "tag2", "tag3"},
// 						},
// 					},
// 				}
//
// 				keyUsernameRegexp := &models.PublicKey{
// 					Fingerprint: "fingerprint",
// 					TenantID:    "tenant",
// 					PublicKeyFields: models.PublicKeyFields{
// 						Name:     "key",
// 						Username: ".*",
// 						Filter: models.PublicKeyFilter{
// 							Tags: []string{"tag1", "tag2", "tag3"},
// 						},
// 					},
// 				}
//
// 				_, err := c.Database("test").Collection("public_keys").InsertOne(context.Background(), keyUsernameEmpty)
// 				assert.NoError(t, err)
//
// 				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[45:46]...)
// 				assert.NoError(t, migrates.Up(context.Background(), migrate.AllAvailable))
//
// 				key := new(models.PublicKey)
// 				result := c.Database("test").Collection("public_keys").FindOne(context.Background(), bson.M{"tenant_id": keyUsernameEmpty.TenantID})
// 				assert.NoError(t, result.Err())
//
// 				err = result.Decode(key)
// 				assert.NoError(t, err)
//
// 				sort.Strings(key.Filter.Tags)
//
// 				assert.Equal(t, keyUsernameRegexp, key)
// 			},
// 		},
// 		{
// 			"Success to apply down on migration 46",
// 			func(t *testing.T) {
// 				t.Helper()
//
// 				keyUsernameEmpty := &models.PublicKey{
// 					Fingerprint: "fingerprint",
// 					TenantID:    "tenant",
// 					PublicKeyFields: models.PublicKeyFields{
// 						Name:     "key",
// 						Username: "",
// 						Filter: models.PublicKeyFilter{
// 							Tags: []string{"tag1", "tag2", "tag3"},
// 						},
// 					},
// 				}
//
// 				keyUsernameRegexp := &models.PublicKey{
// 					Fingerprint: "fingerprint",
// 					TenantID:    "tenant",
// 					PublicKeyFields: models.PublicKeyFields{
// 						Name:     "key",
// 						Username: ".*",
// 						Filter: models.PublicKeyFilter{
// 							Tags: []string{"tag1", "tag2", "tag3"},
// 						},
// 					},
// 				}
//
// 				_, err := c.Database("test").Collection("public_keys").InsertOne(context.Background(), keyUsernameEmpty)
// 				assert.NoError(t, err)
//
// 				migrates := migrate.NewMigrate(c.Database("test"), GenerateMigrations()[45:46]...)
// 				assert.NoError(t, migrates.Down(context.Background(), migrate.AllAvailable))
//
// 				key := new(models.PublicKey)
// 				result := c.Database("test").Collection("public_keys").FindOne(context.Background(), bson.M{"tenant_id": keyUsernameRegexp.TenantID})
// 				assert.NoError(t, result.Err())
//
// 				err = result.Decode(key)
// 				assert.NoError(t, err)
//
// 				assert.Equal(t, keyUsernameEmpty, key)
// 			},
// 		},
// 	}
//
// 	for _, tc := range cases {
// 		t.Run(tc.description, func(t *testing.T) {
// 			t.Cleanup(func() {
// 				assert.NoError(t, srv.Reset())
// 			})
// 			tc.Test(t)
// 		})
// 	}
// }
