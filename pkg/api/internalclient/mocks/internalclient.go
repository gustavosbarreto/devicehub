// Code generated by mockery v2.10.4. DO NOT EDIT.

package mocks

import (
	models "github.com/shellhub-io/shellhub/pkg/models"
	mock "github.com/stretchr/testify/mock"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// BillingEvaluate provides a mock function with given fields: tenantID
func (_m *Client) BillingEvaluate(tenantID string) (*models.Namespace, int, error) {
	ret := _m.Called(tenantID)

	var r0 *models.Namespace
	if rf, ok := ret.Get(0).(func(string) *models.Namespace); ok {
		r0 = rf(tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Namespace)
		}
	}

	var r1 int
	if rf, ok := ret.Get(1).(func(string) int); ok {
		r1 = rf(tenantID)
	} else {
		r1 = ret.Get(1).(int)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(tenantID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CreatePrivateKey provides a mock function with given fields:
func (_m *Client) CreatePrivateKey() (*models.PrivateKey, error) {
	ret := _m.Called()

	var r0 *models.PrivateKey
	if rf, ok := ret.Get(0).(func() *models.PrivateKey); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.PrivateKey)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeviceLookup provides a mock function with given fields: lookup
func (_m *Client) DeviceLookup(lookup map[string]string) (*models.Device, []error) {
	ret := _m.Called(lookup)

	var r0 *models.Device
	if rf, ok := ret.Get(0).(func(map[string]string) *models.Device); ok {
		r0 = rf(lookup)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Device)
		}
	}

	var r1 []error
	if rf, ok := ret.Get(1).(func(map[string]string) []error); ok {
		r1 = rf(lookup)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]error)
		}
	}

	return r0, r1
}

// DevicesHeartbeat provides a mock function with given fields: id
func (_m *Client) DevicesHeartbeat(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DevicesOffline provides a mock function with given fields: id
func (_m *Client) DevicesOffline(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EvaluateKey provides a mock function with given fields: fingerprint, dev, username
func (_m *Client) EvaluateKey(fingerprint string, dev *models.Device, username string) (bool, error) {
	ret := _m.Called(fingerprint, dev, username)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, *models.Device, string) bool); ok {
		r0 = rf(fingerprint, dev, username)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *models.Device, string) error); ok {
		r1 = rf(fingerprint, dev, username)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FinishSession provides a mock function with given fields: uid
func (_m *Client) FinishSession(uid string) []error {
	ret := _m.Called(uid)

	var r0 []error
	if rf, ok := ret.Get(0).(func(string) []error); ok {
		r0 = rf(uid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]error)
		}
	}

	return r0
}

// FirewallEvaluate provides a mock function with given fields: lookup
func (_m *Client) FirewallEvaluate(lookup map[string]string) error {
	ret := _m.Called(lookup)

	var r0 error
	if rf, ok := ret.Get(0).(func(map[string]string) error); ok {
		r0 = rf(lookup)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetDevice provides a mock function with given fields: uid
func (_m *Client) GetDevice(uid string) (*models.Device, error) {
	ret := _m.Called(uid)

	var r0 *models.Device
	if rf, ok := ret.Get(0).(func(string) *models.Device); ok {
		r0 = rf(uid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Device)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetPublicKey provides a mock function with given fields: fingerprint, tenant
func (_m *Client) GetPublicKey(fingerprint string, tenant string) (*models.PublicKey, error) {
	ret := _m.Called(fingerprint, tenant)

	var r0 *models.PublicKey
	if rf, ok := ret.Get(0).(func(string, string) *models.PublicKey); ok {
		r0 = rf(fingerprint, tenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.PublicKey)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(fingerprint, tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// KeepAliveSession provides a mock function with given fields: uid
func (_m *Client) KeepAliveSession(uid string) []error {
	ret := _m.Called(uid)

	var r0 []error
	if rf, ok := ret.Get(0).(func(string) []error); ok {
		r0 = rf(uid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]error)
		}
	}

	return r0
}

// ListDevices provides a mock function with given fields:
func (_m *Client) ListDevices() ([]models.Device, error) {
	ret := _m.Called()

	var r0 []models.Device
	if rf, ok := ret.Get(0).(func() []models.Device); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]models.Device)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Lookup provides a mock function with given fields: lookup
func (_m *Client) Lookup(lookup map[string]string) (string, []error) {
	ret := _m.Called(lookup)

	var r0 string
	if rf, ok := ret.Get(0).(func(map[string]string) string); ok {
		r0 = rf(lookup)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 []error
	if rf, ok := ret.Get(1).(func(map[string]string) []error); ok {
		r1 = rf(lookup)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).([]error)
		}
	}

	return r0, r1
}

// LookupDevice provides a mock function with given fields:
func (_m *Client) LookupDevice() {
	_m.Called()
}

// RecordSession provides a mock function with given fields: session, recordURL
func (_m *Client) RecordSession(session *models.SessionRecorded, recordURL string) {
	_m.Called(session, recordURL)
}

// ReportDelete provides a mock function with given fields: ns
func (_m *Client) ReportDelete(ns *models.Namespace) (int, error) {
	ret := _m.Called(ns)

	var r0 int
	if rf, ok := ret.Get(0).(func(*models.Namespace) int); ok {
		r0 = rf(ns)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*models.Namespace) error); ok {
		r1 = rf(ns)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReportUsage provides a mock function with given fields: ur
func (_m *Client) ReportUsage(ur *models.UsageRecord) (int, error) {
	ret := _m.Called(ur)

	var r0 int
	if rf, ok := ret.Get(0).(func(*models.UsageRecord) int); ok {
		r0 = rf(ur)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*models.UsageRecord) error); ok {
		r1 = rf(ur)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SessionAsAuthenticated provides a mock function with given fields: uid
func (_m *Client) SessionAsAuthenticated(uid string) []error {
	ret := _m.Called(uid)

	var r0 []error
	if rf, ok := ret.Get(0).(func(string) []error); ok {
		r0 = rf(uid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]error)
		}
	}

	return r0
}
