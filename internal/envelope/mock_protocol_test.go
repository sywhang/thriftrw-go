// Code generated by MockGen. DO NOT EDIT.
// Source: go.uber.org/thriftrw/protocol (interfaces: Protocol)

package envelope

import (
	gomock "github.com/golang/mock/gomock"
	wire "go.uber.org/thriftrw/wire"
	io "io"
)

// MockProtocol is a mock of Protocol interface
type MockProtocol struct {
	ctrl     *gomock.Controller
	recorder *MockProtocolMockRecorder
}

// MockProtocolMockRecorder is the mock recorder for MockProtocol
type MockProtocolMockRecorder struct {
	mock *MockProtocol
}

// NewMockProtocol creates a new mock instance
func NewMockProtocol(ctrl *gomock.Controller) *MockProtocol {
	mock := &MockProtocol{ctrl: ctrl}
	mock.recorder = &MockProtocolMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (_m *MockProtocol) EXPECT() *MockProtocolMockRecorder {
	return _m.recorder
}

// Decode mocks base method
func (_m *MockProtocol) Decode(_param0 io.ReaderAt, _param1 wire.Type) (wire.Value, error) {
	ret := _m.ctrl.Call(_m, "Decode", _param0, _param1)
	ret0, _ := ret[0].(wire.Value)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Decode indicates an expected call of Decode
func (_mr *MockProtocolMockRecorder) Decode(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Decode", arg0, arg1)
}

// DecodeEnveloped mocks base method
func (_m *MockProtocol) DecodeEnveloped(_param0 io.ReaderAt) (wire.Envelope, error) {
	ret := _m.ctrl.Call(_m, "DecodeEnveloped", _param0)
	ret0, _ := ret[0].(wire.Envelope)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DecodeEnveloped indicates an expected call of DecodeEnveloped
func (_mr *MockProtocolMockRecorder) DecodeEnveloped(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "DecodeEnveloped", arg0)
}

// Encode mocks base method
func (_m *MockProtocol) Encode(_param0 wire.Value, _param1 io.Writer) error {
	ret := _m.ctrl.Call(_m, "Encode", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Encode indicates an expected call of Encode
func (_mr *MockProtocolMockRecorder) Encode(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Encode", arg0, arg1)
}

// EncodeEnveloped mocks base method
func (_m *MockProtocol) EncodeEnveloped(_param0 wire.Envelope, _param1 io.Writer) error {
	ret := _m.ctrl.Call(_m, "EncodeEnveloped", _param0, _param1)
	ret0, _ := ret[0].(error)
	return ret0
}

// EncodeEnveloped indicates an expected call of EncodeEnveloped
func (_mr *MockProtocolMockRecorder) EncodeEnveloped(arg0, arg1 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "EncodeEnveloped", arg0, arg1)
}
