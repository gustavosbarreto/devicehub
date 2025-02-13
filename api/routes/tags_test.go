package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	servicemock "github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/query"
	"github.com/shellhub-io/shellhub/pkg/api/requests"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_CreateTag(t *testing.T) {
	type Expected struct {
		status int
		header string
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		tenant        string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			tenant:      "00000000-0000-4000-0000-000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"name": "production",
			},
			requiredMocks: func() {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "returns conflict on duplicate tags",
			tenant:      "00000000-0000-4000-0000-000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "owner",
			},
			body: map[string]interface{}{
				"name": "production",
			},
			requiredMocks: func() {
				svcMock.
					On("CreateTag", mock.Anything, &requests.CreateTag{TenantID: "00000000-0000-4000-0000-000000000000", Name: "production"}).
					Return("", []string{"production"}, nil).
					Once()
			},
			expected: Expected{status: http.StatusConflict},
		},
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "owner",
			},
			body: map[string]interface{}{
				"name": "production",
			},
			requiredMocks: func() {
				svcMock.
					On("CreateTag", mock.Anything, &requests.CreateTag{TenantID: "00000000-0000-4000-0000-000000000000", Name: "production"}).
					Return("507f1f77bcf86cd799439011", []string{}, nil).
					Once()
			},
			expected: Expected{
				status: http.StatusCreated,
				header: "507f1f77bcf86cd799439011",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tc.requiredMocks()

			data, err := json.Marshal(tc.body)
			require.NoError(tt, err)

			req := httptest.NewRequest(http.MethodPost, "/api/namespaces/"+tc.tenant+"/tags", strings.NewReader(string(data)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(tt, tc.expected.status, rec.Result().StatusCode)
			if tc.expected.header != "" {
				require.Equal(tt, tc.expected.header, rec.Header().Get("X-Inserted-ID"))
			}
		})
	}
}

func TestHandler_ListTags(t *testing.T) {
	type Expected struct {
		body   []models.Tag
		status int
		count  string
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		tenant        string
		headers       map[string]string
		query         func() string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			requiredMocks: func() {
				svcMock.
					On("ListTags", mock.Anything, &requests.ListTags{TenantID: "00000000-0000-4000-0000-000000000000", Paginator: query.Paginator{Page: 1, PerPage: 10}}).
					Return([]models.Tag{{Name: "production"}}, 1, nil).
					Once()
			},
			expected: Expected{
				body:   []models.Tag{{Name: "production"}},
				status: http.StatusOK,
				count:  "1",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, "/api/namespaces/"+tc.tenant+"/tags?page=1&per_page=10", nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(tt, tc.expected.status, rec.Result().StatusCode)
			if tc.expected.body != nil {
				var responseBody []models.Tag
				require.NoError(tt, json.NewDecoder(rec.Body).Decode(&responseBody))
				require.Equal(tt, tc.expected.body, responseBody)
				require.Equal(tt, tc.expected.count, rec.Header().Get("X-Total-Count"))
			}
		})
	}
}

func TestHandler_UpdateTag(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		tenant        string
		name          string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			tenant:      "00000000-0000-4000-0000-000000000000",
			name:        "production",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "observer",
			},
			body: map[string]interface{}{
				"name": "development",
			},
			requiredMocks: func() {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "returns conflict on duplicate names",
			tenant:      "00000000-0000-4000-0000-000000000000",
			name:        "production",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "owner",
			},
			body: map[string]interface{}{
				"name": "development",
			},
			requiredMocks: func() {
				svcMock.
					On("UpdateTag", mock.Anything, &requests.UpdateTag{TenantID: "00000000-0000-4000-0000-000000000000", Name: "production", NewName: "development"}).
					Return([]string{"development"}, nil).
					Once()
			},
			expected: Expected{status: http.StatusConflict},
		},
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			name:        "production",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Role":       "owner",
			},
			body: map[string]interface{}{
				"name": "development",
			},
			requiredMocks: func() {
				svcMock.
					On("UpdateTag", mock.Anything, &requests.UpdateTag{TenantID: "00000000-0000-4000-0000-000000000000", Name: "production", NewName: "development"}).
					Return([]string{}, nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tc.requiredMocks()

			data, err := json.Marshal(tc.body)
			require.NoError(tt, err)

			req := httptest.NewRequest(http.MethodPatch, "/api/namespaces/"+tc.tenant+"/tags/"+tc.name, strings.NewReader(string(data)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(tt, tc.expected.status, rec.Result().StatusCode)
		})
	}
}

func TestHandler_DeleteTag(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		tenant        string
		name          string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			tenant:      "00000000-0000-4000-0000-000000000000",
			name:        "production",
			headers: map[string]string{
				"X-Role": "observer",
			},
			requiredMocks: func() {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			name:        "production",
			headers: map[string]string{
				"X-Role": "owner",
			},
			requiredMocks: func() {
				svcMock.
					On("DeleteTag", mock.Anything, &requests.DeleteTag{TenantID: "00000000-0000-4000-0000-000000000000", Name: "production"}).
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusNoContent},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, "/api/namespaces/"+tc.tenant+"/tags/"+tc.name, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(tt, tc.expected.status, rec.Result().StatusCode)
		})
	}
}

func TestHandler_PushTagToDevice(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		tenant        string
		deviceUID     string
		tagName       string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			tenant:      "00000000-0000-4000-0000-000000000000",
			deviceUID:   "abc123",
			tagName:     "production",
			headers: map[string]string{
				"X-Role": "observer",
			},
			requiredMocks: func() {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			deviceUID:   "abc123",
			tagName:     "production",
			headers: map[string]string{
				"X-Role": "owner",
			},
			requiredMocks: func() {
				svcMock.
					On("PushTagTo", mock.Anything, models.TagTargetDevice, &requests.PushTag{TenantID: "00000000-0000-4000-0000-000000000000", Name: "production", TargetID: "abc123"}).
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, "/api/namespaces/"+tc.tenant+"/devices/"+tc.deviceUID+"/tags/"+tc.tagName, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(tt, tc.expected.status, rec.Result().StatusCode)
		})
	}
}

func TestHandler_PullTagFromDevice(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		tenant        string
		deviceUID     string
		tagName       string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			tenant:      "00000000-0000-4000-0000-000000000000",
			deviceUID:   "abc123",
			tagName:     "production",
			headers: map[string]string{
				"X-Role": "observer",
			},
			requiredMocks: func() {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			deviceUID:   "abc123",
			tagName:     "production",
			headers: map[string]string{
				"X-Role": "owner",
			},
			requiredMocks: func() {
				svcMock.
					On("PullTagFrom", mock.Anything, models.TagTargetDevice, &requests.PullTag{TenantID: "00000000-0000-4000-0000-000000000000", Name: "production", TargetID: "abc123"}).
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusNoContent},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, "/api/namespaces/"+tc.tenant+"/devices/"+tc.deviceUID+"/tags/"+tc.tagName, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(tt, tc.expected.status, rec.Result().StatusCode)
		})
	}
}

func TestHandler_PushTagToPublicKey(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		tenant        string
		fingerprint   string
		tagName       string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fingerprint: "00:00:00:00:00:00",
			tagName:     "production",
			headers: map[string]string{
				"X-Role": "observer",
			},
			requiredMocks: func() {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fingerprint: "00:00:00:00:00:00",
			tagName:     "production",
			headers: map[string]string{
				"X-Role": "owner",
			},
			requiredMocks: func() {
				svcMock.
					On("PushTagTo", mock.Anything, models.TagTargetPublicKey, &requests.PushTag{TenantID: "00000000-0000-4000-0000-000000000000", Name: "production", TargetID: "00:00:00:00:00:00"}).
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusOK},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodPost, "/api/namespaces/"+tc.tenant+"/sshkeys/public-keys/"+tc.fingerprint+"/tags/"+tc.tagName, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(tt, tc.expected.status, rec.Result().StatusCode)
		})
	}
}

func TestHandler_PullTagFromPublicKey(t *testing.T) {
	type Expected struct {
		status int
	}

	svcMock := new(servicemock.Service)

	cases := []struct {
		description   string
		tenant        string
		fingerprint   string
		tagName       string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fingerprint: "00:00:00:00:00:00",
			tagName:     "production",
			headers: map[string]string{
				"X-Role": "observer",
			},
			requiredMocks: func() {},
			expected:      Expected{status: http.StatusForbidden},
		},
		{
			description: "succeeds",
			tenant:      "00000000-0000-4000-0000-000000000000",
			fingerprint: "00:00:00:00:00:00",
			tagName:     "production",
			headers: map[string]string{
				"X-Role": "owner",
			},
			requiredMocks: func() {
				svcMock.
					On("PullTagFrom", mock.Anything, models.TagTargetPublicKey, &requests.PullTag{TenantID: "00000000-0000-4000-0000-000000000000", Name: "production", TargetID: "00:00:00:00:00:00"}).
					Return(nil).
					Once()
			},
			expected: Expected{status: http.StatusNoContent},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(tt *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, "/api/namespaces/"+tc.tenant+"/sshkeys/public-keys/"+tc.fingerprint+"/tags/"+tc.tagName, nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()
			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			require.Equal(tt, tc.expected.status, rec.Result().StatusCode)
		})
	}
}
