// Code generated by mockery v1.0.0. DO NOT EDIT.

// mock

package parser

import mock "github.com/stretchr/testify/mock"

// MockTestExec is an autogenerated mock type for the TestExec type
type MockTestExec struct {
	mock.Mock
}

// Exec provides a mock function with given fields: ctx, flags, options
func (_m *MockTestExec) Exec(ctx Context, flags FlagValues, options Options) (Value, error) {
	ret := _m.Called(ctx, flags, options)

	var r0 Value
	if rf, ok := ret.Get(0).(func(Context, FlagValues, Options) Value); ok {
		r0 = rf(ctx, flags, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Value)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(Context, FlagValues, Options) error); ok {
		r1 = rf(ctx, flags, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}