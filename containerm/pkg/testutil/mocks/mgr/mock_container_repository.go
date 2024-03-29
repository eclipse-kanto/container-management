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
// Source: ./containerm/mgr/container_repository.go

// Package mocks is a generated GoMock package.
package mocks

import (
	types "github.com/eclipse-kanto/container-management/containerm/containers/types"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockcontainerRepository is a mock of containerRepository interface
type MockcontainerRepository struct {
	ctrl     *gomock.Controller
	recorder *MockcontainerRepositoryMockRecorder
}

// MockcontainerRepositoryMockRecorder is the mock recorder for MockcontainerRepository
type MockcontainerRepositoryMockRecorder struct {
	mock *MockcontainerRepository
}

// NewMockcontainerRepository creates a new mock instance
func NewMockcontainerRepository(ctrl *gomock.Controller) *MockcontainerRepository {
	mock := &MockcontainerRepository{ctrl: ctrl}
	mock.recorder = &MockcontainerRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockcontainerRepository) EXPECT() *MockcontainerRepositoryMockRecorder {
	return m.recorder
}

// Save mocks base method
func (m *MockcontainerRepository) Save(container *types.Container) (*types.Container, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", container)
	ret0, _ := ret[0].(*types.Container)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Save indicates an expected call of Save
func (mr *MockcontainerRepositoryMockRecorder) Save(container interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockcontainerRepository)(nil).Save), container)
}

// ReadAll mocks base method
func (m *MockcontainerRepository) ReadAll() ([]*types.Container, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ReadAll")
	ret0, _ := ret[0].([]*types.Container)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ReadAll indicates an expected call of ReadAll
func (mr *MockcontainerRepositoryMockRecorder) ReadAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ReadAll", reflect.TypeOf((*MockcontainerRepository)(nil).ReadAll))
}

// Read mocks base method
func (m *MockcontainerRepository) Read(containerId string) (*types.Container, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Read", containerId)
	ret0, _ := ret[0].(*types.Container)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Read indicates an expected call of Read
func (mr *MockcontainerRepositoryMockRecorder) Read(containerId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Read", reflect.TypeOf((*MockcontainerRepository)(nil).Read), containerId)
}

// Delete mocks base method
func (m *MockcontainerRepository) Delete(containerId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", containerId)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockcontainerRepositoryMockRecorder) Delete(containerId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockcontainerRepository)(nil).Delete), containerId)
}

// Prune mocks base method
func (m *MockcontainerRepository) Prune() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Prune")
	ret0, _ := ret[0].(error)
	return ret0
}

// Prune indicates an expected call of Prune
func (mr *MockcontainerRepositoryMockRecorder) Prune() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Prune", reflect.TypeOf((*MockcontainerRepository)(nil).Prune))
}
