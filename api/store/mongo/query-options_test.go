package mongo_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
)

func TestCountAcceptedDevices(t *testing.T) {
	type Expected struct {
		count int
		err   error
	}

	cases := []struct {
		description string
		tenant      string
		ctx         func() context.Context
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when context does not have db in values",
			tenant:      "00000000-0000-4000-0000-000000000000",
			ctx: func() context.Context {
				return context.Background()
			},
			fixtures: []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{count: 0, err: errors.New("db not found in context")},
		},
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			ctx: func() context.Context {
				return context.WithValue(context.Background(), "db", db) //nolint:revive
			},
			fixtures: []string{fixtureNamespaces, fixtureDevices},
			expected: Expected{count: 3, err: nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := tc.ctx()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			ns := &models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000"}
			err := s.Options().CountAcceptedDevices()(ctx, ns)
			require.Equal(tt, tc.expected, Expected{ns.DevicesCount, err})
		})
	}
}

func TestEnrichMembersData(t *testing.T) {
	type Expected struct {
		emails []string
		err    error
	}

	cases := []struct {
		description string
		tenant      string
		ctx         func() context.Context
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when context does not have db in values",
			tenant:      "00000000-0000-4000-0000-000000000000",
			ctx: func() context.Context {
				return context.Background()
			},
			fixtures: []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				emails: []string{},
				err:    errors.New("db not found in context"),
			},
		},
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			ctx: func() context.Context {
				return context.WithValue(context.Background(), "db", db) //nolint:revive
			},
			fixtures: []string{fixtureNamespaces, fixtureUsers},
			expected: Expected{
				emails: []string{"john.doe@test.com", "maria.garcia@test.com"},
				err:    nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := tc.ctx()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			ns := &models.Namespace{
				TenantID: "00000000-0000-4000-0000-000000000000",
				Members:  []models.Member{{ID: "507f1f77bcf86cd799439011"}, {ID: "6509e169ae6144b2f56bf288"}},
			}

			err := s.Options().EnrichMembersData()(ctx, ns)
			require.Equal(tt, tc.expected.err, err)

			if err == nil {
				for _, m := range ns.Members {
					require.Equal(tt, true, slices.Contains(tc.expected.emails, m.Email))
				}
			}
		})
	}
}

func TestQueryOptions_DeviceWithTagDetails(t *testing.T) {
	type Expected struct {
		tags []models.Tag
		err  error
	}

	cases := []struct {
		description string
		device      *models.Device
		ctx         func() context.Context
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when context does not have db in values",
			device: &models.Device{
				Taggable: models.Taggable{TagsID: []string{"6791d3ae04ba86e6d7a0514d"}, Tags: nil},
			},
			ctx: func() context.Context {
				return context.Background()
			},
			fixtures: []string{fixtureDevices, fixtureTags},
			expected: Expected{
				tags: nil,
				err:  errors.New("store not found in context"),
			},
		},
		{
			description: "succeeds when device has no tags",
			device: &models.Device{
				UID:      "4300430e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809e",
				Taggable: models.Taggable{TagsID: []string{}, Tags: nil},
			},
			ctx: func() context.Context {
				return context.WithValue(context.Background(), "store", s) //nolint:revive
			},
			fixtures: []string{fixtureDevices, fixtureTags},
			expected: Expected{
				tags: []models.Tag{},
				err:  nil,
			},
		},
		{
			description: "succeeds when device has tags",
			device: &models.Device{
				UID:      "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
				Taggable: models.Taggable{TagsID: []string{"6791d3ae04ba86e6d7a0514d"}, Tags: nil},
			},
			ctx: func() context.Context {
				return context.WithValue(context.Background(), "store", s) //nolint:revive
			},
			fixtures: []string{fixtureDevices, fixtureTags},
			expected: Expected{
				tags: []models.Tag{
					{
						ID:        "6791d3ae04ba86e6d7a0514d",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "production",
						TenantID:  "00000000-0000-4000-0000-000000000000",
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := tc.ctx()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			err := s.Options().DeviceWithTagDetails()(ctx, tc.device)
			require.Equal(tt, tc.expected, Expected{tc.device.Tags, err})
		})
	}
}

func TestQueryOptions_PublicKeyWithTagDetails(t *testing.T) {
	type Expected struct {
		tags []models.Tag
		err  error
	}

	cases := []struct {
		description string
		publicKey   *models.PublicKey
		ctx         func() context.Context
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when context does not have db in values",
			publicKey: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Taggable: models.Taggable{TagsID: []string{"6791d3ae04ba86e6d7a0514d"}, Tags: nil},
					},
				},
			},
			ctx: func() context.Context {
				return context.Background()
			},
			fixtures: []string{fixturePublicKeys, fixtureTags},
			expected: Expected{
				tags: nil,
				err:  errors.New("store not found in context"),
			},
		},
		{
			description: "succeeds when public key has no tags",
			publicKey: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Taggable: models.Taggable{TagsID: []string{}, Tags: nil},
					},
				},
			},
			ctx: func() context.Context {
				return context.WithValue(context.Background(), "store", s) //nolint:revive
			},
			fixtures: []string{fixturePublicKeys, fixtureTags},
			expected: Expected{
				tags: []models.Tag{},
				err:  nil,
			},
		},
		{
			description: "succeeds when public key has tags",
			publicKey: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Taggable: models.Taggable{TagsID: []string{"6791d3ae04ba86e6d7a0514d"}, Tags: nil},
					},
				},
			},
			ctx: func() context.Context {
				return context.WithValue(context.Background(), "store", s) //nolint:revive
			},
			fixtures: []string{fixturePublicKeys, fixtureTags},
			expected: Expected{
				tags: []models.Tag{
					{
						ID:        "6791d3ae04ba86e6d7a0514d",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "production",
						TenantID:  "00000000-0000-4000-0000-000000000000",
					},
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := tc.ctx()

			require.NoError(tt, srv.Apply(tc.fixtures...))
			tt.Cleanup(func() {
				require.NoError(tt, srv.Reset())
			})

			err := s.Options().PublicKeyWithTagDetails()(ctx, tc.publicKey)
			require.Equal(tt, tc.expected, Expected{tc.publicKey.Filter.Tags, err})
		})
	}
}
