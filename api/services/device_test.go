package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/shellhub-io/shellhub/api/store"
	storemock "github.com/shellhub-io/shellhub/api/store/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	storecache "github.com/shellhub-io/shellhub/pkg/cache"
	"github.com/shellhub-io/shellhub/pkg/errors"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListDevices(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		description   string
		req           *requests.DeviceList
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails to list devices",
			req: &requests.DeviceList{
				TenantID:     "",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted, mock.AnythingOfType("store.DeviceQueryOption")).
					Return([]models.Device{}, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "", 0),
			},
		},
		{
			description: "succeeds to list devices",
			req: &requests.DeviceList{
				TenantID:     "",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted, mock.AnythingOfType("store.DeviceQueryOption")).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			devices, count, err := service.ListDevices(ctx, tc.req)
			require.Equal(tt, tc.expected, Expected{devices, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestListDevices_status_removed(t *testing.T) {
	storeMock := new(storemock.Store)

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		description   string
		req           *requests.DeviceList
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when could not list the removed devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusRemoved,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceRemovedList", ctx, "00000000-0000-4000-0000-000000000000", query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}).
					Return([]models.DeviceRemoved{}, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				devices: nil,
				count:   0,
				err:     errors.New("error", "", 0),
			},
		},
		{
			description: "succeeds to list the removed devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusRemoved,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceRemovedList", ctx, "00000000-0000-4000-0000-000000000000", query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}).
					Return([]models.DeviceRemoved{{Device: &models.Device{Name: "dev"}, Timestamp: time.Now()}}, 1, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{
					{
						Name: "dev",
					},
				},
				count: 1,
				err:   nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			devices, count, err := service.ListDevices(ctx, tc.req)
			require.Equal(tt, tc.expected, Expected{devices, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestListDevices_tenant_not_empty(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	type Expected struct {
		devices []models.Device
		count   int
		err     error
	}

	cases := []struct {
		description   string
		req           *requests.DeviceList
		requiredMocks func(context.Context)
		expected      Expected
	}{
		{
			description: "fails when the namespace does not exists",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				devices: nil,
				count:   0,
				err:     NewErrNamespaceNotFound("00000000-0000-4000-0000-000000000000", errors.New("error", "", 0)),
			},
		},
		{
			description: "[cloud] fails when could not count the removed devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesCount: 2}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(0), errors.New("error", "layer", 0)).
					Once()
			},
			expected: Expected{
				devices: nil,
				count:   0,
				err:     NewErrDeviceRemovedCount(errors.New("error", "layer", 0)),
			},
		},
		{
			description: "[cloud] fails when the namespace reached the device limit and cannot list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesCount: 2}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(1), nil).
					Once()
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableFromRemoved, mock.AnythingOfType("store.DeviceQueryOption")).
					Return([]models.Device{}, 0, errors.New("error", "layer", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "layer", 0),
			},
		},
		{
			description: "[cloud] succeeds when the namespace reached the device limit and list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesCount: 2}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(1), nil).
					Once()
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableFromRemoved, mock.AnythingOfType("store.DeviceQueryOption")).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     nil,
			},
		},
		{
			description: "[cloud] fails when the namespace do not reached the device limit and cannot list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesCount: 2}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(0), nil).
					Once()
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted, mock.AnythingOfType("store.DeviceQueryOption")).
					Return([]models.Device{}, 0, errors.New("error", "layer", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "layer", 0),
			},
		},
		{
			description: "[cloud] succeeds when the namespace do not reached the device limit and list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesCount: 2}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-4000-0000-000000000000").
					Return(int64(0), nil).
					Once()
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted, mock.AnythingOfType("store.DeviceQueryOption")).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     nil,
			},
		},
		{
			description: "[enterprise|community] fails when the namespace reached the device limit and cannot list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesCount: 3}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.DeviceQueryOption")).
					Return([]models.Device{}, 0, errors.New("error", "layer", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "layer", 0),
			},
		},
		{
			description: "[enterprise|community] succeeds when the namespace reached the device limit and list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesCount: 3}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableAsFalse, mock.AnythingOfType("store.DeviceQueryOption")).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     nil,
			},
		},
		{
			description: "[enterprise|community] fails when the namespace do not reached the device limit and cannot list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesCount: 2}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted, mock.AnythingOfType("store.DeviceQueryOption")).
					Return([]models.Device{}, 0, errors.New("error", "layer", 0)).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     errors.New("error", "layer", 0),
			},
		},
		{
			description: "[enterprise|community] succeeds when the namespace do not reached the device limit and list the devices",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatusAccepted,
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "created_at", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func(ctx context.Context) {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-4000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(&models.Namespace{TenantID: "00000000-0000-4000-0000-000000000000", MaxDevices: 3, DevicesCount: 2}, nil).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("true").
					Once()
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceList", ctx, models.DeviceStatusAccepted, query.Paginator{Page: 1, PerPage: 10}, query.Filters{}, query.Sorter{By: "created_at", Order: "asc"}, store.DeviceAcceptableIfNotAccepted, mock.AnythingOfType("store.DeviceQueryOption")).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				count:   0,
				err:     nil,
			},
		},
	}

	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			ctx := context.TODO()
			tc.requiredMocks(ctx)

			devices, count, err := service.ListDevices(ctx, tc.req)
			require.Equal(tt, tc.expected, Expected{devices, count, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestGetDevice(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	type Expected struct {
		device *models.Device
		err    error
	}

	cases := []struct {
		description   string
		requiredMocks func()
		uid           models.UID
		expected      Expected
	}{
		{
			description: "fails when the store get device fails",
			requiredMocks: func() {
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.On("DeviceGet", ctx, models.UID("_uid"), mock.AnythingOfType("store.DeviceQueryOption")).
					Return(nil, errors.New("error", "", 0)).Once()
			},
			uid: models.UID("_uid"),
			expected: Expected{
				nil,
				NewErrDeviceNotFound(models.UID("_uid"), errors.New("error", "", 0)),
			},
		},
		{
			description: "succeeds",
			requiredMocks: func() {
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()

				device := &models.Device{UID: "uid"}
				storeMock.On("DeviceGet", ctx, models.UID("uid"), mock.AnythingOfType("store.DeviceQueryOption")).
					Return(device, nil).Once()
			},
			uid: models.UID("uid"),
			expected: Expected{
				&models.Device{UID: "uid"},
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

			returnedDevice, err := service.GetDevice(ctx, tc.uid)
			assert.Equal(t, tc.expected, Expected{returnedDevice, err})
		})
	}

	storeMock.AssertExpectations(t)
}

func TestDeleteDevice(t *testing.T) {
	storeMock := new(storemock.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		requiredMocks func()
		uid           models.UID
		tenant        string
		expected      error
	}{
		{
			description: "fails when the store device get by uid fails",
			uid:         models.UID("_uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("_uid"), "tenant").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("_uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when the store namespace get fails",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(
						&models.Device{
							UID:       "uid",
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "tenant").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrNamespaceNotFound("tenant", errors.New("error", "", 0)),
		},
		{
			description: "fails when the store device delete fails",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(
						&models.Device{
							UID:       "uid",
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "tenant").
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "id",
							TenantID: "tenant",
							Members: []models.Member{
								{
									ID:   "id",
									Role: authorizer.RoleOwner,
								},
								{
									ID:   "id2",
									Role: authorizer.RoleObserver,
								},
							},
							MaxDevices: 3,
						},
						nil,
					).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").Return("false").
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("uid")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "succeeds",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(
						&models.Device{
							UID:       "uid",
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "tenant").
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "id",
							TenantID: "tenant",
							Members: []models.Member{
								{
									ID:   "id",
									Role: authorizer.RoleOwner,
								},
								{
									ID:   "id2",
									Role: authorizer.RoleObserver,
								},
							},
							MaxDevices: 3,
						},
						nil,
					).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").Return("false").
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("uid")).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "[with_billing] fails when device removed insert return error",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(
						&models.Device{
							UID:       "uid",
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "tenant").
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "id",
							TenantID: "tenant",
							Members: []models.Member{
								{
									ID:   "id",
									Role: authorizer.RoleOwner,
								},
								{
									ID:   "id2",
									Role: authorizer.RoleObserver,
								},
							},
							MaxDevices: 3,
						},
						nil,
					).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_BILLING").Return("true").
					Once()
				storeMock.
					On(
						"DeviceRemovedInsert",
						ctx,
						"tenant",
						&models.Device{
							UID:       "uid",
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
					).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceRemovedInsert(errors.New("error", "", 0)),
		},
		{
			description: "[with_billing] succeeds",
			uid:         models.UID("uid"),
			tenant:      "tenant",
			requiredMocks: func() {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").
					Return(
						&models.Device{
							UID:       "uid",
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("NamespaceGet", ctx, "tenant").
					Return(
						&models.Namespace{
							Name:     "group1",
							Owner:    "id",
							TenantID: "tenant",
							Members: []models.Member{
								{
									ID:   "id",
									Role: authorizer.RoleOwner,
								},
								{
									ID:   "id2",
									Role: authorizer.RoleObserver,
								},
							},
							MaxDevices: 3,
						},
						nil,
					).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").Return("true").
					Once()
				envMock.
					On("Get", "SHELLHUB_BILLING").Return("true").
					Once()
				storeMock.
					On(
						"DeviceRemovedInsert",
						ctx,
						"tenant",
						&models.Device{
							UID:       "uid",
							TenantID:  "tenant",
							CreatedAt: time.Time{},
						},
					).
					Return(nil).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("uid")).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.DeleteDevice(ctx, tc.uid, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestRenameDevice(t *testing.T) {
	mock := new(storemock.Store)

	ctx := context.TODO()

	cases := []struct {
		description   string
		requiredMocks func(device *models.Device)
		uid           models.UID
		device        *models.Device
		deviceNewName string
		tenant        string
		expected      error
	}{
		{
			description: "fails when store device get fails",
			tenant:      "tenant",
			uid:         models.UID("uid"),
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description:   "returns nil if the name is the same",
			tenant:        "tenant",
			deviceNewName: "name",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, nil).Once()
			},
			expected: nil,
		},
		{
			description:   "fails when store get by device name fails",
			tenant:        "tenant",
			deviceNewName: "newname",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				device2 := &models.Device{
					UID:      "uid2",
					Name:     "newname",
					TenantID: "tenant2",
				}

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "newname", "tenant", models.DeviceStatusAccepted).Return(device2, errors.New("error", "", 0)).Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description:   "fails when the name already exists",
			tenant:        "tenant",
			deviceNewName: "newname",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				device2 := &models.Device{
					UID:      "uid2",
					Name:     "newname",
					TenantID: "tenant2",
				}

				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "newname", "tenant", models.DeviceStatusAccepted).Return(device2, nil).Once()
			},
			expected: NewErrDeviceDuplicated("newname", nil),
		},
		{
			description:   "fails when the store device rename fails",
			tenant:        "tenant",
			deviceNewName: "anewname",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "anewname", "tenant", models.DeviceStatusAccepted).Return(nil, store.ErrNoDocuments).Once()
				mock.On("DeviceRename", ctx, models.UID("uid"), "anewname").Return(errors.New("error", "", 0)).Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description:   "succeeds",
			tenant:        "tenant",
			deviceNewName: "anewname",
			uid:           models.UID("uid"),
			device:        &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device) {
				mock.On("DeviceGetByUID", ctx, models.UID("uid"), "tenant").Return(device, nil).Once()
				mock.On("DeviceGetByName", ctx, "anewname", "tenant", models.DeviceStatusAccepted).Return(nil, store.ErrNoDocuments).Once()
				mock.On("DeviceRename", ctx, models.UID("uid"), "anewname").Return(nil).Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.device)

			service := NewService(store.Store(mock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.RenameDevice(ctx, tc.uid, tc.deviceNewName, tc.tenant)
			assert.Equal(t, tc.expected, err)
		})
	}

	mock.AssertExpectations(t)
}

func TestLookupDevice(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	type Expected struct {
		device *models.Device
		err    error
	}

	cases := []struct {
		description   string
		namespace     string
		device        *models.Device
		requiredMocks func(device *models.Device, namespace string)
		expected      Expected
	}{
		{
			description: "fails when store device lookup fails",
			namespace:   "namespace",
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device, namespace string) {
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceLookup", ctx, namespace, device.Name, mock.AnythingOfType("store.DeviceQueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: Expected{
				nil,
				NewErrDeviceLookupNotFound("namespace", "name", errors.New("error", "", 0)),
			},
		},
		{
			description: "fails when the device is not found",
			namespace:   "namespace",
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device, namespace string) {
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceLookup", ctx, namespace, device.Name, mock.AnythingOfType("store.DeviceQueryOption")).
					Return(nil, store.ErrNoDocuments).
					Once()
			},
			expected: Expected{
				nil,
				NewErrDeviceLookupNotFound("namespace", "name", store.ErrNoDocuments),
			},
		},
		{
			description: "succeeds",
			namespace:   "namespace",
			device:      &models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
			requiredMocks: func(device *models.Device, namespace string) {
				queryOptionsMock.On("DeviceWithTagDetails").Return(nil).Once()
				storeMock.
					On("DeviceLookup", ctx, namespace, device.Name, mock.AnythingOfType("store.DeviceQueryOption")).
					Return(device, nil).
					Once()
			},
			expected: Expected{
				&models.Device{UID: "uid", Name: "name", TenantID: "tenant", Identity: &models.DeviceIdentity{MAC: "00:00:00:00:00:00"}, Status: "accepted"},
				nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.device, tc.namespace)

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			returnedDevice, err := service.LookupDevice(ctx, tc.namespace, tc.device.Name)
			assert.Equal(t, tc.expected, Expected{returnedDevice, err})
		})
	}
	storeMock.AssertExpectations(t)
}

func TestOfflineDevice(t *testing.T) {
	storeMock := new(storemock.Store)

	cases := []struct {
		name     string
		uid      models.UID
		mocks    func(context.Context)
		expected error
	}{
		{
			name: "fails when operation does not succeeds",
			uid:  models.UID("uid"),
			mocks: func(ctx context.Context) {
				storeMock.
					On("DeviceSetOffline", ctx, "uid").
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			name: "fails when connected_device does not exist",
			uid:  models.UID("uid"),
			mocks: func(ctx context.Context) {
				storeMock.
					On("DeviceSetOffline", ctx, "uid").
					Return(store.ErrNoDocuments).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), store.ErrNoDocuments),
		},
		{
			name: "succeeds",
			uid:  models.UID("uid"),
			mocks: func(ctx context.Context) {
				storeMock.
					On("DeviceSetOffline", ctx, "uid").
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	s := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			tc.mocks(ctx)
			assert.Equal(t, tc.expected, s.OfflineDevice(ctx, tc.uid))
		})
	}

	storeMock.AssertExpectations(t)
}

func TestUpdateDeviceStatus_same_mac(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		status        models.DeviceStatus
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not get the namespace",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
		},
		{
			description: "fails when could not get the device",
			uid:         models.UID("uid"),
			tenant:      "00000000-0000-0000-0000-000000000000",
			status:      "accepted",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "accepted",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when could not get the device by MAC",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when already exist a device with same name and a different mac",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(&models.Device{
						UID:      "uid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(
						&models.Device{
							UID:      "notsameuid",
							Name:     "name",
							Identity: &models.DeviceIdentity{MAC: "anothermac"},
						}, nil,
					).
					Once()
			},
			expected: NewErrDeviceDuplicated("name", nil),
		},
		{
			description: "fails to update device UID",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(
						&models.Device{
							UID:      "notsameuid",
							Name:     "name",
							Identity: &models.DeviceIdentity{MAC: "mac"},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(
						&models.Device{
							UID:      "notsameuid",
							Name:     "name",
							Identity: &models.DeviceIdentity{MAC: "mac"},
						},
						nil,
					).
					Once()
				storeMock.
					On("SessionUpdateDeviceUID", ctx, models.UID("notsameuid"), models.UID("uid")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "fails to update device to the old name",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(
						&models.Device{
							UID:      "notsameuid",
							Name:     "name",
							Identity: &models.DeviceIdentity{MAC: "mac"},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(
						&models.Device{
							UID:      "notsameuid",
							Name:     "name",
							Identity: &models.DeviceIdentity{MAC: "mac"},
						},
						nil,
					).
					Once()
				storeMock.
					On("SessionUpdateDeviceUID", ctx, models.UID("notsameuid"), models.UID("uid")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceRename", ctx, models.UID("uid"), "name").
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "fails to delete device with the same mac",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(
						&models.Device{
							UID:      "notsameuid",
							Name:     "name",
							Identity: &models.DeviceIdentity{MAC: "mac"},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(
						&models.Device{
							UID:      "notsameuid",
							Name:     "name",
							Identity: &models.DeviceIdentity{MAC: "mac"},
						},
						nil,
					).
					Once()
				storeMock.
					On("SessionUpdateDeviceUID", ctx, models.UID("notsameuid"), models.UID("uid")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceRename", ctx, models.UID("uid"), "name").
					Return(nil).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("notsameuid")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "fails to update device status",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(
						&models.Device{
							UID:      "notsameuid",
							Name:     "name",
							Identity: &models.DeviceIdentity{MAC: "mac"},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(
						&models.Device{
							UID:      "notsameuid",
							Name:     "name",
							Identity: &models.DeviceIdentity{MAC: "mac"},
						},
						nil,
					).
					Once()
				storeMock.
					On("SessionUpdateDeviceUID", ctx, models.UID("notsameuid"), models.UID("uid")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceRename", ctx, models.UID("uid"), "name").
					Return(nil).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("notsameuid")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(
						&models.Device{
							UID:      "notsameuid",
							Name:     "name",
							Identity: &models.DeviceIdentity{MAC: "mac"},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(&models.Device{
						UID:      "notsameuid",
						Name:     "name",
						Identity: &models.DeviceIdentity{MAC: "mac"},
					},
						nil,
					).
					Once()
				storeMock.
					On("SessionUpdateDeviceUID", ctx, models.UID("notsameuid"), models.UID("uid")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceRename", ctx, models.UID("uid"), "name").
					Return(nil).
					Once()
				storeMock.
					On("DeviceDelete", ctx, models.UID("notsameuid")).
					Return(nil).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.UpdateDeviceStatus(ctx, tc.tenant, tc.uid, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestUpdateDeviceStatus_community_and_enterprise(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		status        models.DeviceStatus
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not get the namespace",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
		},
		{
			description: "fails when could not get the device",
			uid:         models.UID("uid"),
			tenant:      "00000000-0000-0000-0000-000000000000",
			status:      "accepted",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "accepted",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when could not get the device by MAC",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when already exist a device with same name",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(
						&models.Device{
							UID:  "fb2de504e98d3ccab342b53d83395cd7fda297c71e8da550c31478bae0dbb8c5",
							Name: "name",
						},
						nil,
					).
					Once()
			},
			expected: NewErrDeviceDuplicated("name", nil),
		},
		{
			description: "fails namespace has reached the limit of devices in enterprise instance",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID:     "00000000-0000-0000-0000-000000000000",
							MaxDevices:   3,
							DevicesCount: 3,
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
			},
			expected: NewErrDeviceMaxDevicesReached(3),
		},
		{
			description: "fails when could not update device status on database",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").
					Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("false").Once()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.UpdateDeviceStatus(ctx, tc.tenant, tc.uid, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestUpdateDeviceStatus_cloud_subscription_active(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		status        models.DeviceStatus
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not get the namespace",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
		},
		{
			description: "fails when could not get the device",
			uid:         models.UID("uid"),
			tenant:      "00000000-0000-0000-0000-000000000000",
			status:      "accepted",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "accepted",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when could not get the device by MAC",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when namespace has a subscription active and could not report the device accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
							Billing: &models.Billing{
								Active: true,
							},
						},
						nil,
					).
					Once()

				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").Once()
				clientMock.
					On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").
					Return(0, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrBillingReportNamespaceDelete(errors.New("error", "", 0)),
		},
		{
			description: "fails when namespace has a subscription active and report block the action",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
							Billing: &models.Billing{
								Active: true,
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				clientMock.
					On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").
					Return(402, nil).
					Once()
			},
			expected: NewErrBillingReportNamespaceDelete(ErrPaymentRequired),
		},
		{
			description: "fails when could not update device status on database",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
							Billing: &models.Billing{
								Active: true,
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				clientMock.
					On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").
					Return(200, nil).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
							Billing: &models.Billing{
								Active: true,
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				clientMock.
					On("BillingReport", "00000000-0000-0000-0000-000000000000", "device_accept").
					Return(200, nil).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.UpdateDeviceStatus(ctx, tc.tenant, tc.uid, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestUpdateDeviceStatus_cloud_subscription_inactive(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		status        models.DeviceStatus
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not get the namespace",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
		},
		{
			description: "fails when could not get the device",
			uid:         models.UID("uid"),
			tenant:      "00000000-0000-0000-0000-000000000000",
			status:      "accepted",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "accepted",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when could not get the device by MAC",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("uid"), errors.New("error", "", 0)),
		},
		{
			description: "fails when could not check if device was removed recently",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceRemovedGet(errors.New("error", "", 0)),
		},
		{
			description: "fails when could not count how many devices were removed recently",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(0), errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceRemovedCount(errors.New("error", "", 0)),
		},
		{
			description: "fails when namespace has reached the limit counting with removed devices",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID:     "00000000-0000-0000-0000-000000000000",
							MaxDevices:   3,
							DevicesCount: 1,
							Billing: &models.Billing{
								Active: false,
							},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(2), nil).
					Once()
			},
			expected: NewErrDeviceRemovedFull(3, nil),
		},
		{
			description: "fails when could not evaluate the namespace capabilities when accepted device is not removed",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(1), nil).
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrBillingEvaluate(ErrEvaluate),
		},
		{
			description: "fails when namespace cannot accept more devices",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(1), nil).
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(
						&models.BillingEvaluation{
							CanAccept: false,
						},
						0,
						nil,
					).
					Once()
			},
			expected: ErrDeviceLimit,
		},
		{
			description: "fails to update the device status when device is not on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(1), nil).
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(
						&models.BillingEvaluation{
							CanAccept: true,
						},
						0,
						nil,
					).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update the device status when device is not on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil, nil).
					Once()
				storeMock.
					On("DeviceRemovedCount", ctx, "00000000-0000-0000-0000-000000000000").
					Return(int64(1), nil).
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(
						&models.BillingEvaluation{
							CanAccept: true,
						},
						0,
						nil,
					).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "fail when could not remove the device from removed device list when device is on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(&models.DeviceRemoved{}, nil).
					Once()
				storeMock.
					On("DeviceRemovedDelete", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceRemovedDelete(errors.New("error", "", 0)),
		},
		{
			description: "fail when could not evaluate the namespace when device is on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(&models.DeviceRemoved{}, nil).Once()
				storeMock.
					On("DeviceRemovedDelete", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil).
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(nil, 0, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrBillingEvaluate(ErrEvaluate),
		},
		{
			description: "fails when namespace evaluation block device acceptance when device is on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(&models.DeviceRemoved{}, nil).
					Once()
				storeMock.
					On("DeviceRemovedDelete", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil).
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(
						&models.BillingEvaluation{
							CanAccept: false,
						},
						0,
						nil,
					).
					Once()
			},
			expected: ErrDeviceLimit,
		},
		{
			description: "fails to update device status when device is on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(&models.DeviceRemoved{}, nil).
					Once()
				storeMock.
					On("DeviceRemovedDelete", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil).
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(
						&models.BillingEvaluation{
							CanAccept: true,
						},
						0,
						nil,
					).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status when device is on removed list",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByMac", ctx, "mac", "00000000-0000-0000-0000-000000000000", models.DeviceStatus("accepted")).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "name", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				envMock.
					On("Get", "SHELLHUB_CLOUD").
					Return("true").
					Twice()
				envMock.
					On("Get", "SHELLHUB_ENTERPRISE").
					Return("false").
					Once()
				storeMock.
					On("DeviceRemovedGet", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(&models.DeviceRemoved{}, nil).
					Once()
				storeMock.
					On("DeviceRemovedDelete", ctx, "00000000-0000-0000-0000-000000000000", models.UID("uid")).
					Return(nil).
					Once()
				clientMock.
					On("BillingEvaluate", "00000000-0000-0000-0000-000000000000").
					Return(
						&models.BillingEvaluation{
							CanAccept: true,
						},
						0,
						nil,
					).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("accepted")).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			service := NewService(store.Store(storeMock), privateKey, publicKey, storecache.NewNullCache(), clientMock)
			err := service.UpdateDeviceStatus(ctx, tc.tenant, tc.uid, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}

func TestDeviceUpdate(t *testing.T) {
	storeMock := new(storemock.Store)

	toPointer := func(s string) *string {
		return &s
	}

	other := toPointer("other")

	cases := []struct {
		description   string
		uid           string
		tenant        string
		name          *string
		publicKey     *bool
		requiredMocks func(ctx context.Context)
		expected      error
	}{
		{
			description: "fails when could not get the device by UID",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        nil,
			publicKey:   nil,
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), errors.New("error", "", 0)),
		},
		{
			description: "success when updating the device name to same name",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        toPointer("name"),
			publicKey:   nil,
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:  "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
							Name: "name",
						},
						nil,
					).
					Once()
			},
			expected: nil,
		},
		{
			description: "fails when name does not meet the validation requirements",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        toPointer(""),
			publicKey:   nil,
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:  "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
							Name: "name",
						},
						nil,
					).
					Once()
			},
			expected: NewErrDeviceInvalid(map[string]interface{}{"name": ""}, nil),
		},
		{
			description: "fails when could not get the device by name",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        toPointer("same"),
			publicKey:   nil,
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:  "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
							Name: "name",
						},
						nil,
					).Once()
				storeMock.
					On("DeviceGetByName", ctx, "same", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound(models.UID("same"), fmt.Errorf("failed to get device by name: %w", errors.New("error", "", 0))),
		},
		{
			description: "fails when already exists a device with same name",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        toPointer("same"),
			publicKey:   nil,
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:  "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
							Name: "name",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "same", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(
						&models.Device{
							UID:  "fb2de504e98d3ccab342b53d83395cd7fda297c71e8da550c31478bae0dbb8c5",
							Name: "same",
						},
						nil,
					).
					Once()
			},
			expected: NewErrDeviceDuplicated("same", nil),
		},
		{
			description: "success when update device for a different name",
			uid:         "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
			tenant:      "00000000-0000-0000-0000-000000000000",
			name:        other,
			publicKey:   new(bool),
			requiredMocks: func(ctx context.Context) {
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:  "d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e",
							Name: "name",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByName", ctx, "other", "00000000-0000-0000-0000-000000000000", models.DeviceStatusAccepted).
					Return(nil, store.ErrNoDocuments).
					Once()
				storeMock.
					On("DeviceUpdate", ctx, "00000000-0000-0000-0000-000000000000", models.UID("d6c6a5e97217bbe4467eae46ab004695a766c5c43f70b95efd4b6a4d32b33c6e"), other, new(bool)).
					Return(nil).
					Once()
			},
			expected: nil,
		},
	}

	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, test := range cases {
		t.Run(test.description, func(t *testing.T) {
			ctx := context.Background()
			test.requiredMocks(ctx)

			err := service.UpdateDevice(ctx, test.tenant, models.UID(test.uid), test.name, test.publicKey)
			assert.Equal(t, test.expected, err)
		})
	}
}

func TestUpdateDeviceStatus_other_than_accepted(t *testing.T) {
	storeMock := new(storemock.Store)
	queryOptionsMock := new(storemock.QueryOptions)
	storeMock.On("Options").Return(queryOptionsMock)

	ctx := context.TODO()

	cases := []struct {
		description   string
		uid           models.UID
		status        models.DeviceStatus
		tenant        string
		requiredMocks func()
		expected      error
	}{
		{
			description: "fails when could not get the namespace",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrNamespaceNotFound("00000000-0000-0000-0000-000000000000", errors.New("error", "", 0)),
		},
		{
			description: "fails when could not get the device",
			uid:         models.UID("uid"),
			tenant:      "00000000-0000-0000-0000-000000000000",
			status:      "accepted",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(nil, errors.New("error", "", 0)).
					Once()
			},
			expected: NewErrDeviceNotFound("uid", errors.New("error", "", 0)),
		},
		{
			description: "fails when device already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "accepted",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when the intended status is pending, but store update fails",
			uid:         models.UID("uid"),
			status:      "pending",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("pending")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status when the intended status is pending",
			uid:         models.UID("uid"),
			status:      "pending",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("pending")).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "fails when the intended status is rejected, but store update fails",
			uid:         models.UID("uid"),
			status:      "rejected",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("rejected")).
					Return(errors.New("error", "", 0)).
					Once()
			},
			expected: errors.New("error", "", 0),
		},
		{
			description: "success to update device status when the intended status is rejected",
			uid:         models.UID("uid"),
			status:      "rejected",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
				storeMock.
					On("DeviceUpdateStatus", ctx, models.UID("uid"), models.DeviceStatus("rejected")).
					Return(nil).
					Once()
			},
			expected: nil,
		},
		{
			description: "fails when the device is already accepted",
			uid:         models.UID("uid"),
			status:      "accepted",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "accepted",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						},
						nil,
					).
					Once()
			},
			expected: NewErrDeviceStatusAccepted(nil),
		},
		{
			description: "fails when the intended status is removed",
			uid:         models.UID("uid"),
			status:      "removed",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
			},
			expected: NewErrDeviceStatusInvalid("removed", nil),
		},
		{
			description: "fails when the intended status is unused",
			uid:         models.UID("uid"),
			status:      "unused",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
			},
			expected: NewErrDeviceStatusInvalid("unused", nil),
		},
		{
			description: "fails when the intended status is unknown",
			uid:         models.UID("uid"),
			status:      "unused",
			tenant:      "00000000-0000-0000-0000-000000000000",
			requiredMocks: func() {
				queryOptionsMock.On("CountAcceptedDevices").Return(nil).Once()
				storeMock.
					On("NamespaceGet", ctx, "00000000-0000-0000-0000-000000000000", mock.AnythingOfType("store.NamespaceQueryOption")).
					Return(
						&models.Namespace{
							TenantID: "00000000-0000-0000-0000-000000000000",
						},
						nil,
					).
					Once()
				storeMock.
					On("DeviceGetByUID", ctx, models.UID("uid"), "00000000-0000-0000-0000-000000000000").
					Return(
						&models.Device{
							UID:       "uid",
							Name:      "name",
							TenantID:  "00000000-0000-0000-0000-000000000000",
							Status:    "pending",
							Identity:  &models.DeviceIdentity{MAC: "mac"},
							CreatedAt: time.Time{},
						}, nil,
					).
					Once()
			},
			expected: NewErrDeviceStatusInvalid("unused", nil),
		},
	}

	service := NewService(storeMock, privateKey, publicKey, storecache.NewNullCache(), clientMock)

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			err := service.UpdateDeviceStatus(ctx, tc.tenant, tc.uid, tc.status)
			assert.Equal(t, tc.expected, err)
		})
	}

	storeMock.AssertExpectations(t)
}
