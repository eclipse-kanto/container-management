// Copyright (c) 2021 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0

// Code generated by MockGen. DO NOT EDIT.
// Source: containerm/ctr/ctrd_cio.go

// Package ctr is a generated GoMock package.
package ctr

import (
	reflect "reflect"

	cio "github.com/containerd/containerd/cio"
	logger "github.com/eclipse-kanto/container-management/containerm/logger"
	streams "github.com/eclipse-kanto/container-management/containerm/streams"
	gomock "github.com/golang/mock/gomock"
)

// MockIO is a mock of IO interface.
type MockIO struct {
	ctrl     *gomock.Controller
	recorder *MockIOMockRecorder
}

// MockIOMockRecorder is the mock recorder for MockIO.
type MockIOMockRecorder struct {
	mock *MockIO
}

// NewMockIO creates a new mock instance.
func NewMockIO(ctrl *gomock.Controller) *MockIO {
	mock := &MockIO{ctrl: ctrl}
	mock.recorder = &MockIOMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockIO) EXPECT() *MockIOMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockIO) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockIOMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockIO)(nil).Close))
}

// InitContainerIO mocks base method.
func (m *MockIO) InitContainerIO(dio *cio.DirectIO) (cio.IO, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitContainerIO", dio)
	ret0, _ := ret[0].(cio.IO)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InitContainerIO indicates an expected call of InitContainerIO.
func (mr *MockIOMockRecorder) InitContainerIO(dio interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitContainerIO", reflect.TypeOf((*MockIO)(nil).InitContainerIO), dio)
}

// Reset mocks base method.
func (m *MockIO) Reset() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Reset")
}

// Reset indicates an expected call of Reset.
func (mr *MockIOMockRecorder) Reset() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Reset", reflect.TypeOf((*MockIO)(nil).Reset))
}

// SetLogDriver mocks base method.
func (m *MockIO) SetLogDriver(logDriver logger.LogDriver) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetLogDriver", logDriver)
}

// SetLogDriver indicates an expected call of SetLogDriver.
func (mr *MockIOMockRecorder) SetLogDriver(logDriver interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetLogDriver", reflect.TypeOf((*MockIO)(nil).SetLogDriver), logDriver)
}

// SetMaxBufferSize mocks base method.
func (m *MockIO) SetMaxBufferSize(maxBufferSize int64) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetMaxBufferSize", maxBufferSize)
}

// SetMaxBufferSize indicates an expected call of SetMaxBufferSize.
func (mr *MockIOMockRecorder) SetMaxBufferSize(maxBufferSize interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetMaxBufferSize", reflect.TypeOf((*MockIO)(nil).SetMaxBufferSize), maxBufferSize)
}

// SetNonBlock mocks base method.
func (m *MockIO) SetNonBlock(nonBlock bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetNonBlock", nonBlock)
}

// SetNonBlock indicates an expected call of SetNonBlock.
func (mr *MockIOMockRecorder) SetNonBlock(nonBlock interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetNonBlock", reflect.TypeOf((*MockIO)(nil).SetNonBlock), nonBlock)
}

// Stream mocks base method.
func (m *MockIO) Stream() streams.Stream {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stream")
	ret0, _ := ret[0].(streams.Stream)
	return ret0
}

// Stream indicates an expected call of Stream.
func (mr *MockIOMockRecorder) Stream() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stream", reflect.TypeOf((*MockIO)(nil).Stream))
}

// UseStdin mocks base method.
func (m *MockIO) UseStdin() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UseStdin")
	ret0, _ := ret[0].(bool)
	return ret0
}

// UseStdin indicates an expected call of UseStdin.
func (mr *MockIOMockRecorder) UseStdin() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UseStdin", reflect.TypeOf((*MockIO)(nil).UseStdin))
}

// Wait mocks base method.
func (m *MockIO) Wait() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Wait")
}

// Wait indicates an expected call of Wait.
func (mr *MockIOMockRecorder) Wait() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockIO)(nil).Wait))
}
