// Code generated by mockery v2.43.0. DO NOT EDIT.

package dbconnector

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	querycoordinator "github.com/ncotds/nco-qoordinator/pkg/querycoordinator"
)

// MockExecutorCloser is an autogenerated mock type for the ExecutorCloser type
type MockExecutorCloser struct {
	mock.Mock
}

type MockExecutorCloser_Expecter struct {
	mock *mock.Mock
}

func (_m *MockExecutorCloser) EXPECT() *MockExecutorCloser_Expecter {
	return &MockExecutorCloser_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *MockExecutorCloser) Close() error {
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

// MockExecutorCloser_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type MockExecutorCloser_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *MockExecutorCloser_Expecter) Close() *MockExecutorCloser_Close_Call {
	return &MockExecutorCloser_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *MockExecutorCloser_Close_Call) Run(run func()) *MockExecutorCloser_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExecutorCloser_Close_Call) Return(_a0 error) *MockExecutorCloser_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockExecutorCloser_Close_Call) RunAndReturn(run func() error) *MockExecutorCloser_Close_Call {
	_c.Call.Return(run)
	return _c
}

// Exec provides a mock function with given fields: ctx, query
func (_m *MockExecutorCloser) Exec(ctx context.Context, query querycoordinator.Query) ([]querycoordinator.QueryResultRow, int, error) {
	ret := _m.Called(ctx, query)

	if len(ret) == 0 {
		panic("no return value specified for Exec")
	}

	var r0 []querycoordinator.QueryResultRow
	var r1 int
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, querycoordinator.Query) ([]querycoordinator.QueryResultRow, int, error)); ok {
		return rf(ctx, query)
	}
	if rf, ok := ret.Get(0).(func(context.Context, querycoordinator.Query) []querycoordinator.QueryResultRow); ok {
		r0 = rf(ctx, query)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]querycoordinator.QueryResultRow)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, querycoordinator.Query) int); ok {
		r1 = rf(ctx, query)
	} else {
		r1 = ret.Get(1).(int)
	}

	if rf, ok := ret.Get(2).(func(context.Context, querycoordinator.Query) error); ok {
		r2 = rf(ctx, query)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// MockExecutorCloser_Exec_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Exec'
type MockExecutorCloser_Exec_Call struct {
	*mock.Call
}

// Exec is a helper method to define mock.On call
//   - ctx context.Context
//   - query querycoordinator.Query
func (_e *MockExecutorCloser_Expecter) Exec(ctx interface{}, query interface{}) *MockExecutorCloser_Exec_Call {
	return &MockExecutorCloser_Exec_Call{Call: _e.mock.On("Exec", ctx, query)}
}

func (_c *MockExecutorCloser_Exec_Call) Run(run func(ctx context.Context, query querycoordinator.Query)) *MockExecutorCloser_Exec_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(querycoordinator.Query))
	})
	return _c
}

func (_c *MockExecutorCloser_Exec_Call) Return(rows []querycoordinator.QueryResultRow, affectedRows int, err error) *MockExecutorCloser_Exec_Call {
	_c.Call.Return(rows, affectedRows, err)
	return _c
}

func (_c *MockExecutorCloser_Exec_Call) RunAndReturn(run func(context.Context, querycoordinator.Query) ([]querycoordinator.QueryResultRow, int, error)) *MockExecutorCloser_Exec_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockExecutorCloser creates a new instance of MockExecutorCloser. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockExecutorCloser(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockExecutorCloser {
	mock := &MockExecutorCloser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
