package services

import (
	"context"
	"testing"

	storecache "github.com/shellhub-io/shellhub/api/cache"
	"github.com/shellhub-io/shellhub/api/store"
	"github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/paginator"
	"github.com/shellhub-io/shellhub/pkg/api/request"
	"github.com/shellhub-io/shellhub/pkg/api/response"
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
	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	type Expected struct {
		bool
		error
	}

	keyTagsNoExist := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2"},
			},
		},
	}
	deviceTagsNoExist := models.Device{
		Tags: []string{"tag4"},
	}

	keyTags := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2"},
			},
		},
	}
	deviceTags := models.Device{
		Tags: []string{"tag1"},
	}

	keyHostname := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
			},
		},
	}
	deviceHostname := models.Device{
		Name: "device",
	}

	keyHostnameNoMatch := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: "roo.*",
			},
		},
	}
	deviceHostnameNoMatch := models.Device{
		Name: "device",
	}

	keyNoFilter := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{},
		},
	}
	deviceNoFilter := models.Device{}

	cases := []struct {
		description   string
		key           *models.PublicKey
		device        models.Device
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail to evaluate when filter hostname no match",
			key:         keyHostnameNoMatch,
			device:      deviceHostnameNoMatch,
			requiredMocks: func() {
			},
			expected: Expected{false, nil},
		},
		{
			description: "success to evaluate filter hostname",
			key:         keyHostname,
			device:      deviceHostname,
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
		{
			description: "fail to evaluate filter tags when tag does not exist in device",
			key:         keyTagsNoExist,
			device:      deviceTagsNoExist,
			requiredMocks: func() {
			},
			expected: Expected{false, nil},
		},
		{
			description: "success to evaluate filter tags",
			key:         keyTags,
			device:      deviceTags,
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
		{
			description: "success to evaluate when key has no filter",
			key:         keyNoFilter,
			device:      deviceNoFilter,
			requiredMocks: func() {
			},
			expected: Expected{true, nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			ok, err := s.EvaluateKeyFilter(ctx, tc.key, tc.device)
			assert.Equal(t, tc.expected, Expected{ok, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestListPublicKeys(t *testing.T) {
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	keys := []models.PublicKey{
		{Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"}},
		{Data: []byte("teste2"), Fingerprint: "fingerprint2", CreatedAt: clock.Now(), TenantID: "tenant2", PublicKeyFields: models.PublicKeyFields{Name: "teste2"}},
	}

	validQuery := paginator.Query{Page: 1, PerPage: 10}
	invalidQuery := paginator.Query{Page: -1, PerPage: 10}

	Err := errors.New("error", "", 0)

	type Expected struct {
		returnedKeys []models.PublicKey
		count        int
		err          error
	}

	cases := []struct {
		description   string
		ctx           context.Context
		keys          []models.PublicKey
		query         paginator.Query
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "Fails when the query is invalid",
			ctx:         ctx,
			keys:        keys,
			query:       invalidQuery,
			requiredMocks: func() {
				mock.On("PublicKeyList", ctx, invalidQuery).Return(nil, 0, Err).Once()
			},
			expected: Expected{nil, 0, Err},
		},
		{
			description: "Successful list the keys",
			ctx:         ctx,
			keys:        keys,
			query:       validQuery,
			requiredMocks: func() {
				mock.On("PublicKeyList", ctx, validQuery).Return(keys, len(keys), nil).Once()
			},
			expected: Expected{keys, len(keys), nil},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()
			returnedKeys, count, err := s.ListPublicKeys(ctx, tc.query)
			assert.Equal(t, tc.expected, Expected{returnedKeys, count, err})
		})
	}

	mock.AssertExpectations(t)
}

func TestGetPublicKeys(t *testing.T) {
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	namespace := models.Namespace{TenantID: "tenant1"}

	key := models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}

	Err := errors.New("error", "", 0)

	type Expected struct {
		returnedKey *models.PublicKey
		err         error
	}

	cases := []struct {
		description   string
		ctx           context.Context
		key           *models.PublicKey
		fingerprint   string
		tenantID      string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: invalidTenantIDStr,
			ctx:         ctx,
			key:         nil,
			fingerprint: key.Fingerprint,
			tenantID:    InvalidTenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, InvalidTenantID).Return(nil, Err).Once()
			},
			expected: Expected{nil, NewErrNamespaceNotFound(InvalidTenantID, Err)},
		},
		{
			description: InvalidFingerprintStr,
			ctx:         ctx,
			key:         nil,
			fingerprint: InvalidFingerprint,
			tenantID:    key.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(&namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, InvalidFingerprint, key.TenantID).Return(nil, Err).Once()
			},
			expected: Expected{nil, Err},
		},
		{
			description: "Successful get the key",
			ctx:         ctx,
			key:         &key,
			fingerprint: key.Fingerprint,
			tenantID:    key.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(&namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, key.Fingerprint, key.TenantID).Return(&key, nil).Once()
			},
			expected: Expected{&key, nil},
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
	mock := &mocks.Store{}

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()
	err := errors.New("error", "", 0)

	keyUpdateWithTags := request.PublicKeyUpdate{
		Filter: request.PublicKeyFilter{
			Tags: []string{"tag1", "tag2"},
		},
	}
	keyUpdateWithTagsModel := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2"},
			},
		},
	}

	keyUpdateWithHostname := request.PublicKeyUpdate{
		Filter: request.PublicKeyFilter{
			Hostname: ".*",
		},
	}
	keyUpdateWithHostnameModel := &models.PublicKey{
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
			},
		},
	}
	keyInvalidUpdateTagsEmpty := request.PublicKeyUpdate{
		Filter: request.PublicKeyFilter{
			Tags: []string{},
		},
	}

	type Expected struct {
		key *models.PublicKey
		err error
	}

	cases := []struct {
		description   string
		fingerprint   string
		tenantID      string
		keyUpdate     request.PublicKeyUpdate
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail update the key when filter tags is empty",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyInvalidUpdateTagsEmpty,
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "tenant").Return([]string{}, 0, err).Once()
			},
			expected: Expected{nil, NewErrTagEmpty("tenant", err)},
		},
		{
			description: "fail to update the key when a tag does not exist in a device",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyUpdateWithTags,
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "tenant").Return([]string{"tag1", "tag4"}, 2, nil).Once()
			},
			expected: Expected{nil, NewErrTagNotFound("tag2", nil)},
		},
		{
			description: "Fail update the key when filter is tags",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyUpdateWithTags,
			requiredMocks: func() {
				model := models.PublicKeyUpdate{
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				mock.On("TagsGet", ctx, "tenant").Return([]string{"tag1", "tag2"}, 2, nil).Once()
				mock.On("PublicKeyUpdate", ctx, "fingerprint", "tenant", &model).Return(nil, err).Once()
			},
			expected: Expected{nil, err},
		},
		{
			description: "Successful update the key when filter is tags",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyUpdateWithTags,
			requiredMocks: func() {
				model := models.PublicKeyUpdate{
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Tags: []string{"tag1", "tag2"},
						},
					},
				}

				mock.On("TagsGet", ctx, "tenant").Return([]string{"tag1", "tag2"}, 2, nil).Once()
				mock.On("PublicKeyUpdate", ctx, "fingerprint", "tenant", &model).Return(keyUpdateWithTagsModel, nil).Once()
			},
			expected: Expected{keyUpdateWithTagsModel, nil},
		},
		{
			description: "Fail update the key when filter is hostname",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyUpdateWithHostname,
			requiredMocks: func() {
				model := models.PublicKeyUpdate{
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}

				mock.On("PublicKeyUpdate", ctx, "fingerprint", "tenant", &model).Return(nil, err).Once()
			},
			expected: Expected{nil, err},
		},
		{
			description: "Successful update the key when filter is tags",
			fingerprint: "fingerprint",
			tenantID:    "tenant",
			keyUpdate:   keyUpdateWithHostname,
			requiredMocks: func() {
				model := models.PublicKeyUpdate{
					PublicKeyFields: models.PublicKeyFields{
						Filter: models.PublicKeyFilter{
							Hostname: ".*",
						},
					},
				}

				mock.On("PublicKeyUpdate", ctx, "fingerprint", "tenant", &model).Return(keyUpdateWithHostnameModel, nil).Once()
			},
			expected: Expected{keyUpdateWithHostnameModel, nil},
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
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now).Twice()

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	ctx := context.TODO()

	namespace := &models.Namespace{TenantID: "tenant1"}

	key := &models.PublicKey{
		Data: []byte("teste"), Fingerprint: "fingerprint", CreatedAt: clock.Now(), TenantID: "tenant1", PublicKeyFields: models.PublicKeyFields{Name: "teste"},
	}

	Err := errors.New("error", "", 0)

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
			fingerprint: key.Fingerprint,
			tenantID:    InvalidTenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, InvalidTenantID).Return(nil, Err).Once()
			},
			expected: Expected{NewErrNamespaceNotFound(InvalidTenantID, Err)},
		},
		{
			description: InvalidFingerprintStr,
			ctx:         ctx,
			fingerprint: InvalidFingerprint,
			tenantID:    key.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, InvalidFingerprint, namespace.TenantID).Return(nil, Err).Once()
			},
			expected: Expected{NewErrPublicKeyNotFound(InvalidFingerprint, Err)},
		},
		{
			description: "fail to delete the key",
			ctx:         ctx,
			fingerprint: key.Fingerprint,
			tenantID:    key.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, key.Fingerprint, namespace.TenantID).Return(key, nil).Once()
				mock.On("PublicKeyDelete", ctx, key.Fingerprint, key.TenantID).Return(Err).Once()
			},
			expected: Expected{Err},
		},
		{
			description: "Successful to delete the key",
			ctx:         ctx,
			fingerprint: key.Fingerprint,
			tenantID:    key.TenantID,
			requiredMocks: func() {
				mock.On("NamespaceGet", ctx, namespace.TenantID).Return(namespace, nil).Once()
				mock.On("PublicKeyGet", ctx, key.Fingerprint, namespace.TenantID).Return(key, nil).Once()
				mock.On("PublicKeyDelete", ctx, key.Fingerprint, key.TenantID).Return(nil).Once()
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
	mock := &mocks.Store{}

	clockMock.On("Now").Return(now)

	s := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock, nil)

	err := errors.New("error", "", 0)

	ctx := context.TODO()

	pubKey, _ := ssh.NewPublicKey(publicKey)
	data := ssh.MarshalAuthorizedKey(pubKey)
	fingerprint := ssh.FingerprintLegacyMD5(pubKey)

	keyInvalidData := request.PublicKeyCreate{
		Data:        nil,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		Filter: request.PublicKeyFilter{
			Hostname: ".*",
		},
	}
	keyInvalidEmptyTags := request.PublicKeyCreate{
		Data:        data,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		Filter: request.PublicKeyFilter{
			Tags: []string{},
		},
	}
	keyInvalidNotFound := request.PublicKeyCreate{
		Data:        data,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		Filter: request.PublicKeyFilter{
			Tags: []string{"tag1", "tag2", "tag4"},
		},
	}
	keyWithTags := request.PublicKeyCreate{
		Data:        data,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		Filter: request.PublicKeyFilter{
			Tags: []string{"tag1", "tag2"},
		},
	}
	keyWithTagsModel := models.PublicKey{
		Data:        data,
		Fingerprint: fingerprint,
		CreatedAt:   clock.Now(),
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Tags: []string{"tag1", "tag2"},
			},
		},
	}
	keyWithHostname := request.PublicKeyCreate{
		Data:        data,
		Fingerprint: fingerprint,
		TenantID:    "tenant",
		Filter: request.PublicKeyFilter{
			Hostname: ".*",
		},
	}
	keyWithHostnameModel := models.PublicKey{
		Data:        data,
		Fingerprint: fingerprint,
		CreatedAt:   clock.Now(),
		TenantID:    "tenant",
		PublicKeyFields: models.PublicKeyFields{
			Filter: models.PublicKeyFilter{
				Hostname: ".*",
			},
		},
	}

	resWithHostnameModel := &response.PublicKeyCreate{
		Data:        keyWithHostnameModel.Data,
		Filter:      response.PublicKeyFilter(keyWithHostnameModel.Filter),
		Name:        keyWithHostnameModel.Name,
		Username:    keyWithHostnameModel.Username,
		TenantID:    keyWithHostnameModel.TenantID,
		Fingerprint: keyWithHostnameModel.Fingerprint,
	}

	resWithTagsModel := &response.PublicKeyCreate{
		Data:        keyWithTagsModel.Data,
		Filter:      response.PublicKeyFilter(keyWithTagsModel.Filter),
		Name:        keyWithTagsModel.Name,
		Username:    keyWithTagsModel.Username,
		TenantID:    keyWithTagsModel.TenantID,
		Fingerprint: keyWithTagsModel.Fingerprint,
	}

	type Expected struct {
		res *response.PublicKeyCreate
		err error
	}

	cases := []struct {
		description   string
		tenantID      string
		req           request.PublicKeyCreate
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fail to create the key when filter tags is empty",
			tenantID:    "tenant",
			req:         keyInvalidEmptyTags,
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "tenant").Return([]string{}, 0, err).Once()
			},
			expected: Expected{nil, NewErrTagEmpty("tenant", err)},
		},
		{
			description: "fail to create the key when a tags does not exist in a device",
			tenantID:    "tenant",
			req:         keyInvalidNotFound,
			requiredMocks: func() {
				mock.On("TagsGet", ctx, "tenant").Return([]string{"tag1", "tag4"}, 2, nil).Once()
			},
			expected: Expected{nil, NewErrTagNotFound("tag2", nil)},
		},
		{
			description: "fail when data in public key is not valid",
			tenantID:    "tenant",
			req:         keyInvalidData,
			requiredMocks: func() {
			},
			expected: Expected{nil, NewErrPublicKeyDataInvalid(keyInvalidData.Data, nil)},
		},
		{
			description: "fail when cannot get the public key",
			tenantID:    "tenant",
			req:         keyWithHostname,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, err).Once()
			},
			expected: Expected{nil, NewErrPublicKeyNotFound(keyWithHostname.Fingerprint, err)},
		},
		{
			description: "fail when public key is duplicated",
			tenantID:    "tenant",
			req:         keyWithHostname,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(&keyWithHostnameModel, nil).Once()
			},
			expected: Expected{nil, NewErrPublicKeyDuplicated([]string{keyWithHostname.Fingerprint}, nil)},
		},
		{
			description: "fail to create a public key when filter is hostname",
			tenantID:    "tenant",
			req:         keyWithHostname,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("PublicKeyCreate", ctx, &keyWithHostnameModel).Return(err).Once()
			},
			expected: Expected{nil, err},
		},
		{
			description: "success to create a public key when filter is hostname",
			tenantID:    "tenant",
			req:         keyWithHostname,
			requiredMocks: func() {
				mock.On("PublicKeyGet", ctx, keyWithHostname.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("PublicKeyCreate", ctx, &keyWithHostnameModel).Return(nil).Once()
			},
			expected: Expected{resWithHostnameModel, nil},
		},
		{
			description: "fail to create a public key when filter is tags",
			tenantID:    "tenant",
			req:         keyWithTags,
			requiredMocks: func() {
				mock.On("TagsGet", ctx, keyWithTags.TenantID).Return([]string{"tag1", "tag2"}, 2, nil).Once()
				mock.On("PublicKeyGet", ctx, keyWithTags.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("PublicKeyCreate", ctx, &keyWithTagsModel).Return(err).Once()
			},
			expected: Expected{nil, err},
		},
		{
			description: "success to create a public key when filter is tags",
			tenantID:    "tenant",
			req:         keyWithTags,
			requiredMocks: func() {
				mock.On("TagsGet", ctx, keyWithTags.TenantID).Return([]string{"tag1", "tag2"}, 2, nil).Once()
				mock.On("PublicKeyGet", ctx, keyWithTags.Fingerprint, "tenant").Return(nil, nil).Once()
				mock.On("PublicKeyCreate", ctx, &keyWithTagsModel).Return(nil).Once()
			},
			expected: Expected{resWithTagsModel, nil},
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
