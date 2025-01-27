package services

import (
	"context"
	"testing"

	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/api/responses"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/clock"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/ssh"
)

const (
	InvalidTenantID        = "invalid_tenant_id"
	InvalidFingerprint     = "invalid_fingerprint"
	invalidTenantIDStr     = "Fails when the tenant is invalid"
	InvalidFingerprintStr  = "Fails when the fingerprint is invalid"
	InvalidFingerTenantStr = "Fails when the fingerprint and tenant is invalid"
)

func TestEvaluateKeyFilter(t *testing.T) {
	mock := &mocks.Store{}

	ctx := context.TODO()

	type Expected struct {
		bool
		error
	}

	cases := []struct {
		description   string
		key           *models.PublicKey
		device        models.Device
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail to evaluate when filter hostname no match",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Hostname: "roo.*",
					},
				},
			},
			device: models.Device{
				Name: "device",
			},
			requiredMocks: func() {
			},
			expected: Expected{false, nil},
		},
		{
			description: "success to evaluate filter hostname",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Hostname: ".*",
					},
				},
			},
			device: models.Device{
				Name: "device",
			},
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
		{
			description: "fail to evaluate filter tags when tag does not exist in device",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Tags: []string{"tag1", "tag2"},
					},
				},
			},
			device: models.Device{
				Tags: []string{"tag4"},
			},
			requiredMocks: func() {
			},
			expected: Expected{false, nil},
		},
		{
			description: "success to evaluate filter tags",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Tags: []string{"tag1", "tag2"},
					},
				},
			},
			device: models.Device{
				Tags: []string{"tag1"},
			},
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
		{
			description: "success to evaluate when key has no filter",
			key: &models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{},
				},
			},
			device: models.Device{},
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			ok, err := service.EvaluateKeyFilter(ctx, tc.key, tc.device)
			assert.Equal(t, tc.expected, Expected{ok, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestListPublicKeys(t *testing.T) {
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	ctx := context.TODO()

	keys := []models.PublicKey{
		{Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"}},
		{Data: []byte("teste2"), Fingerprint: "fingerprint2", CreatedAt: clock.Now(), TenantID: "tenant2", PublicKeyFields: models.PublicKeyFields{Name: "teste2"}},
	}

	type Expected struct {
		returnedKeys []models.PublicKey
		count        int
		err          error
	}

	cases := []struct {
		description   string
		keys          []models.PublicKey
		paginator     query.Paginator
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails when the query is invalid",
			paginator:   query.Paginator{Page: -1, PerPage: 10},
			requiredMocks: func() {
				mock.On("PublicKeyList", ctx, query.Paginator{Page: -1, PerPage: 10}).Return(nil, 0, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, 0, errors.New("error", "", 0)},
		},
		{
			description: "Successful list the keys",
			keys:        keys,
			paginator:   query.Paginator{Page: 1, PerPage: 10},
			requiredMocks: func() {
				mock.On("PublicKeyList", ctx, query.Paginator{Page: 1, PerPage: 10}).Return(keys, len(keys), nil).Once()
			},
			expected: Expected{keys, len(keys), nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			returnedKeys, count, err := s.ListPublicKeys(ctx, tc.paginator)
			assert.Equal(t, tc.expected, Expected{returnedKeys, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetPublicKeys(t *testing.T) {
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	ctx := context.TODO()

	type Expected struct {
		returnedKey *models.PublicKey
		err         error
	}

	cases := []struct {
		description   string
		ctx           context.Context
		fingerprint   string
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: invalidTenantIDStr,
			ctx:         ctx,
			fingerprint: "fingerprint",
			tenantID:    InvalidTenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, InvalidTenantID).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrNamespaceNotFound(InvalidTenantID, errors.New("error", "", 0))},
		},
		{
			description: InvalidFingerprintStr,
			ctx:         ctx,
			fingerprint: InvalidFingerprint,
			tenantID:    "tenant1",
			requiredMocks: func() {
				namespace := models.Namespace{TenantID: "tenant1"}

				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(&namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, InvalidFingerprint, "tenant1").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, errors.New("error", "", 0)},
		},
		{
			description: "Successful get the key",
			ctx:         ctx,
			fingerprint: "fingerprint",
			tenantID:    "tenant1",
			requiredMocks: func() {
				namespace := models.Namespace{TenantID: "tenant1"}
				key := models.PublicKey{
					Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
				}
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(&namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", "tenant1").Return(&key, nil).Once()
			},
			expected: Expected{&models.PublicKey{
				Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			returnedKey, err := s.GetPublicKey(ctx, tc.fingerprint, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{returnedKey, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestUpdatePublicKeys(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	type Expected struct {
		key *models.PublicKey
		err error
	}

	cases := []struct {
		description   string
		fingerprint   string
		tenantID      string
		keyUpdate     requests.PublicKeyUpdate
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail update the key when filter tags is empty",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Tags: []string{},
				},
			},
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "tenant").Return([]models.Tags{}, int64(0), errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrTagEmpty("tenant", errors.New("error", "", 0))},
		},
		{
			description: "fail to update the key when a tag does not exist in a device",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "tenant").
					Return([]models.Tags{
						{
							Name:   "tag1",
							Color:  "",
							Tenant: "tenant",
						},
						{
							Name:   "tag4",
							Color:  "",
							Tenant: "tenant",
						},
					}, int64(2), nil).Once()
			},
			expected: Expected{nil, NewErrTagNotFound("tag2", nil)},
		},
		{
			description: "Fail update the key when filter is tags",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				model := models.PublicKeyUpdate{
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				mock.On("TagsGet", ctx, "tenant").
					Return([]models.Tags{
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
					}, int64(2), nil).Once()
				mock.On("PublicKeyUpdate", ctx, "fingerprint", "tenant", &model).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, errors.New("error", "", 0)},
		},
		{
			description: "Successful update the key when filter is tags",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				model := models.PublicKeyUpdate{
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				keyUpdateWithTagsModel := &models.PublicKey{
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				mock.On("TagsGet", ctx, "tenant").
					Return([]models.Tags{
						{
							Name: "tag1",
						},
						{
							Name: "tag2",
						},
					}, int64(2), nil).Once()
				mock.On("PublicKeyUpdate", ctx, "fingerprint", "tenant", &model).Return(keyUpdateWithTagsModel, nil).Once()
			},
			expected: Expected{&models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Tags: []string{"tag1", "tag2"},
					},
				},
			}, nil},
		},
		{
			description: "Fail update the key when filter is hostname",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				model := models.PublicKeyUpdate{
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}

				mock.On("PublicKeyUpdate", ctx, "fingerprint", "tenant", &model).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, errors.New("error", "", 0)},
		},
		{
			description: "Successful update the key when filter is tags",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate: requests.PublicKeyUpdate{
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				model := models.PublicKeyUpdate{
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}

				keyUpdateWithHostnameModel := &models.PublicKey{
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}
				mock.On("PublicKeyUpdate", ctx, "fingerprint", "tenant", &model).Return(keyUpdateWithHostnameModel, nil).Once()
			},
			expected: Expected{&models.PublicKey{
				PublicKeyFields: models.PublicKeyFields{
					Filter: models.PublicKeyFilter{
						Hostname: ".*",
					},
				},
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			returnedKey, err := s.UpdatePublicKey(ctx, tc.fingerprint, tc.tenantID, tc.keyUpdate)
			assert.Equal(t, tc.expected, Expected{returnedKey, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestDeletePublicKeys(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	type Expected struct {
		err error
	}

	cases := []struct {
		description   string
		ctx           context.Context
		fingerprint   string
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: invalidTenantIDStr,
			ctx:         ctx,
			fingerprint: "fingerprint",
			tenantID:    InvalidTenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, InvalidTenantID).Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{NewErrNamespaceNotFound(InvalidTenantID, errors.New("error", "", 0))},
		},
		{
			description: InvalidFingerprintStr,
			ctx:         ctx,
			fingerprint: InvalidFingerprint,
			tenantID:    "tenant1",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "tenant1"}

				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, InvalidFingerprint, namespace.TenantID).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{NewErrPublicKeyNotFound(InvalidFingerprint, errors.New("error", "", 0))},
		},
		{
			description: "fail to delete the key",
			ctx:         ctx,
			fingerprint: "fingerprint",
			tenantID:    "tenant1",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "tenant1"}

				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", namespace.TenantID).
					Return(&models.PublicKey{
						Data:            []byte("teste"),
						Fingerprint:     "fingerprint",
						CreatedAt:       clock.Now(),
						TenantID:        "tenant1",
						PublicKeyFields: models.PublicKeyFields{Name: "teste"},
					}, nil).Once()
				mock.On("PublicKeyDelete", ctx, "fingerprint", "tenant1").
					Return(errors.New("error", "", 0)).Once()
			},
			expected: Expected{errors.New("error", "", 0)},
		},
		{
			description: "Successful to delete the key",
			ctx:         ctx,
			fingerprint: "fingerprint",
			tenantID:    "tenant1",
			requiredMocks: func() {
				namespace := &models.Namespace{TenantID: "tenant1"}

				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, "fingerprint", namespace.TenantID).
					Return(&models.PublicKey{
						Data:            []byte("teste"),
						Fingerprint:     "fingerprint",
						CreatedAt:       clock.Now(),
						TenantID:        "tenant1",
						PublicKeyFields: models.PublicKeyFields{Name: "teste"},
					}, nil).Once()
				mock.On("PublicKeyDelete", ctx, "fingerprint", "tenant1").Return(nil).Once()
			},
			expected: Expected{nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := s.DeletePublicKey(ctx, tc.fingerprint, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{err})
		})
	}

	mock.AssertExpectations(t)
}

func TestCreatePublicKeys(t *testing.T) {
	mock := new(mocks.Store)

	ctx := context.TODO()

	clockMock.On("Now").Return(now)

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	pubKey, _ := ssh.NewPublicKey(publicKey)

	type Expected struct {
		res *responses.PublicKeyCreate
		err error
	}

	cases := []struct {
		description   string
		tenantID      string
		req           requests.PublicKeyCreate
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail to create the key when filter tags is empty",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Tags: []string{},
				},
			},
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "tenant").Return([]models.Tags{}, int64(0), errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrTagEmpty("tenant", errors.New("error", "", 0))},
		},
		{
			description: "fail to create the key when a tags does not exist in a device",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2", "tag4"},
				},
			},
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "tenant").Return([]models.Tags{
					{
						Name:   "tag1",
						Color:  "",
						Tenant: "tenant",
					},
					{
						Name:   "tag4",
						Color:  "",
						Tenant: "tenant",
					},
				}, int64(2), nil).Once()
			},
			expected: Expected{nil, NewErrTagNotFound("tag2", nil)},
		},
		{
			description: "fail when data in public key is not valid",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        nil,
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
			},
			expected: Expected{nil, NewErrPublicKeyDataInvalid(requests.PublicKeyCreate{
				Data:        nil,
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			}.Data, nil)},
		},
		{
			description: "fail when cannot get the public key",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				keyWithHostname := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Hostname: ".*",
					},
				}

				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, NewErrPublicKeyNotFound(requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			}.Fingerprint, errors.New("error", "", 0))},
		},
		{
			description: "fail when public key is duplicated",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				keyWithHostname := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Hostname: ".*",
					},
				}

				keyWithHostnameModel := models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}

				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(&keyWithHostnameModel, nil).Once()
			},
			expected: Expected{nil, NewErrPublicKeyDuplicated([]string{ssh.FingerprintLegacyMD5(pubKey)}, nil)},
		},
		{
			description: "fail to create a public key when filter is hostname",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				keyWithHostnameModel := models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}

				keyWithHostname := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Hostname: ".*",
					},
				}

				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("PublicKeyCreate", ctx, &keyWithHostnameModel).Return(errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, errors.New("error", "", 0)},
		},
		{
			description: "success to create a public key when filter is hostname",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Hostname: ".*",
				},
			},
			requiredMocks: func() {
				keyWithHostnameModel := models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}

				keyWithHostname := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Hostname: ".*",
					},
				}

				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("PublicKeyCreate", ctx, &keyWithHostnameModel).Return(nil).Once()
			},
			expected: Expected{&responses.PublicKeyCreate{
				Data: models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}.Data,
				Filter: responses.PublicKeyFilter(models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}.Filter),
				Name: models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}.Name,
				Username: models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}.Username,
				TenantID: models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}.TenantID,
				Fingerprint: models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}.Fingerprint,
			}, nil},
		},
		{
			description: "fail to create a public key when filter is tags",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				keyWithTags := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Tags: []string{"tag1", "tag2"},
					},
				}

				keyWithTagsModel := models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				mock.On("TagsGet", ctx, keyWithTags.TenantID).Return([]models.Tags{
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
				}, int64(2), nil).Once()
				mock.On("PublicKeyGet", ctx, keyWithTags.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("PublicKeyCreate", ctx, &keyWithTagsModel).Return(errors.New("error", "", 0)).Once()
			},
			expected: Expected{nil, errors.New("error", "", 0)},
		},
		{
			description: "success to create a public key when filter is tags",
			tenantID:    "tenant",
			req: requests.PublicKeyCreate{
				Data:        ssh.MarshalAuthorizedKey(pubKey),
				Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
				TenantID:    "tenant",
				Filter: requests.PublicKeyFilter{
					Tags: []string{"tag1", "tag2"},
				},
			},
			requiredMocks: func() {
				keyWithTags := requests.PublicKeyCreate{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					TenantID:    "tenant",
					Filter: requests.PublicKeyFilter{
						Tags: []string{"tag1", "tag2"},
					},
				}

				keyWithTagsModel := models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				mock.On("TagsGet", ctx, keyWithTags.TenantID).Return([]models.Tags{
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
				}, int64(2), nil).Once()
				mock.On("PublicKeyGet", ctx, keyWithTags.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("PublicKeyCreate", ctx, &keyWithTagsModel).Return(nil).Once()
			},
			expected: Expected{&responses.PublicKeyCreate{
				Data: models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}.Data,
				Filter: responses.PublicKeyFilter(models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}.Filter),
				Name: models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}.Name,
				Username: models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}.Username,
				TenantID: models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}.TenantID,
				Fingerprint: models.PublicKey{
					Data:        ssh.MarshalAuthorizedKey(pubKey),
					Fingerprint: ssh.FingerprintLegacyMD5(pubKey),
					CreatedAt:   clock.Now(),
					TenantID:    "tenant",
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}.Fingerprint,
			}, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			res, err := s.CreatePublicKey(ctx, tc.req, tc.tenantID)
			assert.Equal(t, tc.expected, Expected{res, err})
		})
	}

	mock.AssertExpectations(t)
}
