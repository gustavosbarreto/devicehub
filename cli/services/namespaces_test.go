package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/pkg/guard"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	env_mocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceCreate(t *testing.T) {
	mockClock := &clockmock.Clock{}

	clock.DefaultBackend = mockClock

	now := time.Now()

	mockClock.On("Now").Return(now)

	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	userNotFound := &models.User{
		UserData: models.UserData{
			Name:     "userNotFound",
			Email:    "userNotFound@userNotFound.com",
			Username: "usernameNotFound",
		},
	}
	namespace := &models.Namespace{
		Name:     "namespace",
		Owner:    user.ID,
		TenantID: "tenantID",
		Members:  []models.Member{{ID: user.ID, Role: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		MaxDevices: MaxNumberDevicesUnlimited,
		CreatedAt:  now,
	}
	namespaceInvalid := &models.Namespace{
		Name:     "namespaceInvalid@@",
		Owner:    "",
		TenantID: "tenantIDInvliad",
	}
	namespaceInvalidName := &models.Namespace{
		Name:     "namespaceUpCase",
		Owner:    "",
		TenantID: "tenantIDInvliad",
	}

	type Expected struct {
		user *models.Namespace
		err  error
	}
	tests := []struct {
		description   string
		namespace     string
		username      string
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails when could not find a user",
			namespace:   namespace.Name,
			username:    userNotFound.Username,
			tenantID:    namespace.TenantID,
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				mock.On("UserGetByUsername", ctx, userNotFound.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Fails when namespace is not valid",
			namespace:   namespaceInvalid.Name,
			username:    user.Username,
			tenantID:    namespaceInvalid.TenantID,
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
			},
			expected: Expected{nil, ErrNamespaceInvalid},
		},
		{
			description: "Fails when namespace is not valid due name",
			namespace:   namespaceInvalidName.Name,
			username:    user.Username,
			tenantID:    namespaceInvalidName.TenantID,
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
			},
			expected: Expected{nil, ErrNamespaceInvalid},
		},
		{
			description: "Fails when namespace is duplicated",
			namespace:   namespace.Name,
			username:    user.Username,
			tenantID:    namespace.TenantID,
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceCreate", ctx, namespace).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrDuplicateNamespace},
		},
		{
			description: "Success to create namespace - Community",
			namespace:   namespace.Name,
			username:    user.Username,
			tenantID:    namespace.TenantID,
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceCreate", ctx, namespace).Return(namespace, nil).Once()
			},
			expected: Expected{namespace, nil},
		},
		{
			description: "Success to create namespace - Cloud",
			namespace:   namespace.Name,
			username:    user.Username,
			tenantID:    namespace.TenantID,
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()
				namespace.MaxDevices = MaxNumberDevicesLimited
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceCreate", ctx, namespace).Return(namespace, nil).Once()
			},
			expected: Expected{namespace, nil},
		},
		{
			description: "Success to create namespace - Enterprise",
			namespace:   namespace.Name,
			username:    user.Username,
			tenantID:    namespace.TenantID,
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				namespace.MaxDevices = MaxNumberDevicesUnlimited
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceCreate", ctx, namespace).Return(namespace, nil).Once()
			},
			expected: Expected{namespace, nil},
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			ns, err := s.NamespaceCreate(ctx, test.namespace, test.username, test.tenantID)
			assert.Equal(t, test.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestAddUserNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	userNotFound := &models.User{
		UserData: models.UserData{
			Name:     "userNotFound",
			Email:    "userNotFound@userNotFound.com",
			Username: "usernameNotFound",
		},
	}
	namespace := &models.Namespace{
		Name:     "namespace",
		Owner:    user.ID,
		TenantID: "tenantID",
		Members:  []models.Member{{ID: user.ID, Role: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}

	namespaceNotFound := &models.Namespace{
		Name:     "namespaceNotFound",
		Owner:    user.ID,
		TenantID: "tenantIDNotFound",
		Members:  []models.Member{{ID: user.ID, Role: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}

	type Expected struct {
		user *models.Namespace
		err  error
	}
	tests := []struct {
		description   string
		username      string
		namespace     string
		role          string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails to find the user",
			username:    userNotFound.Username,
			namespace:   namespace.Name,
			role:        guard.RoleObserver,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, userNotFound.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Fails to find the namespace",
			username:    user.Username,
			namespace:   namespaceNotFound.Name,
			role:        guard.RoleObserver,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespaceNotFound.Name).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "Successfully add user to the Namespace",
			username:    user.Username,
			namespace:   namespace.Name,
			role:        guard.RoleObserver,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
				mock.On("NamespaceAddMember", ctx, namespace.TenantID, user.ID, guard.RoleObserver).Return(namespace, nil).Once()
			},
			expected: Expected{namespace, nil},
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			ns, err := s.NamespaceAddMember(ctx, test.username, test.namespace, test.role)
			assert.Equal(t, test.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDelUserNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")

	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	userNotFound := &models.User{
		UserData: models.UserData{
			Name:     "userNotFound",
			Email:    "userNotFound@userNotFound.com",
			Username: "usernameNotFound",
		},
	}
	namespace := &models.Namespace{
		Name:     "namespace",
		Owner:    user.ID,
		TenantID: "tenantID",
		Members:  []models.Member{{ID: user.ID, Role: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}
	namespaceNotFound := &models.Namespace{
		Name:     "namespaceNotFound",
		Owner:    user.ID,
		TenantID: "tenantIDNotFound",
		Members:  []models.Member{{ID: user.ID, Role: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}

	type Expected struct {
		user *models.Namespace
		err  error
	}
	tests := []struct {
		description   string
		username      string
		namespace     string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails to find the user",
			username:    userNotFound.Username,
			namespace:   namespace.Name,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, userNotFound.Username).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "Fails to find the namespace",
			username:    user.Username,
			namespace:   namespaceNotFound.Name,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespaceNotFound.Name).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "Fails remove member from the namespace",
			username:    user.Username,
			namespace:   namespace.Name,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
				mock.On("NamespaceRemoveMember", ctx, namespace.TenantID, user.ID).Return(nil, Err).Once()
			},
			expected: Expected{nil, ErrFailedNamespaceRemoveMember},
		},
		{
			description: "Successfully remove member from the namespace",
			username:    user.Username,
			namespace:   namespace.Name,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, user.Username).Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
				mock.On("NamespaceRemoveMember", ctx, namespace.TenantID, user.ID).Return(namespace, nil).Once()
			},
			expected: Expected{namespace, nil},
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			ns, err := s.NamespaceRemoveMember(ctx, test.username, test.namespace)
			assert.Equal(t, test.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDelNamespace(t *testing.T) {
	mock := &mocks.Store{}
	s := NewService(store.Store(mock))

	ctx := context.TODO()

	Err := errors.New("error")
	user := &models.User{
		ID: "userID",
		UserData: models.UserData{
			Name:     "user",
			Email:    "user@user.com",
			Username: "username",
		},
	}
	namespace := &models.Namespace{
		Name:     "namespace",
		Owner:    user.ID,
		TenantID: "tenantID",
		Members:  []models.Member{{ID: user.ID, Role: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}
	namespaceNotFound := &models.Namespace{
		Name:     "namespaceNotFound",
		Owner:    user.ID,
		TenantID: "tenantIDNotFound",
		Members:  []models.Member{{ID: user.ID, Role: "owner"}},
		Settings: &models.NamespaceSettings{
			SessionRecord: true,
		},
		CreatedAt: clock.Now(),
	}

	tests := []struct {
		description   string
		namespace     string
		requiredMocks func()
		expected      error
	}{
		{
			description: "Fails to find the namespace",
			namespace:   namespaceNotFound.Name,
			requiredMocks: func() {
				mock.On("NamespaceGetByName", ctx, namespaceNotFound.Name).Return(nil, Err).Once()
			},
			expected: ErrNamespaceNotFound,
		},
		{
			description: "Fails to delete the namespace",
			namespace:   namespace.Name,
			requiredMocks: func() {
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(Err).Once()
			},
			expected: ErrFailedDeleteNamespace,
		},
		{
			description: "Success to delete the namespace",
			namespace:   namespace.Name,
			requiredMocks: func() {
				mock.On("NamespaceGetByName", ctx, namespace.Name).Return(namespace, nil).Once()
				mock.On("NamespaceDelete", ctx, namespace.TenantID).Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, ts := range tests {
		test := ts
		t.Run(test.description, func(t *testing.T) {
			test.requiredMocks()
			err := s.NamespaceDelete(ctx, test.namespace)
			assert.Equal(t, test.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
