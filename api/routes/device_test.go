package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	svc "github.com/shellhub-io/shellhub/api/services"
	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetDevice(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession *models.Device
		expectedStatus  int
	}
	cases := []struct {
		title         string
		uid           string
		requiredMocks func()
		expected      Expected
	}{
		{
			title:         "fails when bind fails to validate uid",
			uid:           "",
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title: "fails when try to get a non-existing device",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("GetDevice", gomock.Anything, models.UID("1234")).Return(nil, svc.ErrDeviceNotFound)
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title: "success when try to get a existing device",
			uid:   "123",
			requiredMocks: func() {
				mock.On("GetDevice", gomock.Anything, models.UID("123")).Return(&models.Device{}, nil)
			},
			expected: Expected{
				expectedSession: &models.Device{},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/devices/%s", tc.uid), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}
}

func TestDeleteDevice(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		uid            string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title:          "fails when bind fails to validate uid",
			uid:            "",
			requiredMocks:  func() {},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "fails when try to deleting a non-existing device",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("DeleteDevice", gomock.Anything, models.UID("1234"), "").Return(svc.ErrDeviceNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to deleting an existing device",
			uid:   "123",
			requiredMocks: func() {
				mock.On("DeleteDevice", gomock.Anything, models.UID("123"), "").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/devices/%s", tc.uid), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestRenameDevice(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		renamePayload  requests.DeviceRename
		tenant         string
		requiredMocks  func(req requests.DeviceRename)
		expectedStatus int
	}{
		{
			title: "fails when bind fails to validate uid",
			renamePayload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: ""},
			},
			tenant:         "tenant-id",
			requiredMocks:  func(_ requests.DeviceRename) {},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "fails when try to rename a non-existing device",
			renamePayload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Name:        "name",
			},
			tenant: "tenant-id",
			requiredMocks: func(req requests.DeviceRename) {
				mock.On("RenameDevice", gomock.Anything, models.UID("1234"), req.Name, "tenant-id").Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to rename an existing device",
			renamePayload: requests.DeviceRename{
				DeviceParam: requests.DeviceParam{UID: "123"},
				Name:        "name",
			},
			tenant: "tenant-id",
			requiredMocks: func(req requests.DeviceRename) {
				mock.On("RenameDevice", gomock.Anything, models.UID("123"), req.Name, "tenant-id").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.renamePayload)

			jsonData, err := json.Marshal(tc.renamePayload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/devices/%s", tc.renamePayload.UID), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-Tenant-ID", tc.tenant)
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestGetDeviceByPublicURLAddress(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession *models.Device
		expectedStatus  int
	}
	cases := []struct {
		title         string
		address       string
		requiredMocks func()
		expected      Expected
	}{
		{
			title:         "fails when bind fails to validate uid",
			address:       "",
			requiredMocks: func() {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title:   "fails when try to searching a device by the public URL address",
			address: "exampleaddress",
			requiredMocks: func() {
				mock.On("GetDeviceByPublicURLAddress", gomock.Anything, "exampleaddress").Return(nil, svc.ErrDeviceNotFound)
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title:   "success when try to searching a device by the public URL address",
			address: "example",
			requiredMocks: func() {
				mock.On("GetDeviceByPublicURLAddress", gomock.Anything, "example").Return(&models.Device{}, nil)
			},
			expected: Expected{
				expectedSession: &models.Device{},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/internal/devices/public/%s", tc.address), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}
}

func TestGetDeviceList(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		devices []models.Device
		status  int
	}

	cases := []struct {
		description   string
		req           *requests.DeviceList
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when try to get a device list existing",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatus("online"),
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "name", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func() {
				mock.
					On("ListDevices", gomock.Anything, gomock.AnythingOfType("*requests.DeviceList")).
					Return(nil, 0, svc.ErrDeviceNotFound).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				status:  http.StatusNotFound,
			},
		},
		{
			description: "fails when try to get a device list existing",
			req: &requests.DeviceList{
				TenantID:     "00000000-0000-4000-0000-000000000000",
				DeviceStatus: models.DeviceStatus("online"),
				Paginator:    query.Paginator{Page: 1, PerPage: 10},
				Sorter:       query.Sorter{By: "name", Order: "asc"},
				Filters:      query.Filters{},
			},
			requiredMocks: func() {
				mock.
					On("ListDevices", gomock.Anything, gomock.AnythingOfType("*requests.DeviceList")).
					Return([]models.Device{}, 0, nil).
					Once()
			},
			expected: Expected{
				devices: []models.Device{},
				status:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			urlVal := &url.Values{}
			urlVal.Set("page", strconv.Itoa(tc.req.Page))
			urlVal.Set("per_page", strconv.Itoa(tc.req.PerPage))
			urlVal.Set("sort_by", tc.req.By)
			urlVal.Set("order_by", tc.req.Order)
			urlVal.Set("status", string(tc.req.DeviceStatus))

			req := httptest.NewRequest(http.MethodGet, "/api/devices?"+urlVal.Encode(), nil)
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-Tenant-ID", tc.req.TenantID)

			rec := httptest.NewRecorder()
			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			devices := make([]models.Device, 0)
			if len(tc.expected.devices) != 0 {
				if err := json.NewDecoder(rec.Result().Body).Decode(&devices); err != nil {
					require.ErrorIs(t, io.EOF, err)
				}
			}

			require.Equal(t, tc.expected.status, rec.Result().StatusCode)
			require.Equal(t, tc.expected.devices, devices)
		})
	}
}

func TestOfflineDevice(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		uid            string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title:          "fails when bind fails to validate uid",
			uid:            "",
			requiredMocks:  func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			title: "fails when try to setting a non-existing device as offline",
			uid:   "1234",
			requiredMocks: func() {
				mock.On("OfflineDevice", gomock.Anything, models.UID("1234")).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to setting an existing device as offline",
			uid:   "123",
			requiredMocks: func() {
				mock.On("OfflineDevice", gomock.Anything, models.UID("123")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/internal/devices/%s/offline", tc.uid), nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-Tenant-ID", "tenant-id")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}

func TestLookupDevice(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession *models.Device
		expectedStatus  int
	}
	tests := []struct {
		title         string
		request       requests.DeviceLookup
		requiredMocks func(requests.DeviceLookup)
		expected      Expected
	}{
		{
			title: "fails when bind fails to validate uid",
			request: requests.DeviceLookup{
				Username:  "user1",
				IPAddress: "192.168.1.100",
			},
			requiredMocks: func(_ requests.DeviceLookup) {},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusBadRequest,
			},
		},
		{
			title: "fails when try to look up of a existing device",
			request: requests.DeviceLookup{
				Domain:    "example.com",
				Name:      "device1",
				Username:  "user1",
				IPAddress: "192.168.1.100",
			},
			requiredMocks: func(req requests.DeviceLookup) {
				mock.On("LookupDevice", gomock.Anything, req.Domain, req.Name).Return(nil, svc.ErrDeviceNotFound).Once()
			},
			expected: Expected{
				expectedSession: nil,
				expectedStatus:  http.StatusNotFound,
			},
		},
		{
			title: "success when try to look up of a existing device",
			request: requests.DeviceLookup{
				Domain:    "example.com",
				Name:      "device1",
				Username:  "user1",
				IPAddress: "192.168.1.100",
			},
			requiredMocks: func(req requests.DeviceLookup) {
				mock.On("LookupDevice", gomock.Anything, req.Domain, req.Name).Return(&models.Device{}, nil)
			},
			expected: Expected{
				expectedSession: &models.Device{},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.request)

			jsonData, err := json.Marshal(tc.request)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, "/internal/lookup", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.Device
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}

			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}
}

func TestUpdateDevice(t *testing.T) {
	mock := new(mocks.Service)
	name := "new device name"
	url := true

	cases := []struct {
		title          string
		updatePayload  requests.DeviceUpdate
		requiredMocks  func(req requests.DeviceUpdate)
		expectedStatus int
	}{
		{
			title: "fails when try to update a existing device",
			updatePayload: requests.DeviceUpdate{
				DeviceParam: requests.DeviceParam{UID: "1234"},
				Name:        &name,
				PublicURL:   &url,
			},
			requiredMocks: func(req requests.DeviceUpdate) {
				mock.On("UpdateDevice", gomock.Anything, "tenant-id", models.UID("1234"), req.Name, req.PublicURL).Return(svc.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to update a existing device",
			updatePayload: requests.DeviceUpdate{
				DeviceParam: requests.DeviceParam{UID: "123"},
				Name:        &name,
				PublicURL:   &url,
			},

			requiredMocks: func(req requests.DeviceUpdate) {
				mock.On("UpdateDevice", gomock.Anything, "tenant-id", models.UID("123"), req.Name, req.PublicURL).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.updatePayload)

			jsonData, err := json.Marshal(tc.updatePayload)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/devices/%s", tc.updatePayload.UID), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			req.Header.Set("X-Tenant-ID", "tenant-id")
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}
