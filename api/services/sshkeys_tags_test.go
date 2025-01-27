package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestAddPublicKeyTag(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		tenant        string
		fingerprint   string
		tag           string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fail when namespace was not found",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("tenant", errors.New("error", "", 0)),
		},
		{
			description: "fail when public key was not found",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "tenant"}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrPublicKeyNotFound("fingerprint", errors.New("error", "", 0)),
		},
		{
			description: "fail when the tag limit on public key has reached",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}

				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2", "tag3"},
						},
					},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
			},
			expected: NewErrTagLimit(DeviceMaxTags, nil),
		},
		{
			description: "fail when the tag does not exist in a device",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tagsNames := []string{
					"tag-1", "tag-2",
				}
				tags := []models.Tags{
					{
						Name: "tag-1",
					},
					{
						Name: "tag-2",
					},
				}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: tagsNames,
						},
					},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(tags, int64(len(tags)), nil).Once()
			},
			expected: NewErrTagNotFound("tag", nil),
		},
		{
			description: "fail when cannot add tag to public key",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				// tagsNames := []string{"tag", "tag3", "tag6"}
				tags := []models.Tags{
					{
						Name:   "tag",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag3",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag6",
						Color:  "",
						Tenant: "tenant",
					},
				}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(tags, int64(len(tags)), nil).Once()
				mock.On("PublicKeyPushTag", ctx, "tenant", "fingerprint", "tag").Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to add a to public key",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tags := []models.Tags{
					{
						Name:   "tag",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag3",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag6",
						Color:  "",
						Tenant: "tenant",
					},
				}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(tags, int64(len(tags)), nil).Once()
				mock.On("PublicKeyPushTag", ctx, "tenant", "fingerprint", "tag").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := services.AddPublicKeyTag(ctx, tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestRemovePublicKeyTag(t *testing.T) {
	mock := &mocks.Store{}

	ctx := context.TODO()

	cases := []struct {
		description   string
		tenant        string
		fingerprint   string
		tag           string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fail when namespace was not found",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("tenant", nil),
		},
		{
			description: "fail when public key was not found",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "tenant"}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrPublicKeyNotFound("fingerprint", errors.New("error", "", 0)),
		},
		{
			description: "fail when the tag does not exist in public key",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tags := []string{"tag1", "tag2"}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: tags,
						},
					},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
			},
			expected: NewErrTagNotFound("tag", nil),
		},
		{
			description: "fail when remove the tag from public key",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tags := []string{"tag", "tag1", "tag2"}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: tags,
						},
					},
				}
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("PublicKeyPullTag", ctx, "tenant", "fingerprint", "tag").Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success when remove a from public key",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tag:         "tag",
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tags := []string{"tag", "tag1", "tag2"}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: tags,
						},
					},
				}
				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("PublicKeyPullTag", ctx, "tenant", "fingerprint", "tag").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := services.RemovePublicKeyTag(ctx, tc.tenant, tc.fingerprint, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdatePublicKeyTags(t *testing.T) {
	mock := &mocks.Store{}

	ctx := context.TODO()

	cases := []struct {
		description   string
		tenant        string
		fingerprint   string
		tags          []string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fail when namespace was not found",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tags:        []string{"tag1", "tag2", "tag3"},
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "tenant").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("tenant", nil),
		},
		{
			description: "fail when public key was not found",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tags:        []string{"tag1", "tag2", "tag3"},
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "tenant"}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrPublicKeyNotFound("fingerprint", errors.New("error", "", 0)),
		},
		{
			description: "fail when tags are great the tag limit",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tags:        []string{"tag4", "tag5", "tag7", "tag5"},
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
			},
			expected: NewErrTagLimit(DeviceMaxTags, nil),
		},
		{
			description: "fail when a tag does not exist in a device",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tags:        []string{"tag2", "tag4", "tag5"},
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tags := []models.Tags{
					{
						Name:   "tag4",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag5",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag7",
						Color:  "",
						Tenant: "tenant",
					},
				}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(tags, int64(len(tags)), nil).Once()
			},
			expected: NewErrTagNotFound("tag2", nil),
		},
		{
			description: "fail when update tags in public key fails",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tags:        []string{"tag1", "tag2", "tag3"},
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tags := []models.Tags{
					{
						Name:   "tag1",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag2",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag3",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag4",
						Color:  "",
						Tenant: "tenant",
					},
				}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(tags, int64(len(tags)), nil).Once()
				mock.On("PublicKeySetTags", ctx, "tenant", "fingerprint", []string{"tag1", "tag2", "tag3"}).Return(int64(0), int64(0), errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success update tags in public key",
			tenant:      "tenant",
			fingerprint: "fingerprint",
			tags:        []string{"tag1", "tag2", "tag3"},
			requiredMocks: func() {
				namespace := &models.Namespace{
					TenantID: "tenant",
				}
				tags := []models.Tags{
					{
						Name:   "tag1",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag2",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag3",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag4",
						Color:  "",
						Tenant: "tenant",
					},
				}
				key := &models.PublicKey{
					TenantID:    "tenant",
					Fingerprint: "fingerprint",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant").Return(key, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(tags, int64(len(tags)), nil).Once()
				mock.On("PublicKeySetTags", ctx, "tenant", "fingerprint", []string{"tag1", "tag2", "tag3"}).Return(int64(1), int64(1), nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			services := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := services.UpdatePublicKeyTags(ctx, tc.tenant, tc.fingerprint, tc.tags)
			assert.Equal(t, tc.expected, err)
		})
	}
}
