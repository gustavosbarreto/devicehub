package mongo_test

import (
	"context"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
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
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			fixtures:    []string{fixturePublicKeys},
			expected: Expected{
				pubKey: nil,
				err:    store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixturePublicKeys},
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
							Taggable: models.Taggable{TagsID: []string{"6791d3ae04ba86e6d7a0514d"}, Tags: nil},
						},
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			pubKey, err := s.PublicKeyGet(ctx, tc.fingerprint, tc.tenant)
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
			fixtures:    []string{fixturePublicKeys},
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
								Taggable: models.Taggable{TagsID: []string{"6791d3ae04ba86e6d7a0514d"}, Tags: nil},
							},
						},
					},
				},
				len: 1,
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			pubKey, count, err := s.PublicKeyList(ctx, query.Paginator{Page: -1, PerPage: -1})
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

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.PublicKeyCreate(ctx, tc.key)
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
			fixtures: []string{fixturePublicKeys},
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
			fixtures: []string{fixturePublicKeys},
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
						Taggable: models.Taggable{TagsID: []string{"6791d3ae04ba86e6d7a0514d"}, Tags: nil},
					},
				},
			},
			fixtures: []string{fixturePublicKeys},
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
							Taggable: models.Taggable{TagsID: []string{"6791d3ae04ba86e6d7a0514d"}, Tags: nil},
						},
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			pubKey, err := s.PublicKeyUpdate(ctx, tc.fingerprint, tc.tenant, tc.key)
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
			fixtures:    []string{fixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when public key is not found due to tenant",
			fingerprint: "fingerprint",
			tenant:      "nonexistent",
			fixtures:    []string{fixturePublicKeys},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when public key is found",
			fingerprint: "fingerprint",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixturePublicKeys},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			assert.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				assert.NoError(t, srv.Reset())
			})

			err := s.PublicKeyDelete(ctx, tc.fingerprint, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}
}
