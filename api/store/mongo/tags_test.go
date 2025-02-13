package mongo_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmocks "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestStore_TagCreate(t *testing.T) {
	now := time.Now()

	clockMock := new(clockmocks.Clock)
	clockMock.On("Now").Return(now)
	clock.DefaultBackend = clockMock

	cases := []struct {
		description string
		tag         *models.Tag
		expected    error
	}{
		{
			description: "succeeds when tag data is valid",
			tag:         &models.Tag{Name: "staging", TenantID: "00000000-0000-4000-0000-000000000000"},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.Background()

			insertedID, err := s.TagCreate(ctx, tc.tag)
			require.Equal(tt, tc.expected, err)

			if err == nil {
				objID, _ := primitive.ObjectIDFromHex(insertedID)

				tag := make(map[string]interface{})
				require.NoError(tt, db.Collection("tags").FindOne(ctx, bson.M{"_id": objID}).Decode(tag))

				require.Equal(
					tt,
					map[string]interface{}{
						"_id":        objID,
						"created_at": primitive.NewDateTimeFromTime(now),
						"updated_at": primitive.NewDateTimeFromTime(now),
						"name":       "staging",
						"tenant_id":  "00000000-0000-4000-0000-000000000000",
					},
					tag,
				)
			}
		})
	}
}

func TestStore_TagConflicts(t *testing.T) {
	type Expected struct {
		conflicts []string
		has       bool
		err       error
	}

	cases := []struct {
		description string
		tenantID    string
		target      *models.TagConflicts
		fixtures    []string
		expected    Expected
	}{
		{
			description: "no conflicts when target is empty",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			target:      &models.TagConflicts{},
			fixtures:    []string{fixtureTags},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflicts with non existing name",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			target:      &models.TagConflicts{Name: "nonexistent"},
			fixtures:    []string{fixtureTags},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "no conflicts when namespace is different",
			tenantID:    "00000000-0000-4001-0000-000000000000",
			target:      &models.TagConflicts{Name: "production"},
			fixtures:    []string{fixtureTags},
			expected:    Expected{[]string{}, false, nil},
		},
		{
			description: "conflict detected with existing name",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			target:      &models.TagConflicts{Name: "production"},
			fixtures:    []string{fixtureTags},
			expected:    Expected{[]string{"name"}, true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				require.NoError(t, srv.Reset())
			})

			conflicts, has, err := s.TagConflicts(ctx, tc.tenantID, tc.target)
			require.Equal(t, tc.expected, Expected{conflicts, has, err})
		})
	}
}

func TestStore_TagList(t *testing.T) {
	type Expected struct {
		tags  []models.Tag
		count int
		err   error
	}

	cases := []struct {
		description string
		tenantID    string
		paginator   query.Paginator
		filters     query.Filters
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when tenantID is empty",
			tenantID:    "",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			filters:     query.Filters{},
			fixtures:    []string{fixtureTags},
			expected: Expected{
				tags: []models.Tag{
					{
						ID:        "6791d3c2a62aafaefe821ab3",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "owners",
						TenantID:  "00000000-0000-4001-0000-000000000000",
					},
					{
						ID:        "6791d3ae04ba86e6d7a0514d",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "production",
						TenantID:  "00000000-0000-4000-0000-000000000000",
					},
					{
						ID:        "6791d3be5a201d874c4c2885",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "development",
						TenantID:  "00000000-0000-4000-0000-000000000000",
					},
				},
				count: 3,
				err:   nil,
			},
		},
		{
			description: "succeeds when tenantID is not empty",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			paginator:   query.Paginator{Page: -1, PerPage: -1},
			filters:     query.Filters{},
			fixtures:    []string{fixtureTags},
			expected: Expected{
				tags: []models.Tag{
					{
						ID:        "6791d3ae04ba86e6d7a0514d",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "production",
						TenantID:  "00000000-0000-4000-0000-000000000000",
					},
					{
						ID:        "6791d3be5a201d874c4c2885",
						CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
						Name:      "development",
						TenantID:  "00000000-0000-4000-0000-000000000000",
					},
				},
				count: 2,
				err:   nil,
			},
		},
	}

	// Due to the non-deterministic order of applying fixtures when dealing with multiple datasets,
	// we ensure that both the expected and result arrays are correctly sorted.
	sort := func(ns []models.Tag) {
		sort.Slice(ns, func(i, j int) bool {
			return ns[i].Name < ns[j].Name
		})
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				require.NoError(t, srv.Reset())
			})

			tags, count, err := s.TagList(ctx, tc.tenantID, tc.paginator, tc.filters, query.Sorter{})

			sort(tc.expected.tags)
			sort(tags)

			require.Equal(t, tc.expected, Expected{tags: tags, count: count, err: err})
		})
	}
}

func TestStore_TagGetByID(t *testing.T) {
	type Expected struct {
		tag *models.Tag
		err error
	}

	cases := []struct {
		description string
		id          string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tag is not found",
			id:          "000000000000000000000000",
			fixtures:    []string{fixtureTags},
			expected: Expected{
				tag: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when tag is found",
			id:          "6791d3ae04ba86e6d7a0514d",
			fixtures:    []string{fixtureTags},
			expected: Expected{
				tag: &models.Tag{
					ID:        "6791d3ae04ba86e6d7a0514d",
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "production",
					TenantID:  "00000000-0000-4000-0000-000000000000",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				require.NoError(t, srv.Reset())
			})

			tag, err := s.TagGetByID(ctx, tc.id)
			require.Equal(t, tc.expected, Expected{tag: tag, err: err})
		})
	}
}

func TestStore_TagGetByName(t *testing.T) {
	type Expected struct {
		tag *models.Tag
		err error
	}

	cases := []struct {
		description string
		tenantID    string
		name        string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tag is not found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "nonexistent",
			fixtures:    []string{fixtureTags},
			expected: Expected{
				tag: nil,
				err: store.ErrNoDocuments,
			},
		},
		{
			description: "succeeds when tag is found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "production",
			fixtures:    []string{fixtureTags},
			expected: Expected{
				tag: &models.Tag{
					ID:        "6791d3ae04ba86e6d7a0514d",
					CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
					Name:      "production",
					TenantID:  "00000000-0000-4000-0000-000000000000",
				},
				err: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				require.NoError(t, srv.Reset())
			})

			tag, err := s.TagGetByName(ctx, tc.tenantID, tc.name)
			require.Equal(t, tc.expected, Expected{tag: tag, err: err})
		})
	}
}

func TestStore_TagUpdate(t *testing.T) {
	cases := []struct {
		description   string
		tenantID      string
		name          string
		changes       *models.TagChanges
		fixtures      []string
		expected      error
		assertChanges func(context.Context) error
	}{
		{
			description: "fails when tag is not found",
			tenantID:    "nonexistent",
			name:        "nonexistent",
			changes: &models.TagChanges{
				Name: "edited-tag",
			},
			fixtures:      []string{fixtureTags},
			expected:      store.ErrNoDocuments,
			assertChanges: nil,
		},
		{
			description: "succeeds when tag is found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "production",
			changes: &models.TagChanges{
				Name: "edited-tag",
			},
			fixtures: []string{fixtureTags},
			expected: nil,
			assertChanges: func(ctx context.Context) error {
				tag := new(models.Tag)
				err := db.Collection("tags").FindOne(ctx, bson.M{"tenant_id": "00000000-0000-4000-0000-000000000000", "name": "edited-tag"}).Decode(tag)

				return err
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				require.NoError(t, srv.Reset())
			})

			err := s.TagUpdate(ctx, tc.tenantID, tc.name, tc.changes)
			require.Equal(t, tc.expected, err)

			if err == nil {
				require.NoError(t, tc.assertChanges(ctx))
			}
		})
	}
}

func TestStore_TagPushToTarget(t *testing.T) {
	cases := []struct {
		description string
		tenantID    string
		name        string
		target      models.TagTarget
		targetID    string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tag does not exist",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "nonexistent",
			target:      models.TagTargetDevice,
			targetID:    "656f605bafb652df9927adef",
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when device does not exist",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "development",
			target:      models.TagTargetDevice,
			targetID:    "nonexistent",
			fixtures:    []string{fixtureTags},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds to push a tag to device",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "development",
			target:      models.TagTargetDevice,
			targetID:    "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
			fixtures:    []string{fixtureTags, fixtureDevices},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				require.NoError(t, srv.Reset())
			})

			err := s.TagPushToTarget(ctx, tc.tenantID, tc.name, tc.target, tc.targetID)
			require.Equal(t, tc.expected, err)

			if err == nil {
				var device struct {
					Tags []string `bson:"tags"`
				}

				err := db.Collection("devices").FindOne(ctx, bson.M{"uid": tc.targetID}).Decode(&device)
				require.NoError(t, err)

				tag, err := s.TagGetByName(ctx, tc.tenantID, tc.name)
				require.NoError(t, err)
				require.Contains(t, device.Tags, tag.ID)
			}
		})
	}
}

func TestTagPullFromTarget(t *testing.T) {
	cases := []struct {
		description string
		tenantID    string
		name        string
		target      models.TagTarget
		targetID    string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tag does not exist",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "nonexistent",
			target:      models.TagTargetDevice,
			targetID:    "656f605bafb652df9927adef",
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when device does not exist",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "production",
			target:      models.TagTargetDevice,
			targetID:    "nonexistent",
			fixtures:    []string{fixtureTags},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds to pull a tag from device",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "production",
			target:      models.TagTargetDevice,
			targetID:    "5300530e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809f",
			fixtures:    []string{fixtureTags, fixtureDevices},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				require.NoError(t, srv.Reset())
			})

			err := s.TagPullFromTarget(ctx, tc.tenantID, tc.name, tc.target, tc.targetID)
			require.Equal(t, tc.expected, err)

			if err == nil {
				var device struct {
					Tags []string `bson:"tags"`
				}

				err := db.Collection("devices").FindOne(ctx, bson.M{"uid": tc.targetID}).Decode(&device)
				require.NoError(t, err)

				tag, err := s.TagGetByName(ctx, tc.tenantID, tc.name)
				require.NoError(t, err)
				require.NotContains(t, device.Tags, tag.ID)
			}
		})
	}
}

func TestStore_TagDelete(t *testing.T) {
	cases := []struct {
		description string
		tenantID    string
		name        string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when tag is not found due to tenant ID",
			tenantID:    "nonexistent",
			name:        "production",
			fixtures:    []string{fixtureTags},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when tag is not found due to name",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "nonexistent",
			fixtures:    []string{fixtureTags},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "succeeds when tag is found",
			tenantID:    "00000000-0000-4000-0000-000000000000",
			name:        "production",
			fixtures:    []string{fixtureTags},
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			ctx := context.Background()

			require.NoError(t, srv.Apply(tc.fixtures...))
			t.Cleanup(func() {
				require.NoError(t, srv.Reset())
			})

			err := s.TagDelete(ctx, tc.tenantID, tc.name)
			require.Equal(t, tc.expected, err)

			if err == nil {
				count, err := db.Collection("tags").CountDocuments(ctx, bson.M{"tenant_id": tc.tenantID, "name": tc.name})
				require.NoError(t, err)
				require.Equal(t, int64(0), count)
			}
		})
	}
}
