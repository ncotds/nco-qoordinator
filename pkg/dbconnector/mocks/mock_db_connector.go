// Code generated by mockery v2.43.0. DO NOT EDIT.

package dbconnector

import (
	context "context"

	dbconnector "github.com/ncotds/nco-qoordinator/pkg/dbconnector"
	mock "github.com/stretchr/testify/mock"

	querycoordinator "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

// MockDBConnector is an autogenerated mock type for the DBConnector type
type MockDBConnector struct {
	mock.Mock
}

type MockDBConnector_Expecter struct {
	mock *mock.Mock
}

func (_m *MockDBConnector) EXPECT() *MockDBConnector_Expecter {
	return &MockDBConnector_Expecter{mock: &_m.Mock}
}

// Connect provides a mock function with given fields: ctx, addr, credentials
func (_m *MockDBConnector) Connect(ctx context.Context, addr dbconnector.Addr, credentials querycoordinator.Credentials) (dbconnector.ExecutorCloser, error) {
	ret := _m.Called(ctx, addr, credentials)

	if len(ret) == 0 {
		panic("no return value specified for Connect")
	}

	var r0 dbconnector.ExecutorCloser
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, dbconnector.Addr, querycoordinator.Credentials) (dbconnector.ExecutorCloser, error)); ok {
		return rf(ctx, addr, credentials)
	}
	if rf, ok := ret.Get(0).(func(context.Context, dbconnector.Addr, querycoordinator.Credentials) dbconnector.ExecutorCloser); ok {
		r0 = rf(ctx, addr, credentials)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dbconnector.ExecutorCloser)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, dbconnector.Addr, querycoordinator.Credentials) error); ok {
		r1 = rf(ctx, addr, credentials)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockDBConnector_Connect_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Connect'
type MockDBConnector_Connect_Call struct {
	*mock.Call
}

// Connect is a helper method to define mock.On call
//   - ctx context.Context
//   - addr dbconnector.Addr
//   - credentials querycoordinator.Credentials
func (_e *MockDBConnector_Expecter) Connect(ctx interface{}, addr interface{}, credentials interface{}) *MockDBConnector_Connect_Call {
	return &MockDBConnector_Connect_Call{Call: _e.mock.On("Connect", ctx, addr, credentials)}
}

func (_c *MockDBConnector_Connect_Call) Run(run func(ctx context.Context, addr dbconnector.Addr, credentials querycoordinator.Credentials)) *MockDBConnector_Connect_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(dbconnector.Addr), args[2].(querycoordinator.Credentials))
	})
	return _c
}

func (_c *MockDBConnector_Connect_Call) Return(conn dbconnector.ExecutorCloser, err error) *MockDBConnector_Connect_Call {
	_c.Call.Return(conn, err)
	return _c
}

func (_c *MockDBConnector_Connect_Call) RunAndReturn(run func(context.Context, dbconnector.Addr, querycoordinator.Credentials) (dbconnector.ExecutorCloser, error)) *MockDBConnector_Connect_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockDBConnector creates a new instance of MockDBConnector. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockDBConnector(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockDBConnector {
	mock := &MockDBConnector{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
