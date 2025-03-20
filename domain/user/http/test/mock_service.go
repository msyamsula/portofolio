// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/msyamsula/portofolio/domain/user/http (interfaces: Service)

// Package test is a generated GoMock package.
package test

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	repository "github.com/msyamsula/portofolio/domain/user/repository"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// GetUser mocks base method.
func (m *MockService) GetUser(arg0 context.Context, arg1 string) (repository.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUser", arg0, arg1)
	ret0, _ := ret[0].(repository.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUser indicates an expected call of GetUser.
func (mr *MockServiceMockRecorder) GetUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUser", reflect.TypeOf((*MockService)(nil).GetUser), arg0, arg1)
}

// SetUser mocks base method.
func (m *MockService) SetUser(arg0 context.Context, arg1 repository.User) (repository.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetUser", arg0, arg1)
	ret0, _ := ret[0].(repository.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SetUser indicates an expected call of SetUser.
func (mr *MockServiceMockRecorder) SetUser(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetUser", reflect.TypeOf((*MockService)(nil).SetUser), arg0, arg1)
}
