// Code generated by MockGen. DO NOT EDIT.
// Source: internal/controller/mystatefulset_controller.go

// Package mock_controller is a generated GoMock package.
package mock_controller

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockMgr is a mock of Mgr interface.
type MockMgr struct {
	ctrl     *gomock.Controller
	recorder *MockMgrMockRecorder
}

// MockMgrMockRecorder is the mock recorder for MockMgr.
type MockMgrMockRecorder struct {
	mock *MockMgr
}

// NewMockMgr creates a new mock instance.
func NewMockMgr(ctrl *gomock.Controller) *MockMgr {
	mock := &MockMgr{ctrl: ctrl}
	mock.recorder = &MockMgrMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMgr) EXPECT() *MockMgrMockRecorder {
	return m.recorder
}

// Sync mocks base method.
func (m *MockMgr) Sync(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Sync indicates an expected call of Sync.
func (mr *MockMgrMockRecorder) Sync(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MockMgr)(nil).Sync), arg0)
}

// Teardown mocks base method.
func (m *MockMgr) Teardown(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Teardown", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Teardown indicates an expected call of Teardown.
func (mr *MockMgrMockRecorder) Teardown(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Teardown", reflect.TypeOf((*MockMgr)(nil).Teardown), arg0)
}
