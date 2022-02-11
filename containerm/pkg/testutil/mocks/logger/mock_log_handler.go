// Copyright (c) 2021 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

// Code generated by MockGen. DO NOT EDIT.
// Source: containerm/logger/log_handler.go

// Package mocks is a generated GoMock package.
package mocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockLogHandler is a mock of LogHandler interface
type MockLogHandler struct {
	ctrl     *gomock.Controller
	recorder *MockLogHandlerMockRecorder
}

// MockLogHandlerMockRecorder is the mock recorder for MockLogHandler
type MockLogHandlerMockRecorder struct {
	mock *MockLogHandler
}

// NewMockLogHandler creates a new mock instance
func NewMockLogHandler(ctrl *gomock.Controller) *MockLogHandler {
	mock := &MockLogHandler{ctrl: ctrl}
	mock.recorder = &MockLogHandlerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockLogHandler) EXPECT() *MockLogHandlerMockRecorder {
	return m.recorder
}

// StartCopyToLogDriver mocks base method
func (m *MockLogHandler) StartCopyToLogDriver() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartCopyToLogDriver")
}

// StartCopyToLogDriver indicates an expected call of StartCopyToLogDriver
func (mr *MockLogHandlerMockRecorder) StartCopyToLogDriver() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartCopyToLogDriver", reflect.TypeOf((*MockLogHandler)(nil).StartCopyToLogDriver))
}

// Wait mocks base method
func (m *MockLogHandler) Wait() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Wait")
}

// Wait indicates an expected call of Wait
func (mr *MockLogHandlerMockRecorder) Wait() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockLogHandler)(nil).Wait))
}
