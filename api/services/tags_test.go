package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	mocksGeoIp "github.com/shellhub-io/shellhub/pkg/geoip/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/shellhub-io/shellhub/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestGetTags(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	type Expected struct {
		Tags  []models.Tags
		Count int
		Error error
	}

	cases := []struct {
		name          string
		uid           models.UID
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			name:     "fail when namespace is not found",
			tenantID: "not_found_tenant",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "not_found_tenant").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{
				Tags:  nil,
				Count: 0,
				Error: NewErrNamespaceNotFound("not_found_tenant", errors.New("error", "", 0)),
			},
		},
		{
			name:     "fail when store function to get tags fails",
			tenantID: "tenant",
			requiredMocks: func() {
				namespace := &models.Namespace{Name: "namespace", TenantID: "tenant"}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(nil, int64(0), errors.New("error", "", 0)).Once()
			},
			expected: Expected{
				Tags:  nil,
				Count: 0,
				Error: errors.New("error", "", 0),
			},
		},
		{
			name:     "success to get tags",
			tenantID: "tenant",
			requiredMocks: func() {
				device := &models.Device{
					UID:       "uid",
					Namespace: "namespace",
					TenantID:  "tenant",
					Tags:      []string{"device1", "device2"},
				}

				namespace := &models.Namespace{Name: "namespace", TenantID: "tenant"}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return([]models.Tags{
					{
						Name:   "device1",
						Tenant: "tenant",
						Color:  "#efdaef",
					},
					{
						Name:   "device2",
						Tenant: "tenant",
						Color:  "#efdaef",
					},
				}, int64(len(device.Tags)), nil).Once()
			},
			expected: Expected{
				Tags: []models.Tags{
					{
						Name:   "device1",
						Tenant: "tenant",
						Color:  "#efdaef",
					},
					{
						Name:   "device2",
						Tenant: "tenant",
						Color:  "#efdaef",
					},
				},
				Count: len([]string{"device1", "device2"}),
				Error: nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			locator := &mocksGeoIp.Locator{}
			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, WithLocator(locator))

			tags, count, err := service.GetTags(ctx, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{tags, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestRenameTag(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		name          string
		tenantID      string
		currentTag    string
		newTag        string
		requiredMocks func()
		expected      error
	}{
		{
			name:          "fail when tag is invalid",
			tenantID:      "tenant",
			currentTag:    "currentTag",
			newTag:        "invalid_tag",
			requiredMocks: func() {},
			expected:      NewErrTagInvalid("invalid_tag", validator.ErrStructureInvalid),
		},
		{
			name:       "fail when device has no tags",
			tenantID:   "namespaceTenantIDNoTag",
			currentTag: "device3",
			newTag:     "device1",
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "namespaceTenantIDNoTag").Return(nil, int64(0), errors.New("error", "", 0))
			},
			expected: NewErrTagEmpty("namespaceTenantIDNoTag", errors.New("error", "", 0)),
		},
		{
			name:       "fail when device don't have the tag",
			tenantID:   "namespaceTenantID",
			currentTag: "device2",
			newTag:     "device1",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespaceName",
					Owner:    "owner",
					TenantID: "namespaceTenantID",
				}

				deviceWithTags := &models.Device{
					UID:      "deviceWithTagsUID",
					Name:     "deviceWithTagsName",
					TenantID: "deviceWithTagsTenantID",
					Tags:     []string{"device3", "device4", "device5"},
				}

				mock.On("TagsGet", ctx, namespace.TenantID).Return([]models.Tags{
					{
						Name:   "device3",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device4",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device5",
						Color:  "",
						Tenant: "tenant",
					},
				}, int64(len(deviceWithTags.Tags)), nil).Once()
			},
			expected: NewErrTagNotFound("device2", nil),
		},
		{
			name:       "fail when device already have the tag",
			tenantID:   "namespaceTenantID",
			currentTag: "device3",
			newTag:     "device5",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespaceName",
					Owner:    "owner",
					TenantID: "namespaceTenantID",
				}

				deviceWithTags := &models.Device{
					UID:      "deviceWithTagsUID",
					Name:     "deviceWithTagsName",
					TenantID: "deviceWithTagsTenantID",
					Tags:     []string{"device3", "device4", "device5"},
				}

				mock.On("TagsGet", ctx, namespace.TenantID).Return([]models.Tags{
					{
						Name:   "device3",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device4",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device5",
						Color:  "",
						Tenant: "tenant",
					},
				}, int64(len(deviceWithTags.Tags)), nil).Once()
			},
			expected: NewErrTagDuplicated("device5", nil),
		},
		{
			name:       "fail when the store function to rename the tag fails",
			tenantID:   "namespaceTenantID",
			currentTag: "device3",
			newTag:     "device1",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespaceName",
					Owner:    "owner",
					TenantID: "namespaceTenantID",
				}

				deviceWithTags := &models.Device{
					UID:      "deviceWithTagsUID",
					Name:     "deviceWithTagsName",
					TenantID: "deviceWithTagsTenantID",
					Tags:     []string{"device3", "device4", "device5"},
				}

				mock.On("TagsGet", ctx, namespace.TenantID).Return([]models.Tags{
					{
						Name:   "device3",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device4",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device5",
						Color:  "",
						Tenant: "tenant",
					},
				}, int64(len(deviceWithTags.Tags)), nil).Once()
				mock.On("TagsRename", ctx, namespace.TenantID, "device3", "device1").Return(int64(0), errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			name:       "success to rename the tag",
			tenantID:   "namespaceTenantID",
			currentTag: "device3",
			newTag:     "device1",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespaceName",
					Owner:    "owner",
					TenantID: "namespaceTenantID",
				}

				deviceWithTags := &models.Device{
					UID:      "deviceWithTagsUID",
					Name:     "deviceWithTagsName",
					TenantID: "deviceWithTagsTenantID",
					Tags:     []string{"device3", "device4", "device5"},
				}

				mock.On("TagsGet", ctx, namespace.TenantID).Return([]models.Tags{
					{
						Name:   "device3",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device4",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device5",
						Color:  "",
						Tenant: "tenant",
					},
				}, int64(len(deviceWithTags.Tags)), nil).Once()
				mock.On("TagsRename", ctx, namespace.TenantID, "device3", "device1").Return(int64(1), nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			locator := &mocksGeoIp.Locator{}
			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, WithLocator(locator))

			err := service.RenameTag(ctx, tc.tenantID, tc.currentTag, tc.newTag)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestDeleteTag(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		name          string
		tag           string
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			name:   "fail when tag is invalid",
			tag:    "invalid_tag",
			tenant: "tenant",
			requiredMocks: func() {
			},
			expected: NewErrTagInvalid("invalid_tag", validator.ErrStructureInvalid),
		},
		{
			name:   "fail when could not find the namespace",
			tag:    "device1",
			tenant: "not_found_tenant",
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, "not_found_tenant").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: NewErrNamespaceNotFound("not_found_tenant", errors.New("error", "", 0)),
		},
		{
			name:   "fail when tags are empty",
			tag:    "device1",
			tenant: "tenant",
			requiredMocks: func() {
				namespace := &models.Namespace{Name: "namespace", TenantID: "tenant"}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return(nil, int64(0), errors.New("error", "", 0)).Once()
			},
			expected: NewErrTagEmpty("tenant", errors.New("error", "", 0)),
		},
		{
			name:   "fail when tag does not exist",
			tag:    "device3",
			tenant: "tenant",
			requiredMocks: func() {
				namespace := &models.Namespace{Name: "namespace", TenantID: "tenant"}

				device := &models.Device{
					UID:       "uid",
					Namespace: "namespace",
					TenantID:  "tenant",
					Tags:      []string{"device1", "device2"},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return([]models.Tags{
					{
						Name:   "device1",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device2",
						Color:  "",
						Tenant: "tenant",
					},
				}, int64(len(device.Tags)), nil).Once()
			},
			expected: NewErrTagNotFound("device3", nil),
		},
		{
			name:   "fail when the store function to delete the tag fails",
			tag:    "device1",
			tenant: "tenant",
			requiredMocks: func() {
				namespace := &models.Namespace{Name: "namespace", TenantID: "tenant"}

				device := &models.Device{
					UID:       "uid",
					Namespace: "namespace",
					TenantID:  "tenant",
					Tags:      []string{"device1", "device2"},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return([]models.Tags{
					{
						Name:   "device1",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device2",
						Color:  "",
						Tenant: "tenant",
					},
				}, int64(len(device.Tags)), nil).Once()
				mock.On("TagsDelete", ctx, "tenant", "device1").Return(int64(0), errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			name:   "success to delete tags",
			tag:    "device1",
			tenant: "tenant",
			requiredMocks: func() {
				namespace := &models.Namespace{Name: "namespace", TenantID: "tenant"}

				device := &models.Device{
					UID:       "uid",
					Namespace: "namespace",
					TenantID:  "tenant",
					Tags:      []string{"device1", "device2"},
				}

				mock.On("NamespaceGet", ctx, "tenant").Return(namespace, nil).Once()
				mock.On("TagsGet", ctx, "tenant").Return([]models.Tags{
					{
						Name:   "device1",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "device2",
						Color:  "",
						Tenant: "tenant",
					},
				}, int64(len(device.Tags)), nil).Once()
				mock.On("TagsDelete", ctx, "tenant", "device1").Return(int64(1), nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.requiredMocks()

			locator := &mocksGeoIp.Locator{}
			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, WithLocator(locator))

			err := service.DeleteTag(ctx, tc.tenant, tc.tag)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
