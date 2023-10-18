// Code generated by mockery v2.20.2. DO NOT EDIT.

package mocks

import (
	context "context"
	http "net/http"

	mock "github.com/stretchr/testify/mock"

	net "net"
)

// Tunneler is an autogenerated mock type for the Tunneler type
type Tunneler struct {
	mock.Mock
}

// ConnectionPath provides a mock function with given fields:
func (_m *Tunneler) ConnectionPath() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Dial provides a mock function with given fields: ctx, id
func (_m *Tunneler) Dial(ctx context.Context, id string) (net.Conn, error) {
	ret := _m.Called(ctx, id)

	var r0 net.Conn
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (net.Conn, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) net.Conn); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(net.Conn)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DialerPath provides a mock function with given fields:
func (_m *Tunneler) DialerPath() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ForwardResponse provides a mock function with given fields: resp, w
func (_m *Tunneler) ForwardResponse(resp *http.Response, w http.ResponseWriter) {
	_m.Called(resp, w)
}

// Router provides a mock function with given fields:
func (_m *Tunneler) Router() http.Handler {
	ret := _m.Called()

	var r0 http.Handler
	if rf, ok := ret.Get(0).(func() http.Handler); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(http.Handler)
		}
	}

	return r0
}

// SendRequest provides a mock function with given fields: ctx, id, req
func (_m *Tunneler) SendRequest(ctx context.Context, id string, req *http.Request) (*http.Response, error) {
	ret := _m.Called(ctx, id, req)

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *http.Request) (*http.Response, error)); ok {
		return rf(ctx, id, req)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, *http.Request) *http.Response); ok {
		r0 = rf(ctx, id, req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, *http.Request) error); ok {
		r1 = rf(ctx, id, req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetCloseHandler provides a mock function with given fields: callback
func (_m *Tunneler) SetCloseHandler(callback func(string)) {
	_m.Called(callback)
}

// SetConnectionHandler provides a mock function with given fields: callback
func (_m *Tunneler) SetConnectionHandler(callback func(*http.Request) (string, error)) {
	_m.Called(callback)
}

// SetKeepAliveHandler provides a mock function with given fields: callback
func (_m *Tunneler) SetKeepAliveHandler(callback func(string)) {
	_m.Called(callback)
}

type mockConstructorTestingTNewTunneler interface {
	mock.TestingT
	Cleanup(func())
}

// NewTunneler creates a new instance of Tunneler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTunneler(t mockConstructorTestingTNewTunneler) *Tunneler {
	mock := &Tunneler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
