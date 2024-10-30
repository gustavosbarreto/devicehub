package services

import (
	"context"
	"errors"
	"testing"

	"github.com/shellhub-io/shellhub/adm/pkg/inputs"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/hash"
	hashmock "github.com/shellhub-io/shellhub/pkg/hash/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestUserCreate(t *testing.T) {
	type Expected struct {
		user *models.User
		err  error
	}

	mock := new(mocks.Store)
	hashMock := &hashmock.Hasher{}
	hash.Backend = hashMock
	ctx := context.TODO()
	now := clock.Now()

	mockClock := new(clockmock.Clock)
	clock.DefaultBackend = mockClock
	mockClock.On("Now").Return(now)

	cases := []struct {
		description   string
		requiredMocks func()
		username      string
		password      string
		email         string
		expected      Expected
	}{
		{
			description: "fails when email is invalid",
			username:    "john_doe",
			email:       "invalidmail.com",
			password:    "secret",
			requiredMocks: func() {
			},
			expected: Expected{nil, ErrUserDataInvalid},
		},
		{
			description: "fails when username is invalid",
			username:    "",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
			},
			expected: Expected{nil, ErrUserDataInvalid},
		},
		{
			description: "fails when email is duplicated",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				mock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{"email"}, true, nil).
					Once()
			},
			expected: Expected{nil, ErrUserEmailExists},
		},
		{
			description: "fails when username is duplicated",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				mock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{"username"}, true, nil).
					Once()
			},
			expected: Expected{nil, ErrUserNameExists},
		},
		{
			description: "fails when email and username is duplicated",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				mock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{"username", "email"}, true, nil).
					Once()
			},
			expected: Expected{nil, ErrUserNameAndEmailExists},
		},
		{
			description: "fails when the password is invalid",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				mock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{}, false, nil).
					Once()
				hashMock.
					On("Do", "secret").
					Return("", errors.New("error")).
					Once()
			},
			expected: Expected{nil, ErrUserPasswordInvalid},
		},
		{
			description: "fails creates a user",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				mock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{}, false, nil).
					Once()
				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					UserData: models.UserData{
						Name:     "john_doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Status:        models.UserStatusConfirmed,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserCreate", ctx, user).Return("", errors.New("error")).Once()
			},
			expected: Expected{nil, ErrCreateNewUser},
		},
		{
			description: "successfully creates a user",
			username:    "john_doe",
			email:       "john.doe@test.com",
			password:    "secret",
			requiredMocks: func() {
				mock.
					On("UserConflicts", ctx, &models.UserConflicts{Username: "john_doe", Email: "john.doe@test.com"}).
					Return([]string{}, false, nil).
					Once()
				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi", nil).
					Once()

				user := &models.User{
					UserData: models.UserData{
						Name:     "john_doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
					Password: models.UserPassword{
						Plain: "secret",
						Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
					},
					Status:        models.UserStatusConfirmed,
					CreatedAt:     clock.Now(),
					MaxNamespaces: MaxNumberNamespacesCommunity,
				}
				mock.On("UserCreate", ctx, user).Return("000000000000000000000000", nil).Once()
			},
			expected: Expected{&models.User{
				UserData: models.UserData{
					Name:     "john_doe",
					Email:    "john.doe@test.com",
					Username: "john_doe",
				},
				Password: models.UserPassword{
					Plain: "secret",
					Hash:  "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YVVCIa2UYuFV4OJby7Yi",
				},
				Status:        models.UserStatusConfirmed,
				CreatedAt:     clock.Now(),
				MaxNamespaces: MaxNumberNamespacesCommunity,
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock))
			user, err := service.UserCreate(ctx, &inputs.UserCreate{Username: tc.username, Password: tc.password, Email: tc.email})

			assert.Equal(t, tc.expected, Expected{user, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestUserDelete(t *testing.T) {
	mock := new(mocks.Store)
	ctx := context.TODO()

	cases := []struct {
		description   string
		username      string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not find a user",
			username:    "john_doe",
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "john_doe").Return(nil, errors.New("error")).Once()
			},
			expected: ErrUserNotFound,
		},
		{
			description: "fails to delete the user and associated data when namespace not found",
			username:    "john_doe",
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
				mock.On("UserGetInfo", ctx, "507f191e810c19729de860ea").Return(nil, errors.New("error")).Once()
			},
			expected: ErrNamespaceNotFound,
		},
		{
			description: "successfully delete the user and associated data",
			username:    "john_doe",
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()

				namespaceOwned := []models.Namespace{
					{
						Name:     "namespace1",
						Owner:    "507f191e810c19729de860ea",
						TenantID: "10000000-0000-0000-0000-000000000000",
						Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
						Settings: &models.NamespaceSettings{
							SessionRecord: true,
						},
						CreatedAt: clock.Now(),
					},
					{
						Name:     "namespace2",
						Owner:    "507f191e810c19729de860ea",
						TenantID: "20000000-0000-0000-0000-000000000000",
						Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
						Settings: &models.NamespaceSettings{
							SessionRecord: true,
						},
						CreatedAt: clock.Now(),
					},
				}
				namespaceMember := []models.Namespace{
					{
						Name:     "namespace3",
						Owner:    "507f191e810c19729de86000",
						TenantID: "30000000-0000-0000-0000-000000000000",
						Members: []models.Member{
							{ID: "507f191e810c19729de86000", Role: authorizer.RoleObserver},
							{ID: "507f191e810c19729de860ea", Role: authorizer.RoleObserver},
						},
						Settings: &models.NamespaceSettings{
							SessionRecord: true,
						},
						CreatedAt: clock.Now(),
					},
					{
						Name:     "namespace1",
						Owner:    "507f191e810c19729de86000",
						TenantID: "tenantID1",
						Members: []models.Member{
							{ID: "507f191e810c19729de86000", Role: authorizer.RoleObserver},
							{ID: "507f191e810c19729de860ea", Role: authorizer.RoleObserver},
						},
						Settings: &models.NamespaceSettings{
							SessionRecord: true,
						},
						CreatedAt: clock.Now(),
					},
				}

				mock.On("UserGetInfo", ctx, "507f191e810c19729de860ea").Return(&models.UserInfo{
					OwnedNamespaces:      namespaceOwned,
					AssociatedNamespaces: namespaceMember,
				}, nil)

				for _, v := range namespaceOwned {
					mock.On("NamespaceDelete", ctx, v.TenantID).Return(nil).Once()
				}
				for _, v := range namespaceMember {
					mock.On("NamespaceRemoveMember", ctx, v.TenantID, "507f191e810c19729de860ea").Return(nil, nil).Once()
				}

				mock.On("UserDelete", ctx, "507f191e810c19729de860ea").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock))
			err := service.UserDelete(ctx, &inputs.UserDelete{Username: tc.username})
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestUserResetPassword(t *testing.T) {
	mock := new(mocks.Store)
	hashMock := &hashmock.Hasher{}
	hash.Backend = hashMock

	ctx := context.TODO()

	cases := []struct {
		description   string
		username      string
		password      string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not find a user",
			username:    "john_doe",
			password:    "password",
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "john_doe").Return(nil, errors.New("error")).Once()
			},
			expected: ErrUserNotFound,
		},
		{
			description: "fails to reset the user password",
			username:    "john_doe",
			password:    "secret",
			requiredMocks: func() {
				user := &models.User{ID: "507f191e810c19729de860ea"}

				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()

				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi", nil).
					Once()

				mock.
					On("UserUpdate", ctx, "507f191e810c19729de860ea", &models.UserChanges{Password: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi"}).
					Return(errors.New("error")).
					Once()
			},
			expected: ErrFailedUpdateUser,
		},
		{
			description: "successfully reset the user password",
			username:    "john_doe",
			password:    "secret",
			requiredMocks: func() {
				user := &models.User{ID: "507f191e810c19729de860ea"}

				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()

				hashMock.
					On("Do", "secret").
					Return("$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi", nil).
					Once()

				mock.
					On("UserUpdate", ctx, "507f191e810c19729de860ea", &models.UserChanges{Password: "$2a$10$V/6N1wsjheBVvWosPfv02uf4WAOb9lmp8YWQCIa2UYuFV4OJby7Yi"}).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			service := NewService(store.Store(mock))
			err := service.UserUpdate(ctx, &inputs.UserUpdate{Username: tc.username, Password: tc.password})
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
