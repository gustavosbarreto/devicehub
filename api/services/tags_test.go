package services

import (
	"context"
	"errors"
	"testing"

	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateTag(t *testing.T) {
	storeMock := new(storemock.Store)
	ctx := context.TODO()

	type Expected struct {
		insertedID string
		conflicts  []string
		err        error
	}

	cases := []struct {
		description   string
		req           *requests.CreateTag
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when namespace not found",
			req: &requests.CreateTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				insertedID: "",
				conflicts:  []string{},
				err:        NewErrNamespaceNotFound("tenant1", errors.New("error")),
			},
		},
		{
			description: "fails when tag name conflicts",
			req: &requests.CreateTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, "tenant1", &models.TagConflicts{Name: "production"}).
					Return([]string{"name"}, true, nil).
					Once()
			},
			expected: Expected{
				insertedID: "",
				conflicts:  []string{"name"},
				err:        nil,
			},
		},
		{
			description: "fails when tag create fails",
			req: &requests.CreateTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, "tenant1", &models.TagConflicts{Name: "production"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("TagCreate", ctx, &models.Tag{Name: "production", TenantID: "tenant1"}).
					Return("", errors.New("error")).
					Once()
			},
			expected: Expected{
				insertedID: "",
				conflicts:  []string{},
				err:        errors.New("error"),
			},
		},
		{
			description: "succeeds creating tag",
			req: &requests.CreateTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, "tenant1", &models.TagConflicts{Name: "production"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("TagCreate", ctx, &models.Tag{Name: "production", TenantID: "tenant1"}).
					Return("000000000000000000000000", nil).
					Once()
			},
			expected: Expected{
				insertedID: "000000000000000000000000",
				conflicts:  []string{},
				err:        nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			insertedID, conflicts, err := service.CreateTag(ctx, tc.req)
			require.Equal(t, tc.expected, Expected{insertedID, conflicts, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_PushTagTo(t *testing.T) {
	storeMock := new(storemock.Store)
	ctx := context.TODO()

	cases := []struct {
		description   string
		target        models.TagTarget
		req           *requests.PushTag
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when namespace not found",
			target:      models.TagTargetDevice,
			req: &requests.PushTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "dev1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrNamespaceNotFound("tenant1", errors.New("error")),
		},
		{
			description: "fails when tag not found",
			target:      models.TagTargetDevice,
			req: &requests.PushTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "dev1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrTagNotFound("production", errors.New("error")),
		},
		{
			description: "fails when tag push fails",
			target:      models.TagTargetDevice,
			req: &requests.PushTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "dev1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(&models.Tag{}, nil).
					Once()
				storeMock.
					On("TagPushToTarget", ctx, "tenant1", "production", models.TagTargetDevice, "dev1").
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds pushing tag",
			target:      models.TagTargetDevice,
			req: &requests.PushTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "dev1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(&models.Tag{}, nil).
					Once()
				storeMock.
					On("TagPushToTarget", ctx, "tenant1", "production", models.TagTargetDevice, "dev1").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := service.PushTagTo(ctx, tc.target, tc.req)
			require.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_PullTagFrom(t *testing.T) {
	storeMock := new(storemock.Store)
	ctx := context.TODO()

	cases := []struct {
		description   string
		target        models.TagTarget
		req           *requests.PullTag
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when namespace not found",
			target:      models.TagTargetDevice,
			req: &requests.PullTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "dev1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrNamespaceNotFound("tenant1", errors.New("error")),
		},
		{
			description: "fails when tag not found",
			target:      models.TagTargetDevice,
			req: &requests.PullTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "dev1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrTagNotFound("production", errors.New("error")),
		},
		{
			description: "fails when tag pull fails",
			target:      models.TagTargetDevice,
			req: &requests.PullTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "dev1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(&models.Tag{}, nil).
					Once()
				storeMock.
					On("TagPullFromTarget", ctx, "tenant1", "production", models.TagTargetDevice, "dev1").
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds pulling tag",
			target:      models.TagTargetDevice,
			req: &requests.PullTag{
				Name:     "production",
				TenantID: "tenant1",
				TargetID: "dev1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(&models.Tag{}, nil).
					Once()
				storeMock.
					On("TagPullFromTarget", ctx, "tenant1", "production", models.TagTargetDevice, "dev1").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := service.PullTagFrom(ctx, tc.target, tc.req)
			require.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_ListTags(t *testing.T) {
	storeMock := new(storemock.Store)
	ctx := context.TODO()

	type Expected struct {
		tags       []models.Tag
		totalCount int
		err        error
	}

	cases := []struct {
		description   string
		req           *requests.ListTags
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when namespace not found",
			req: &requests.ListTags{
				TenantID: "tenant1",
				Paginator: query.Paginator{
					Page:    1,
					PerPage: 10,
				},
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				tags:       []models.Tag{},
				totalCount: 0,
				err:        NewErrNamespaceNotFound("tenant1", errors.New("error")),
			},
		},
		{
			description: "fails when tag list fails",
			req: &requests.ListTags{
				TenantID: "tenant1",
				Paginator: query.Paginator{
					Page:    1,
					PerPage: 10,
				},
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagList", ctx, "tenant1", query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{}).
					Return(nil, 0, errors.New("error")).
					Once()
			},
			expected: Expected{
				tags:       []models.Tag{},
				totalCount: 0,
				err:        errors.New("error"),
			},
		},
		{
			description: "succeeds listing tags",
			req: &requests.ListTags{
				TenantID: "tenant1",
				Paginator: query.Paginator{
					Page:    1,
					PerPage: 10,
				},
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagList", ctx, "tenant1", query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{}).
					Return([]models.Tag{{Name: "production", TenantID: "tenant1"}}, 1, nil).
					Once()
			},
			expected: Expected{
				tags:       []models.Tag{{Name: "production", TenantID: "tenant1"}},
				totalCount: 1,
				err:        nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			tags, count, err := service.ListTags(ctx, tc.req)
			require.Equal(t, tc.expected, Expected{tags, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_UpdateTag(t *testing.T) {
	storeMock := new(storemock.Store)
	ctx := context.TODO()

	type Expected struct {
		conflicts []string
		err       error
	}

	cases := []struct {
		description   string
		req           *requests.UpdateTag
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when namespace not found",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				conflicts: []string{},
				err:       NewErrNamespaceNotFound("tenant1", errors.New("error")),
			},
		},
		{
			description: "fails when tag not found",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: Expected{
				conflicts: []string{},
				err:       NewErrTagNotFound("production", errors.New("error")),
			},
		},
		{
			description: "fails when new name conflicts",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(&models.Tag{}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, "tenant1", &models.TagConflicts{Name: "staging"}).
					Return([]string{"name"}, true, nil).
					Once()
			},
			expected: Expected{
				conflicts: []string{"name"},
				err:       nil,
			},
		},
		{
			description: "fails when tag update fails",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(&models.Tag{}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, "tenant1", &models.TagConflicts{Name: "staging"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("TagUpdate", ctx, "tenant1", "production", &models.TagChanges{Name: "staging"}).
					Return(errors.New("error")).
					Once()
			},
			expected: Expected{
				conflicts: nil,
				err:       errors.New("error"),
			},
		},
		{
			description: "succeeds updating tag",
			req: &requests.UpdateTag{
				Name:     "production",
				NewName:  "staging",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(&models.Tag{}, nil).
					Once()
				storeMock.
					On("TagConflicts", ctx, "tenant1", &models.TagConflicts{Name: "staging"}).
					Return([]string{}, false, nil).
					Once()
				storeMock.
					On("TagUpdate", ctx, "tenant1", "production", &models.TagChanges{Name: "staging"}).
					Return(nil).
					Once()
			},
			expected: Expected{
				conflicts: []string{},
				err:       nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			conflicts, err := service.UpdateTag(ctx, tc.req)
			require.Equal(t, tc.expected, Expected{conflicts, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_DeleteTag(t *testing.T) {
	storeMock := new(storemock.Store)
	ctx := context.TODO()

	cases := []struct {
		description   string
		req           *requests.DeleteTag
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when namespace not found",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrNamespaceNotFound("tenant1", errors.New("error")),
		},
		{
			description: "fails when tag not found",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(nil, errors.New("error")).
					Once()
			},
			expected: NewErrTagNotFound("production", errors.New("error")),
		},
		{
			description: "fails when transaction fails",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(&models.Tag{}, nil).
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds deleting tag",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				storeMock.
					On("NamespaceGet", ctx, "tenant1").
					Return(&models.Namespace{}, nil).
					Once()
				storeMock.
					On("TagGetByName", ctx, "tenant1", "production").
					Return(&models.Tag{}, nil).
					Once()
				storeMock.
					On("WithTransaction", ctx, mock.AnythingOfType("store.TransactionCb")).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := service.DeleteTag(ctx, tc.req)
			require.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestService_deleteTagCallback(t *testing.T) {
	storeMock := new(storemock.Store)
	ctx := context.TODO()

	cases := []struct {
		description   string
		req           *requests.DeleteTag
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when tag pull fails",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				for _, target := range models.TagTargets() {
					storeMock.
						On("TagPullFromTarget", ctx, "tenant1", "production", target).
						Return(errors.New("error")).
						Once()

					break
				}
			},
			expected: errors.New("error"),
		},
		{
			description: "fails when tag delete fails",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				for _, target := range models.TagTargets() {
					storeMock.
						On("TagPullFromTarget", ctx, "tenant1", "production", target).
						Return(nil).
						Once()
				}

				storeMock.
					On("TagDelete", ctx, "tenant1", "production").
					Return(errors.New("error")).
					Once()
			},
			expected: errors.New("error"),
		},
		{
			description: "succeeds",
			req: &requests.DeleteTag{
				Name:     "production",
				TenantID: "tenant1",
			},
			requiredMocks: func() {
				for _, target := range models.TagTargets() {
					storeMock.
						On("TagPullFromTarget", ctx, "tenant1", "production", target).
						Return(nil).
						Once()
				}

				storeMock.
					On("TagDelete", ctx, "tenant1", "production").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	service := NewService(storeMock, privateKey, publicKey, nil, nil)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			callback := service.deleteTagCallback(tc.req)
			require.Equal(t, tc.expected, callback(ctx))
		})
	}

	storeMock.AssertExpectations(t)
}
