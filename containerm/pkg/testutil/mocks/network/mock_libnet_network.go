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
// Source: network.go

// Package mocks is a generated GoMock package.
package mocks

import (
	libnetwork "github.com/docker/docker/libnetwork"
	networkdb "github.com/docker/docker/libnetwork/networkdb"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockNetwork is a mock of Network interface
type MockNetwork struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkMockRecorder
}

// MockNetworkMockRecorder is the mock recorder for MockNetwork
type MockNetworkMockRecorder struct {
	mock *MockNetwork
}

// NewMockNetwork creates a new mock instance
func NewMockNetwork(ctrl *gomock.Controller) *MockNetwork {
	mock := &MockNetwork{ctrl: ctrl}
	mock.recorder = &MockNetworkMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNetwork) EXPECT() *MockNetworkMockRecorder {
	return m.recorder
}

// Name mocks base method
func (m *MockNetwork) Name() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Name")
	ret0, _ := ret[0].(string)
	return ret0
}

// Name indicates an expected call of Name
func (mr *MockNetworkMockRecorder) Name() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Name", reflect.TypeOf((*MockNetwork)(nil).Name))
}

// ID mocks base method
func (m *MockNetwork) ID() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ID")
	ret0, _ := ret[0].(string)
	return ret0
}

// ID indicates an expected call of ID
func (mr *MockNetworkMockRecorder) ID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ID", reflect.TypeOf((*MockNetwork)(nil).ID))
}

// Type mocks base method
func (m *MockNetwork) Type() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Type")
	ret0, _ := ret[0].(string)
	return ret0
}

// Type indicates an expected call of Type
func (mr *MockNetworkMockRecorder) Type() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Type", reflect.TypeOf((*MockNetwork)(nil).Type))
}

// CreateEndpoint mocks base method
func (m *MockNetwork) CreateEndpoint(name string, options ...libnetwork.EndpointOption) (libnetwork.Endpoint, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{name}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateEndpoint", varargs...)
	ret0, _ := ret[0].(libnetwork.Endpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateEndpoint indicates an expected call of CreateEndpoint
func (mr *MockNetworkMockRecorder) CreateEndpoint(name interface{}, options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{name}, options...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEndpoint", reflect.TypeOf((*MockNetwork)(nil).CreateEndpoint), varargs...)
}

// Delete mocks base method
func (m *MockNetwork) Delete(options ...libnetwork.NetworkDeleteOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range options {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Delete", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete
func (mr *MockNetworkMockRecorder) Delete(options ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockNetwork)(nil).Delete), options...)
}

// Endpoints mocks base method
func (m *MockNetwork) Endpoints() []libnetwork.Endpoint {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Endpoints")
	ret0, _ := ret[0].([]libnetwork.Endpoint)
	return ret0
}

// Endpoints indicates an expected call of Endpoints
func (mr *MockNetworkMockRecorder) Endpoints() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Endpoints", reflect.TypeOf((*MockNetwork)(nil).Endpoints))
}

// WalkEndpoints mocks base method
func (m *MockNetwork) WalkEndpoints(walker libnetwork.EndpointWalker) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WalkEndpoints", walker)
}

// WalkEndpoints indicates an expected call of WalkEndpoints
func (mr *MockNetworkMockRecorder) WalkEndpoints(walker interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WalkEndpoints", reflect.TypeOf((*MockNetwork)(nil).WalkEndpoints), walker)
}

// EndpointByName mocks base method
func (m *MockNetwork) EndpointByName(name string) (libnetwork.Endpoint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EndpointByName", name)
	ret0, _ := ret[0].(libnetwork.Endpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EndpointByName indicates an expected call of EndpointByName
func (mr *MockNetworkMockRecorder) EndpointByName(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EndpointByName", reflect.TypeOf((*MockNetwork)(nil).EndpointByName), name)
}

// EndpointByID mocks base method
func (m *MockNetwork) EndpointByID(id string) (libnetwork.Endpoint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EndpointByID", id)
	ret0, _ := ret[0].(libnetwork.Endpoint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EndpointByID indicates an expected call of EndpointByID
func (mr *MockNetworkMockRecorder) EndpointByID(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EndpointByID", reflect.TypeOf((*MockNetwork)(nil).EndpointByID), id)
}

// Info mocks base method
func (m *MockNetwork) Info() libnetwork.NetworkInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Info")
	ret0, _ := ret[0].(libnetwork.NetworkInfo)
	return ret0
}

// Info indicates an expected call of Info
func (mr *MockNetworkMockRecorder) Info() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockNetwork)(nil).Info))
}

// MockNetworkInfo is a mock of NetworkInfo interface
type MockNetworkInfo struct {
	ctrl     *gomock.Controller
	recorder *MockNetworkInfoMockRecorder
}

// MockNetworkInfoMockRecorder is the mock recorder for MockNetworkInfo
type MockNetworkInfoMockRecorder struct {
	mock *MockNetworkInfo
}

// NewMockNetworkInfo creates a new mock instance
func NewMockNetworkInfo(ctrl *gomock.Controller) *MockNetworkInfo {
	mock := &MockNetworkInfo{ctrl: ctrl}
	mock.recorder = &MockNetworkInfoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockNetworkInfo) EXPECT() *MockNetworkInfoMockRecorder {
	return m.recorder
}

// IpamConfig mocks base method
func (m *MockNetworkInfo) IpamConfig() (string, map[string]string, []*libnetwork.IpamConf, []*libnetwork.IpamConf) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IpamConfig")
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(map[string]string)
	ret2, _ := ret[2].([]*libnetwork.IpamConf)
	ret3, _ := ret[3].([]*libnetwork.IpamConf)
	return ret0, ret1, ret2, ret3
}

// IpamConfig indicates an expected call of IpamConfig
func (mr *MockNetworkInfoMockRecorder) IpamConfig() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IpamConfig", reflect.TypeOf((*MockNetworkInfo)(nil).IpamConfig))
}

// IpamInfo mocks base method
func (m *MockNetworkInfo) IpamInfo() ([]*libnetwork.IpamInfo, []*libnetwork.IpamInfo) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IpamInfo")
	ret0, _ := ret[0].([]*libnetwork.IpamInfo)
	ret1, _ := ret[1].([]*libnetwork.IpamInfo)
	return ret0, ret1
}

// IpamInfo indicates an expected call of IpamInfo
func (mr *MockNetworkInfoMockRecorder) IpamInfo() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IpamInfo", reflect.TypeOf((*MockNetworkInfo)(nil).IpamInfo))
}

// DriverOptions mocks base method
func (m *MockNetworkInfo) DriverOptions() map[string]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DriverOptions")
	ret0, _ := ret[0].(map[string]string)
	return ret0
}

// DriverOptions indicates an expected call of DriverOptions
func (mr *MockNetworkInfoMockRecorder) DriverOptions() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DriverOptions", reflect.TypeOf((*MockNetworkInfo)(nil).DriverOptions))
}

// Scope mocks base method
func (m *MockNetworkInfo) Scope() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Scope")
	ret0, _ := ret[0].(string)
	return ret0
}

// Scope indicates an expected call of Scope
func (mr *MockNetworkInfoMockRecorder) Scope() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Scope", reflect.TypeOf((*MockNetworkInfo)(nil).Scope))
}

// IPv6Enabled mocks base method
func (m *MockNetworkInfo) IPv6Enabled() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IPv6Enabled")
	ret0, _ := ret[0].(bool)
	return ret0
}

// IPv6Enabled indicates an expected call of IPv6Enabled
func (mr *MockNetworkInfoMockRecorder) IPv6Enabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IPv6Enabled", reflect.TypeOf((*MockNetworkInfo)(nil).IPv6Enabled))
}

// Internal mocks base method
func (m *MockNetworkInfo) Internal() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Internal")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Internal indicates an expected call of Internal
func (mr *MockNetworkInfoMockRecorder) Internal() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Internal", reflect.TypeOf((*MockNetworkInfo)(nil).Internal))
}

// Attachable mocks base method
func (m *MockNetworkInfo) Attachable() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Attachable")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Attachable indicates an expected call of Attachable
func (mr *MockNetworkInfoMockRecorder) Attachable() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Attachable", reflect.TypeOf((*MockNetworkInfo)(nil).Attachable))
}

// Ingress mocks base method
func (m *MockNetworkInfo) Ingress() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ingress")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Ingress indicates an expected call of Ingress
func (mr *MockNetworkInfoMockRecorder) Ingress() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ingress", reflect.TypeOf((*MockNetworkInfo)(nil).Ingress))
}

// ConfigFrom mocks base method
func (m *MockNetworkInfo) ConfigFrom() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConfigFrom")
	ret0, _ := ret[0].(string)
	return ret0
}

// ConfigFrom indicates an expected call of ConfigFrom
func (mr *MockNetworkInfoMockRecorder) ConfigFrom() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConfigFrom", reflect.TypeOf((*MockNetworkInfo)(nil).ConfigFrom))
}

// ConfigOnly mocks base method
func (m *MockNetworkInfo) ConfigOnly() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConfigOnly")
	ret0, _ := ret[0].(bool)
	return ret0
}

// ConfigOnly indicates an expected call of ConfigOnly
func (mr *MockNetworkInfoMockRecorder) ConfigOnly() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConfigOnly", reflect.TypeOf((*MockNetworkInfo)(nil).ConfigOnly))
}

// Labels mocks base method
func (m *MockNetworkInfo) Labels() map[string]string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Labels")
	ret0, _ := ret[0].(map[string]string)
	return ret0
}

// Labels indicates an expected call of Labels
func (mr *MockNetworkInfoMockRecorder) Labels() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Labels", reflect.TypeOf((*MockNetworkInfo)(nil).Labels))
}

// Dynamic mocks base method
func (m *MockNetworkInfo) Dynamic() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dynamic")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Dynamic indicates an expected call of Dynamic
func (mr *MockNetworkInfoMockRecorder) Dynamic() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dynamic", reflect.TypeOf((*MockNetworkInfo)(nil).Dynamic))
}

// Created mocks base method
func (m *MockNetworkInfo) Created() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Created")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// Created indicates an expected call of Created
func (mr *MockNetworkInfoMockRecorder) Created() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Created", reflect.TypeOf((*MockNetworkInfo)(nil).Created))
}

// Peers mocks base method
func (m *MockNetworkInfo) Peers() []networkdb.PeerInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Peers")
	ret0, _ := ret[0].([]networkdb.PeerInfo)
	return ret0
}

// Peers indicates an expected call of Peers
func (mr *MockNetworkInfoMockRecorder) Peers() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Peers", reflect.TypeOf((*MockNetworkInfo)(nil).Peers))
}

// Services mocks base method
func (m *MockNetworkInfo) Services() map[string]libnetwork.ServiceInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Services")
	ret0, _ := ret[0].(map[string]libnetwork.ServiceInfo)
	return ret0
}

// Services indicates an expected call of Services
func (mr *MockNetworkInfoMockRecorder) Services() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Services", reflect.TypeOf((*MockNetworkInfo)(nil).Services))
}
