package routes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shellhub-io/shellhub/api/services/mocks"
	"github.com/shellhub-io/shellhub/pkg/api/authorizer"
	"github.com/shellhub-io/shellhub/pkg/models"
	"github.com/stretchr/testify/assert"
	gomock "github.com/stretchr/testify/mock"
)

func TestGetTags(t *testing.T) {
	mock := new(mocks.Service)

	cases := []struct {
		title          string
		requiredMocks  func()
		expectedStatus int
	}{
		{
			title:          "success when try to get an existing tag",
			expectedStatus: http.StatusOK,
			requiredMocks: func() {
				mock.On("GetTags", gomock.Anything, "").Return([]models.Tags{
					{
						Name:   "tag-1",
						Color:  "#ff0000",
						Tenant: "00000000-0000-4000-0000-000000000000",
					},
					{
						Name:   "tag-2",
						Color:  "green",
						Tenant: "00000000-0000-4000-0000-000000000000",
					},
				}, 2, nil)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.title, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Role", authorizer.RoleOwner.String())
			rec := httptest.NewRecorder()

			e := NewRouter(mock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expectedStatus, rec.Result().StatusCode)
		})
	}

	mock.AssertExpectations(t)
}

func TestRenameTag(t *testing.T) {
	svcMock := new(mocks.Service)

	type Expected struct {
		status int
	}

	cases := []struct {
		description   string
		tag           string
		headers       map[string]string
		body          map[string]interface{}
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			tag:         "tag",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"tag": "newTag",
			},
			requiredMocks: func() {
			},
			expected: Expected{
				status: http.StatusForbidden,
			},
		},
		{
			description: "fails when validate because the tag does not have a min of 3 characters",
			tag:         "ta",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"tag": "newTag",
			},
			requiredMocks: func() {},
			expected: Expected{
				status: http.StatusBadRequest,
			},
		},
		{
			description: "fails when validate because the tag does not have a max of 255 characters",
			tag:         "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"tag": "newTag",
			},
			requiredMocks: func() {},
			expected: Expected{
				status: http.StatusBadRequest,
			},
		},
		{
			description: "fails when validate because have a '/' with in your characters",
			tag:         "tag/",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"tag": "newTag",
			},
			requiredMocks: func() {},
			expected: Expected{
				status: http.StatusBadRequest,
			},
		},
		{
			description: "fails when validate because have a '&' with in your characters",
			tag:         "tag&",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"tag": "newTag",
			},
			requiredMocks: func() {},
			expected: Expected{
				status: http.StatusBadRequest,
			},
		},
		{
			description: "fails when validate because have a '@' with in your characters",
			tag:         "tag@",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"tag": "newTag",
			},
			requiredMocks: func() {},
			expected: Expected{
				status: http.StatusBadRequest,
			},
		},
		{
			description: "success when try to renaming an existing tag",
			tag:         "oldTag",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			body: map[string]interface{}{
				"tag": "newTag",
			},
			requiredMocks: func() {
				svcMock.
					On("RenameTag", gomock.Anything, "00000000-0000-4000-0000-000000000000", "oldTag", "newTag").
					Return(nil).
					Once()
			},
			expected: Expected{
				status: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			jsonData, err := json.Marshal(tc.body)
			if err != nil {
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/tags/%s", tc.tag), strings.NewReader(string(jsonData)))
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.status, rec.Result().StatusCode)
		})
	}

	svcMock.AssertExpectations(t)
}

func TestDeleteTag(t *testing.T) {
	svcMock := new(mocks.Service)

	type Expected struct {
		status int
	}

	cases := []struct {
		description   string
		tag           string
		headers       map[string]string
		requiredMocks func()
		expected      Expected
	}{
		{
			description: "fails when role is observer",
			tag:         "tag",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "observer",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
			},
			expected: Expected{
				status: http.StatusForbidden,
			},
		},
		{
			description: "fails when validate because the tag does not have a min of 3 characters",
			tag:         "ta",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {},
			expected: Expected{
				status: http.StatusBadRequest,
			},
		},
		{
			description: "fails when validate because the tag does not have a max of 255 characters",
			tag:         "BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9BCD3821E12F7A6D89295D86E277F2C365D7A4C3FCCD75D8A2F46C0A556A8EBAAF0845C85D50241FC2F9806D8668FF75D262FDA0A055784AD36D8CA7D2BB600C9",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {},
			expected: Expected{
				status: http.StatusBadRequest,
			},
		},
		{
			description: "fails when validate because have a '/' with in your characters",
			tag:         "tag/",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {},
			expected: Expected{
				status: http.StatusBadRequest,
			},
		},
		{
			description: "fails when validate because have a '&' with in your characters",
			tag:         "tag&",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {},
			expected: Expected{
				status: http.StatusBadRequest,
			},
		},
		{
			description: "fails when validate because have a '@' with in your characters",
			tag:         "tag@",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {},
			expected: Expected{
				status: http.StatusBadRequest,
			},
		},
		{
			description: "success when try to deleting an existing tag",
			tag:         "tag1",
			headers: map[string]string{
				"Content-Type": "application/json",
				"X-Tenant-ID":  "00000000-0000-4000-0000-000000000000",
				"X-Role":       "owner",
				"X-ID":         "000000000000000000000000",
			},
			requiredMocks: func() {
				svcMock.
					On("DeleteTag", gomock.Anything, "00000000-0000-4000-0000-000000000000", "tag1").
					Return(nil).
					Once()
			},
			expected: Expected{
				status: http.StatusOK,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.description, func(t *testing.T) {
			tc.requiredMocks()

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/tags/%s", tc.tag), nil)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			rec := httptest.NewRecorder()

			e := NewRouter(svcMock)
			e.ServeHTTP(rec, req)

			assert.Equal(t, tc.expected.status, rec.Result().StatusCode)
		})
	}

	svcMock.AssertExpectations(t)
}
