// Code generated by mockery v2.43.1. DO NOT EDIT.

package mocks

import (
	context "context"

	worker "github.com/shellhub-io/shellhub/pkg/worker"
	mock "github.com/stretchr/testify/mock"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Client) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Submit provides a mock function with given fields: ctx, pattern, payload
func (_m *Client) Submit(ctx context.Context, pattern worker.TaskPattern, payload []byte) error {
	ret := _m.Called(ctx, pattern, payload)

	if len(ret) == 0 {
		panic("no return value specified for Submit")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, worker.TaskPattern, []byte) error); ok {
		r0 = rf(ctx, pattern, payload)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SubmitToBatch provides a mock function with given fields: ctx, pattern, payload
func (_m *Client) SubmitToBatch(ctx context.Context, pattern worker.TaskPattern, payload []byte) error {
	ret := _m.Called(ctx, pattern, payload)

	if len(ret) == 0 {
		panic("no return value specified for SubmitToBatch")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, worker.TaskPattern, []byte) error); ok {
		r0 = rf(ctx, pattern, payload)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewClient creates a new instance of Client. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}