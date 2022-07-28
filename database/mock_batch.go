// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/ava-labs/avalanchego/database (interfaces: Batch)

// Package database is a generated GoMock package.
package database

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockBatch is a mock of Batch interface.
type MockBatch struct {
	ctrl     *gomock.Controller
	recorder *MockBatchMockRecorder
}

// MockBatchMockRecorder is the mock recorder for MockBatch.
type MockBatchMockRecorder struct {
	mock *MockBatch
}

// NewMockBatch creates a new mock instance.
func NewMockBatch(ctrl *gomock.Controller) *MockBatch {
	mock := &MockBatch{ctrl: ctrl}
	mock.recorder = &MockBatchMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBatch) EXPECT() *MockBatchMockRecorder {
	return m.recorder
}

// Delete mocks base method.
func (m *MockBatch) Delete(arg0 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockBatchMockRecorder) Delete(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockBatch)(nil).Delete), arg0)
}

// Inner mocks base method.
func (m *MockBatch) Inner() Batch {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Inner")
	ret0, _ := ret[0].(Batch)
	return ret0
}

// Inner indicates an expected call of Inner.
func (mr *MockBatchMockRecorder) Inner() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Inner", reflect.TypeOf((*MockBatch)(nil).Inner))
}

// Put mocks base method.
func (m *MockBatch) Put(arg0, arg1 []byte) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Put indicates an expected call of Put.
func (mr *MockBatchMockRecorder) Put(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockBatch)(nil).Put), arg0, arg1)
}

// Replay mocks base method.
func (m *MockBatch) Replay(arg0 KeyValueWriterDeleter) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Replay", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Replay indicates an expected call of Replay.
func (mr *MockBatchMockRecorder) Replay(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Replay", reflect.TypeOf((*MockBatch)(nil).Replay), arg0)
}

// Reset mocks base method.
func (m *MockBatch) Reset() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Reset")
}

// Reset indicates an expected call of Reset.
func (mr *MockBatchMockRecorder) Reset() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reset", reflect.TypeOf((*MockBatch)(nil).Reset))
}

// Size mocks base method.
func (m *MockBatch) Size() int {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size")
	ret0, _ := ret[0].(int)
	return ret0
}

// Size indicates an expected call of Size.
func (mr *MockBatchMockRecorder) Size() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockBatch)(nil).Size))
}

// Write mocks base method.
func (m *MockBatch) Write() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Write")
	ret0, _ := ret[0].(error)
	return ret0
}

// Write indicates an expected call of Write.
func (mr *MockBatchMockRecorder) Write() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Write", reflect.TypeOf((*MockBatch)(nil).Write))
}
