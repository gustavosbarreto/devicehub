package mongo

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/dbtest"
	"github.com/shellhub-io/shellhub/api/pkg/fixtures"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestPublicKeyGet(t *testing.T) {
	type Expected struct {
		pubKey *models.PublicKey
		err    error
	}

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				pubKey: &models.PublicKey{
					Data:        []byte("test"),
					CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Fingerprint: "fingerprint",
					TenantID:    "00000000-0000-4000-0000-000000000000",
					PublicKeyFields: models.PublicKeyFields{
						Name: "public_key",
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
							Tags:     []string{"tag-1"},
						},
					},
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			pubKey, err := mongostore.PublicKeyGet(context.TODO(), tc.fingerprint, tc.tenant)
			assert.Equal(t, tc.expected, Expected{pubKey: pubKey, err: err})
		})
	}
}

func TestPublicKeyList(t *testing.T) {
	type Expected struct {
		pubKey []models.PublicKey
		len    int
		err    error
	}

	cases := []struct {
		description string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when public key list is empty",
			fixtures:    []string{},
			expected: Expected{
				pubKey: []models.PublicKey{},
				len:    0,
				err:    nil,
			},
		},
		{
			description: "succeeds when public key list len is greater than 1",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected: Expected{
				pubKey: []models.PublicKey{
					{
						Data:        []byte("test"),
						CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Fingerprint: "fingerprint",
						TenantID:    "00000000-0000-4000-0000-000000000000",
						PublicKeyFields: models.PublicKeyFields{
							Name: "public_key",
							Filter: models.PublicKeyFilter{
								Hostname: ".*",
								Tags:     []string{"tag-1"},
							},
						},
					},
				},
				len: 1,
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			pubKey, count, err := mongostore.PublicKeyList(context.TODO(), query.Paginator{Page: -1, PerPage: -1})
			assert.Equal(t, tc.expected, Expected{pubKey: pubKey, len: count, err: err})
		})
	}
}

func TestPublicKeyCreate(t *testing.T) {
	cases := []struct {
		description string
		key         *models.PublicKey
		fixtures    []string
		expected    error
	}{
		{
			description: "succeeds when data is valid",
			key: &models.PublicKey{
				Data:            []byte("test"),
				Fingerprint:     "fingerprint",
				TenantID:        "00000000-0000-4000-0000-000000000000",
				PublicKeyFields: models.PublicKeyFields{Name: "public_key", Filter: models.PublicKeyFilter{Hostname: ".*"}},
			},
			fixtures: []string{},
			expected: nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.PublicKeyCreate(context.TODO(), tc.key)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestPublicKeyUpdate(t *testing.T) {
	type Expected struct {
		pubKey *models.PublicKey
		err    error
	}

	cases := []struct {
		description string
		fingerprint string
		tenant      string
		key         *models.PublicKeyUpdate
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			key: &models.PublicKeyUpdate{
				PublicKeyFields: models.PublicKeyFields{
					Name:   "edited_name",
					Filter: models.PublicKeyFilter{Hostname: ".*"},
				},
			},
			fixtures: []string{fixtures.FixturePublicKeys},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			key: &models.PublicKeyUpdate{
				PublicKeyFields: models.PublicKeyFields{
					Name:   "edited_name",
					Filter: models.PublicKeyFilter{Hostname: ".*"},
				},
			},
			fixtures: []string{fixtures.FixturePublicKeys},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			key: &models.PublicKeyUpdate{
				PublicKeyFields: models.PublicKeyFields{
					Name: "edited_key",
					Filter: models.PublicKeyFilter{
						Hostname: ".*",
						Tags:     []string{"edited-tag"},
					},
				},
			},
			fixtures: []string{fixtures.FixturePublicKeys},
			expected: Expected{
				pubKey: &models.PublicKey{
					Data:        []byte("test"),
					CreatedAt:   time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Fingerprint: "fingerprint",
					TenantID:    "00000000-0000-4000-0000-000000000000",
					PublicKeyFields: models.PublicKeyFields{
						Name: "edited_key",
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
							Tags:     []string{"edited-tag"},
						},
					},
				},
				err: nil,
			},
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			pubKey, err := mongostore.PublicKeyUpdate(context.TODO(), tc.fingerprint, tc.tenant, tc.key)
			assert.Equal(t, tc.expected, Expected{pubKey: pubKey, err: err})
		})
	}
}

func TestPublicKeyDelete(t *testing.T) {
	cases := []struct {
		description string
		fingerprint string
		tenant      string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when public key is not found due to fingerprint",
			fingerprint: "nonexistent",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtures.FixturePublicKeys},
			expected:    nil,
		},
	}

	db := dbtest.DBServer{}
	defer db.Stop()

	mongostore := NewStore(db.Client().Database("test"), cache.NewNullCache())
	fixtures.Init(db.Host, "test")

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			assert.NoError(t, fixtures.Apply(tc.fixtures...))
			defer fixtures.Teardown() // nolint: errcheck

			err := mongostore.PublicKeyDelete(context.TODO(), tc.fingerprint, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}
}
