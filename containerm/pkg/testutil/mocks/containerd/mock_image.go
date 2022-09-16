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
// Source: github.com/containerd/containerd (interfaces: Image)

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	containerd "github.com/containerd/containerd"
	content "github.com/containerd/containerd/content"
	images "github.com/containerd/containerd/images"
	gomock "github.com/golang/mock/gomock"
	digest "github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// MockImage is a mock of Image interface.
type MockImage struct {
	ctrl     *gomock.Controller
	recorder *MockImageMockRecorder
}

// MockImageMockRecorder is the mock recorder for MockImage.
type MockImageMockRecorder struct {
	mock *MockImage
}

// NewMockImage creates a new mock instance.
func NewMockImage(ctrl *gomock.Controller) *MockImage {
	mock := &MockImage{ctrl: ctrl}
	mock.recorder = &MockImageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockImage) EXPECT() *MockImageMockRecorder {
	return m.recorder
}

// Config mocks base method.
func (m *MockImage) Config(arg0 context.Context) (v1.Descriptor, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Config", arg0)
	ret0, _ := ret[0].(v1.Descriptor)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Config indicates an expected call of Config.
func (mr *MockImageMockRecorder) Config(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Config", reflect.TypeOf((*MockImage)(nil).Config), arg0)
}

// ContentStore mocks base method.
func (m *MockImage) ContentStore() content.Store {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ContentStore")
	ret0, _ := ret[0].(content.Store)
	return ret0
}

// ContentStore indicates an expected call of ContentStore.
func (mr *MockImageMockRecorder) ContentStore() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ContentStore", reflect.TypeOf((*MockImage)(nil).ContentStore))
}

// IsUnpacked mocks base method.
func (m *MockImage) IsUnpacked(arg0 context.Context, arg1 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsUnpacked", arg0, arg1)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsUnpacked indicates an expected call of IsUnpacked.
func (mr *MockImageMockRecorder) IsUnpacked(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsUnpacked", reflect.TypeOf((*MockImage)(nil).IsUnpacked), arg0, arg1)
}

// Labels mocks base method.
func (m *MockImage) Labels() map[string]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Labels")
	ret0, _ := ret[0].(map[string]string)
	return ret0
}

// Labels indicates an expected call of Labels.
func (mr *MockImageMockRecorder) Labels() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Labels", reflect.TypeOf((*MockImage)(nil).Labels))
}

// Metadata mocks base method.
func (m *MockImage) Metadata() images.Image {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Metadata")
	ret0, _ := ret[0].(images.Image)
	return ret0
}

// Metadata indicates an expected call of Metadata.
func (mr *MockImageMockRecorder) Metadata() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Metadata", reflect.TypeOf((*MockImage)(nil).Metadata))
}

// Name mocks base method.
func (m *MockImage) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name.
func (mr *MockImageMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockImage)(nil).Name))
}

// RootFS mocks base method.
func (m *MockImage) RootFS(arg0 context.Context) ([]digest.Digest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RootFS", arg0)
	ret0, _ := ret[0].([]digest.Digest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RootFS indicates an expected call of RootFS.
func (mr *MockImageMockRecorder) RootFS(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RootFS", reflect.TypeOf((*MockImage)(nil).RootFS), arg0)
}

// Size mocks base method.
func (m *MockImage) Size(arg0 context.Context) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Size", arg0)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Size indicates an expected call of Size.
func (mr *MockImageMockRecorder) Size(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Size", reflect.TypeOf((*MockImage)(nil).Size), arg0)
}

// Target mocks base method.
func (m *MockImage) Target() v1.Descriptor {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Target")
	ret0, _ := ret[0].(v1.Descriptor)
	return ret0
}

// Target indicates an expected call of Target.
func (mr *MockImageMockRecorder) Target() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Target", reflect.TypeOf((*MockImage)(nil).Target))
}

// Unpack mocks base method.
func (m *MockImage) Unpack(arg0 context.Context, arg1 string, arg2 ...containerd.UnpackOpt) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Unpack", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unpack indicates an expected call of Unpack.
func (mr *MockImageMockRecorder) Unpack(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unpack", reflect.TypeOf((*MockImage)(nil).Unpack), varargs...)
}

// Usage mocks base method.
func (m *MockImage) Usage(arg0 context.Context, arg1 ...containerd.UsageOpt) (int64, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Usage", varargs...)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Usage indicates an expected call of Usage.
func (mr *MockImageMockRecorder) Usage(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Usage", reflect.TypeOf((*MockImage)(nil).Usage), varargs...)
}
