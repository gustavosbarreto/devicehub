package mongo_test

import (
	"context"
	// "errors"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	// mongo "github.com/shellhub-io/shellhub/api/store/mongo"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestDevicePushTag(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when device doesn't exist",
			uid:         models.UID("nonexistent"),
			tag:         "tag4",
			fixtures:    []string{fixtureTags, fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "successfully creates single tag for an existing device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tag:         "tag4",
			fixtures:    []string{fixtureTags, fixtureDevices},
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

			err := s.DevicePushTag(ctx, tc.uid, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDevicePullTag(t *testing.T) {
	cases := []struct {
		description string
		uid         models.UID
		tag         string
		fixtures    []string
		expected    error
	}{
		{
			description: "fails when device doesn't exist",
			uid:         models.UID("nonexistent"),
			tag:         "tag-1",
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "fails when device's tag doesn't exist",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tag:         "nonexistent",
			fixtures:    []string{fixtureDevices},
			expected:    store.ErrNoDocuments,
		},
		{
			description: "successfully remove a single tag for an existing device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tag:         "tag-1",
			fixtures:    []string{fixtureDevices},
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

			err := s.DevicePullTag(ctx, tc.uid, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}
}

func TestDeviceSetTags(t *testing.T) {
	type Expected struct {
		matchedCount int64
		updatedCount int64
		err          error
	}
	cases := []struct {
		description string
		uid         models.UID
		tags        []string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when device doesn't exist",
			uid:         models.UID("nonexistent"),
			tags:        []string{"tag-1"},
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				matchedCount: 0,
				updatedCount: 0,
				err:          store.ErrNoDocuments,
			},
		},
		{
			description: "successfully when tags are equal than current device's tags",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tags:        []string{"tag-1"},
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				matchedCount: 1,
				updatedCount: 0,
				err:          nil,
			},
		},
		{
			description: "successfully update tags for an existing device",
			uid:         models.UID("2300230e3ca2f637636b4d025d2235269014865db5204b6d115386cbee89809c"),
			tags:        []string{"new-tag"},
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				matchedCount: 1,
				updatedCount: 1,
				err:          nil,
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

			matchedCount, updatedCount, err := s.DeviceSetTags(ctx, tc.uid, tc.tags)
			assert.Equal(t, tc.expected, Expected{matchedCount, updatedCount, err})
		})
	}
}

func TestDeviceBulkRenameTag(t *testing.T) {
	type Expected struct {
		count int64
		err   error
	}

	cases := []struct {
		description string
		tenant      string
		oldTag      string
		newTag      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant doesn't exist",
			tenant:      "nonexistent",
			oldTag:      "tag-1",
			newTag:      "newtag",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when device's tag doesn't exist",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "nonexistent",
			newTag:      "newtag",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "successfully rename tag for an existing device",
			tenant:      "00000000-0000-4000-0000-000000000000",
			oldTag:      "tag-1",
			newTag:      "newtag",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				count: 2,
				err:   nil,
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

			count, err := s.DeviceBulkRenameTag(ctx, tc.tenant, tc.oldTag, tc.newTag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}

func TestDeviceBulkDeleteTag(t *testing.T) {
	type Expected struct {
		count int64
		err   error
	}

	cases := []struct {
		description string
		tenant      string
		tag         string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "fails when tenant doesn't exist",
			tenant:      "nonexistent",
			tag:         "tag-1",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "fails when device's tag doesn't exist",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "nonexistent",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				count: 0,
				err:   nil,
			},
		},
		{
			description: "successfully delete single tag for an existing device",
			tenant:      "00000000-0000-4000-0000-000000000000",
			tag:         "tag-1",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				count: 2,
				err:   nil,
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

			count, err := s.DeviceBulkDeleteTag(ctx, tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, Expected{count, err})
		})
	}
}

func TestDeviceGetTags(t *testing.T) {
	type Expected struct {
		tags []string
		len  int
		err  error
	}

	cases := []struct {
		description string
		tenant      string
		fixtures    []string
		expected    Expected
	}{
		{
			description: "succeeds when tags list is greater than 1",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fixtures:    []string{fixtureDevices},
			expected: Expected{
				tags: []string{"tag-1"},
				len:  1,
				err:  nil,
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

			tags, count, err := s.DeviceGetTags(ctx, tc.tenant)
			assert.Equal(t, tc.expected, Expected{tags: tags, len: count, err: err})
		})
	}
}
