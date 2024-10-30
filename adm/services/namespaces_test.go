package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/adm/pkg/inputs"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/clock"
	clockmock "github.com/shellhub-io/shellhub/pkg/clock/mocks"
	"github.com/shellhub-io/shellhub/pkg/envs"
	env_mocks "github.com/shellhub-io/shellhub/pkg/envs/mocks"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceCreate(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}
	now := time.Now()

	mock := new(mocks.Store)
	ctx := context.TODO()

	mockClock := new(clockmock.Clock)
	clock.DefaultBackend = mockClock
	mockClock.On("Now").Return(now)

	cases := []struct {
		description   string
		namespace     string
		username      string
		tenant        string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when namespace is not valid",
			namespace:   "",
			username:    "john_doe",
			tenant:      "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
			},
			expected: Expected{nil, ErrNamespaceInvalid},
		},
		{
			description: "fails when namespace is not valid due name",
			namespace:   "invalid_namespace",
			username:    "john_doe",
			tenant:      "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
			},
			expected: Expected{nil, ErrNamespaceInvalid},
		},
		{
			description: "fails when could not find a user",
			namespace:   "namespace",
			username:    "john_doe",
			tenant:      "00000000-0000-4000-0000-000000000000",
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				mock.On("UserGetByUsername", ctx, "john_doe").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "fails when namespace is duplicated",
			namespace:   "namespace",
			username:    "john_doe",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members: []models.Member{
						{
							ID:      "507f191e810c19729de860ea",
							Role:    "owner",
							AddedAt: now,
							Status:  models.MemberStatusAccepted,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: models.DefaultAnnouncementMessage,
					},
					MaxDevices: MaxNumberDevicesUnlimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceCreate", ctx, namespace).Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrDuplicateNamespace},
		},
		{
			description: "succeeds in creating a namespace when user and namespace data are valid - Community",
			namespace:   "namespace",
			username:    "john_doe",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members: []models.Member{
						{
							ID:      "507f191e810c19729de860ea",
							Role:    "owner",
							AddedAt: now,
							Status:  models.MemberStatusAccepted,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: models.DefaultAnnouncementMessage,
					},
					MaxDevices: MaxNumberDevicesUnlimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceCreate", ctx, namespace).Return(namespace, nil).Once()
			},
			expected: Expected{&models.Namespace{
				Name:     "namespace",
				Owner:    "507f191e810c19729de860ea",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Members: []models.Member{
					{
						ID:      "507f191e810c19729de860ea",
						Role:    "owner",
						AddedAt: now,
						Status:  models.MemberStatusAccepted,
					},
				},
				Settings: &models.NamespaceSettings{
					SessionRecord:          true,
					ConnectionAnnouncement: models.DefaultAnnouncementMessage,
				},
				MaxDevices: MaxNumberDevicesUnlimited,
				CreatedAt:  now,
			}, nil},
		},
		{
			description: "succeeds in creating a namespace when user and namespace data are valid - Cloud",
			namespace:   "namespace",
			username:    "john_doe",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("false").Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("true").Once()
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members: []models.Member{
						{
							ID:      "507f191e810c19729de860ea",
							Role:    "owner",
							AddedAt: now,
							Status:  models.MemberStatusAccepted,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: models.DefaultAnnouncementMessage,
					},
					MaxDevices: MaxNumberDevicesLimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceCreate", ctx, namespace).Return(namespace, nil).Once()
			},
			expected: Expected{&models.Namespace{
				Name:     "namespace",
				Owner:    "507f191e810c19729de860ea",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Members: []models.Member{
					{
						ID:      "507f191e810c19729de860ea",
						Role:    "owner",
						AddedAt: now,
						Status:  models.MemberStatusAccepted,
					},
				},
				Settings: &models.NamespaceSettings{
					SessionRecord:          true,
					ConnectionAnnouncement: models.DefaultAnnouncementMessage,
				},
				MaxDevices: MaxNumberDevicesLimited,
				CreatedAt:  now,
			}, nil},
		},
		{
			description: "succeeds in creating a namespace when user and namespace data are valid - Enterprise",
			namespace:   "namespace",
			username:    "john_doe",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				envMock := &env_mocks.Backend{}
				envs.DefaultBackend = envMock
				envMock.On("Get", "SHELLHUB_ENTERPRISE").Return("true").Once()
				envMock.On("Get", "SHELLHUB_CLOUD").Return("false").Once()
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john_doe",
					},
				}
				mock.On("UserGetByUsername", ctx, "john_doe").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members: []models.Member{
						{
							ID:      "507f191e810c19729de860ea",
							Role:    "owner",
							AddedAt: now,
							Status:  models.MemberStatusAccepted,
						},
					},
					Settings: &models.NamespaceSettings{
						SessionRecord:          true,
						ConnectionAnnouncement: models.DefaultAnnouncementMessage,
					},
					MaxDevices: MaxNumberDevicesUnlimited,
					CreatedAt:  now,
				}
				mock.On("NamespaceCreate", ctx, namespace).Return(namespace, nil).Once()
			},
			expected: Expected{&models.Namespace{
				Name:     "namespace",
				Owner:    "507f191e810c19729de860ea",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Members: []models.Member{
					{
						ID:      "507f191e810c19729de860ea",
						Role:    "owner",
						AddedAt: now,
						Status:  models.MemberStatusAccepted,
					},
				},
				Settings: &models.NamespaceSettings{
					SessionRecord:          true,
					ConnectionAnnouncement: models.DefaultAnnouncementMessage,
				},
				MaxDevices: MaxNumberDevicesUnlimited,
				CreatedAt:  now,
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock))
			ns, err := s.NamespaceCreate(ctx, &inputs.NamespaceCreate{Namespace: tc.namespace, Owner: tc.username, TenantID: tc.tenant})
			assert.Equal(t, tc.expected, Expected{ns, err})

			mock.AssertExpectations(t)
		})
	}
}

func TestNamespaceAddMember(t *testing.T) {
	type Expected struct {
		namespace *models.Namespace
		err       error
	}

	mock := new(mocks.Store)

	ctx := context.TODO()
	now := time.Now()

	cases := []struct {
		description   string
		username      string
		namespace     string
		role          authorizer.Role
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when could not find a user",
			username:    "john",
			namespace:   "namespace",
			role:        authorizer.RoleObserver,
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "john").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "fails when could not find a namespace",
			username:    "john",
			namespace:   "invalid_namespace",
			role:        authorizer.RoleObserver,
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john",
					},
				}
				mock.On("UserGetByUsername", ctx, "john").Return(user, nil).Once()
				mock.On("NamespaceGetByName", ctx, "invalid_namespace").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "successfully add user to the Namespace",
			username:    "john",
			namespace:   "namespace",
			role:        authorizer.RoleObserver,
			requiredMocks: func() {
				user := &models.User{
					ID: "507f191e810c19729de860ea",
					UserData: models.UserData{
						Name:     "John Doe",
						Email:    "john.doe@test.com",
						Username: "john",
					},
				}
				mock.On("UserGetByUsername", ctx, "john").Return(user, nil).Once()
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					CreatedAt: now,
				}
				mock.On("NamespaceGetByName", ctx, "namespace").Return(namespace, nil).Once()
				mock.On("NamespaceAddMember", ctx, "00000000-0000-0000-0000-000000000000", &models.Member{ID: "507f191e810c19729de860ea", Role: authorizer.RoleObserver}).Return(nil).Once()
			},
			expected: Expected{&models.Namespace{
				Name:     "namespace",
				Owner:    "507f191e810c19729de860ea",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
				Settings: &models.NamespaceSettings{
					SessionRecord: true,
				},
				CreatedAt: now,
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock))
			ns, err := s.NamespaceAddMember(ctx, &inputs.MemberAdd{Username: tc.username, Namespace: tc.namespace, Role: tc.role})
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestNamespaceRemoveMember(t *testing.T) {
	type Expected struct {
		user *models.Namespace
		err  error
	}

	mock := new(mocks.Store)

	ctx := context.TODO()
	now := time.Now()

	cases := []struct {
		description   string
		username      string
		namespace     string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when could not find a user",
			username:    "john_doe",
			namespace:   "namespace",
			requiredMocks: func() {
				mock.On("UserGetByUsername", ctx, "john_doe").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrUserNotFound},
		},
		{
			description: "fails when could not find a namespace",
			username:    "john_doe",
			namespace:   "namespace",
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
				mock.On("NamespaceGetByName", ctx, "namespace").Return(nil, errors.New("error")).Once()
			},
			expected: Expected{nil, ErrNamespaceNotFound},
		},
		{
			description: "fails remove member from the namespace",
			username:    "john_doe",
			namespace:   "namespace",
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
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					CreatedAt: now,
				}
				mock.On("NamespaceGetByName", ctx, "namespace").Return(namespace, nil).Once()
				mock.On("NamespaceRemoveMember", ctx, "00000000-0000-0000-0000-000000000000", "507f191e810c19729de860ea").Return(errors.New("error")).Once()
			},
			expected: Expected{nil, ErrFailedNamespaceRemoveMember},
		},
		{
			description: "successfully remove member from the namespace",
			username:    "john_doe",
			namespace:   "namespace",
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
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					CreatedAt: now,
				}
				mock.On("NamespaceGetByName", ctx, "namespace").Return(namespace, nil).Once()
				mock.On("NamespaceRemoveMember", ctx, "00000000-0000-0000-0000-000000000000", "507f191e810c19729de860ea").Return(nil).Once()
			},
			expected: Expected{&models.Namespace{
				Name:     "namespace",
				Owner:    "507f191e810c19729de860ea",
				TenantID: "00000000-0000-0000-0000-000000000000",
				Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
				Settings: &models.NamespaceSettings{
					SessionRecord: true,
				},
				CreatedAt: now,
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock))
			ns, err := s.NamespaceRemoveMember(ctx, &inputs.MemberRemove{Username: tc.username, Namespace: tc.namespace})
			assert.Equal(t, tc.expected, Expected{ns, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestNamespaceDelete(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		namespace     string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not find a namespace",
			namespace:   "namespace",
			requiredMocks: func() {
				mock.On("NamespaceGetByName", ctx, "namespace").Return(nil, errors.New("error")).Once()
			},
			expected: ErrNamespaceNotFound,
		},
		{
			description: "fails to delete the namespace",
			namespace:   "namespace",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					MaxDevices: MaxNumberDevicesUnlimited,
					CreatedAt:  clock.Now(),
				}
				mock.On("NamespaceGetByName", ctx, "namespace").Return(namespace, nil).Once()
				mock.On("NamespaceDelete", ctx, "00000000-0000-0000-0000-000000000000").Return(errors.New("error")).Once()
			},
			expected: ErrFailedDeleteNamespace,
		},
		{
			description: "success to delete the namespace",
			namespace:   "namespace",
			requiredMocks: func() {
				namespace := &models.Namespace{
					Name:     "namespace",
					Owner:    "507f191e810c19729de860ea",
					TenantID: "00000000-0000-0000-0000-000000000000",
					Members:  []models.Member{{ID: "507f191e810c19729de860ea", Role: "owner"}},
					Settings: &models.NamespaceSettings{
						SessionRecord: true,
					},
					MaxDevices: MaxNumberDevicesUnlimited,
					CreatedAt:  clock.Now(),
				}
				mock.On("NamespaceGetByName", ctx, "namespace").Return(namespace, nil).Once()
				mock.On("NamespaceDelete", ctx, "00000000-0000-0000-0000-000000000000").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			s := NewService(store.Store(mock))
			err := s.NamespaceDelete(ctx, &inputs.NamespaceDelete{Namespace: tc.namespace})
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}
