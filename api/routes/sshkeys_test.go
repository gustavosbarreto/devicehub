package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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
)

func TestGetPublicKeys(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession []models.PublicKey
		expectedStatus  int
	}
	cases := []struct {
		description   string
		paginator     query.Paginator
		requiredMocks func(query query.Paginator)
		expected      Expected
	}{
		{
			description: "success when try to list a publics keys exists",
			paginator: query.Paginator{
				Page:    1,
				PerPage: 10,
			},
			requiredMocks: func(query query.Paginator) {
				mock.On("ListPublicKeys", gomock.Anything, query).Return([]models.PublicKey{}, 1, nil)
			},
			expected: Expected{
				expectedSession: []models.PublicKey{},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks(tc.paginator)

			jsonData, err := json.Marshal(tc.paginator)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, "/api/sshkeys/public-keys", strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session []models.PublicKey
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}
}

func TestGetPublicKey(t *testing.T) {
	mock := new(mocks.Service)

	type Expected struct {
		expectedSession *models.PublicKey
		expectedStatus  int
	}
	cases := []struct {
		title         string
		query         requests.PublicKeyGet
		requiredMocks func(query requests.PublicKeyGet)
		expected      Expected
	}{
		{
			title: "fails when validate because the tag does not have a min of 3 characters",
			query: requests.PublicKeyGet{
				TenantParam: requests.TenantParam{Tenant: "00000000-0000-4000-0000-000000000000"},
			},
			expected:      Expected{expectedStatus: http.StatusBadRequest},
			requiredMocks: func(_ requests.PublicKeyGet) {},
		},
		{
			title: "fails when validate because the tag does not have a max of 255 characters",
			query: requests.PublicKeyGet{
				TenantParam: requests.TenantParam{Tenant: "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9"},
			},
			expected:      Expected{expectedStatus: http.StatusBadRequest},
			requiredMocks: func(_ requests.PublicKeyGet) {},
		},
		{
			title: "fails when validate because have a '/' with in your characters",
			query: requests.PublicKeyGet{
				TenantParam: requests.TenantParam{Tenant: "test/"},
			},
			expected:      Expected{expectedStatus: http.StatusBadRequest},
			requiredMocks: func(_ requests.PublicKeyGet) {},
		},
		{
			title: "fails when validate because have a '&' with in your characters",
			query: requests.PublicKeyGet{
				TenantParam: requests.TenantParam{Tenant: "test&"},
			},
			expected:      Expected{expectedStatus: http.StatusBadRequest},
			requiredMocks: func(_ requests.PublicKeyGet) {},
		},
		{
			title: "fails when validate because have a '@' with in your characters",
			query: requests.PublicKeyGet{
				TenantParam: requests.TenantParam{Tenant: "test@"},
			},
			expected:      Expected{expectedStatus: http.StatusBadRequest},
			requiredMocks: func(_ requests.PublicKeyGet) {},
		},
		{
			title: "success when try to get a public key exists",
			query: requests.PublicKeyGet{
				FingerprintParam: requests.FingerprintParam{Fingerprint: "fingerprint"},
				TenantParam:      requests.TenantParam{Tenant: "00000000-0000-4000-0000-000000000000"},
			},
			requiredMocks: func(query requests.PublicKeyGet) {
				mock.On("GetPublicKey", gomock.Anything, query.Fingerprint, query.Tenant).Return(&models.PublicKey{}, nil)
			},
			expected: Expected{
				expectedSession: &models.PublicKey{},
				expectedStatus:  http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks(tc.query)

			jsonData, err := json.Marshal(tc.query)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/internal/sshkeys/public-keys/%s/%s", tc.query.Fingerprint, tc.query.Tenant), strings.NewReader(string(jsonData)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.expectedStatus, rec.Result().StatusCode)

			var session *models.PublicKey
			if err := json.NewDecoder(rec.Result().Body).Decode(&session); err != nil {
				assert.ErrorIs(t, io.EOF, err)
			}
			assert.Equal(t, tc.expected.expectedSession, session)
		})
	}
}

func TestDeletePublicKey(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(mocks.Service)

	cases := []struct {
		description   string
		fingerprint   string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			fingerprint: "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when role is operator",
			fingerprint: "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "operator",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
			},
			expected: Expected{status: http.StatusForbidden},
		},
		{
			description: "fails when try to deleting a non existent public key",
			fingerprint: "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
				svcMock.
					On("DeletePublicKey", gomock.Anything, "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8", "00000000-0000-4000-0000-000000000000").
					Return(svc.ErrNotFound).
					Once()
			},
			expected: Expected{status: http.StatusNotFound},
		},
		{
			description: "success when fingerprint is encoded",
			fingerprint: "8e%3Ab3%3Ae2%3Ace%3A3c%3A6c%3A27%3Aff%3A51%3Ac9%3A5d%3A77%3Aaf%3A92%3A2f%3Ad8",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
				svcMock.
					On("DeletePublicKey", gomock.Anything, "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8", "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
		{
			description: "success when try to deleting an existing public key",
			fingerprint: "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
				svcMock.
					On("DeletePublicKey", gomock.Anything, "8e:b3:e2:ce:3c:6c:27:ff:51:c9:5d:77:af:92:2f:d8", "00000000-0000-4000-0000-000000000000").
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/sshkeys/public-keys/%s", tc.fingerprint), nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.status, rec.Result().StatusCode)
		})
	}
}

func TestCreatePrivateKey(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title: "fails when try to deleting an existing public key",
			requiredMocks: func() {
				mock.On("CreatePrivateKey", gomock.Anything).Return(nil, svc.ErrNotFound).Once()
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			title: "success when try to creating an existing private key",
			requiredMocks: func() {
				mock.On("CreatePrivateKey", gomock.Anything).Return(&models.PrivateKey{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, "/internal/sshkeys/private-keys", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}
}
