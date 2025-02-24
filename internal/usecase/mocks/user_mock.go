// Code generated by MockGen. DO NOT EDIT.
// Source: shop/internal/usecase (interfaces: UserUseCaseInterface)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"
	models "shop/internal/models"

	gomock "github.com/golang/mock/gomock"
)

// MockUserUseCaseInterface is a mock of UserUseCaseInterface interface.
type MockUserUseCaseInterface struct {
	ctrl     *gomock.Controller
	recorder *MockUserUseCaseInterfaceMockRecorder
}

// MockUserUseCaseInterfaceMockRecorder is the mock recorder for MockUserUseCaseInterface.
type MockUserUseCaseInterfaceMockRecorder struct {
	mock *MockUserUseCaseInterface
}

// NewMockUserUseCaseInterface creates a new mock instance.
func NewMockUserUseCaseInterface(ctrl *gomock.Controller) *MockUserUseCaseInterface {
	mock := &MockUserUseCaseInterface{ctrl: ctrl}
	mock.recorder = &MockUserUseCaseInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserUseCaseInterface) EXPECT() *MockUserUseCaseInterfaceMockRecorder {
	return m.recorder
}

// Auth mocks base method.
func (m *MockUserUseCaseInterface) Auth(arg0 context.Context, arg1, arg2 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Auth", arg0, arg1, arg2)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Auth indicates an expected call of Auth.
func (mr *MockUserUseCaseInterfaceMockRecorder) Auth(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Auth", reflect.TypeOf((*MockUserUseCaseInterface)(nil).Auth), arg0, arg1, arg2)
}

// GenerateJWTToken mocks base method.
func (m *MockUserUseCaseInterface) GenerateJWTToken(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateJWTToken", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateJWTToken indicates an expected call of GenerateJWTToken.
func (mr *MockUserUseCaseInterfaceMockRecorder) GenerateJWTToken(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateJWTToken", reflect.TypeOf((*MockUserUseCaseInterface)(nil).GenerateJWTToken), arg0)
}

// GetUserInfo mocks base method.
func (m *MockUserUseCaseInterface) GetUserInfo(arg0 context.Context, arg1 string) (*models.InfoResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserInfo", arg0, arg1)
	ret0, _ := ret[0].(*models.InfoResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserInfo indicates an expected call of GetUserInfo.
func (mr *MockUserUseCaseInterfaceMockRecorder) GetUserInfo(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserInfo", reflect.TypeOf((*MockUserUseCaseInterface)(nil).GetUserInfo), arg0, arg1)
}

// VerifyJWTToken mocks base method.
func (m *MockUserUseCaseInterface) VerifyJWTToken(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyJWTToken", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// VerifyJWTToken indicates an expected call of VerifyJWTToken.
func (mr *MockUserUseCaseInterfaceMockRecorder) VerifyJWTToken(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyJWTToken", reflect.TypeOf((*MockUserUseCaseInterface)(nil).VerifyJWTToken), arg0)
}
