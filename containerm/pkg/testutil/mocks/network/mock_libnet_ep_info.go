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
// Source: endpoint_info.go

// Package mocks is a generated GoMock package.
package mocks

import (
	libnetwork "github.com/docker/libnetwork"
	types "github.com/docker/libnetwork/types"
	gomock "github.com/golang/mock/gomock"
	net "net"
	reflect "reflect"
)

// MockEndpointInfo is a mock of EndpointInfo interface
type MockEndpointInfo struct {
	ctrl     *gomock.Controller
	recorder *MockEndpointInfoMockRecorder
}

// MockEndpointInfoMockRecorder is the mock recorder for MockEndpointInfo
type MockEndpointInfoMockRecorder struct {
	mock *MockEndpointInfo
}

// NewMockEndpointInfo creates a new mock instance
func NewMockEndpointInfo(ctrl *gomock.Controller) *MockEndpointInfo {
	mock := &MockEndpointInfo{ctrl: ctrl}
	mock.recorder = &MockEndpointInfoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockEndpointInfo) EXPECT() *MockEndpointInfoMockRecorder {
	return m.recorder
}

// Iface mocks base method
func (m *MockEndpointInfo) Iface() libnetwork.InterfaceInfo {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Iface")
	ret0, _ := ret[0].(libnetwork.InterfaceInfo)
	return ret0
}

// Iface indicates an expected call of Iface
func (mr *MockEndpointInfoMockRecorder) Iface() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Iface", reflect.TypeOf((*MockEndpointInfo)(nil).Iface))
}

// Gateway mocks base method
func (m *MockEndpointInfo) Gateway() net.IP {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Gateway")
	ret0, _ := ret[0].(net.IP)
	return ret0
}

// Gateway indicates an expected call of Gateway
func (mr *MockEndpointInfoMockRecorder) Gateway() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Gateway", reflect.TypeOf((*MockEndpointInfo)(nil).Gateway))
}

// GatewayIPv6 mocks base method
func (m *MockEndpointInfo) GatewayIPv6() net.IP {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GatewayIPv6")
	ret0, _ := ret[0].(net.IP)
	return ret0
}

// GatewayIPv6 indicates an expected call of GatewayIPv6
func (mr *MockEndpointInfoMockRecorder) GatewayIPv6() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GatewayIPv6", reflect.TypeOf((*MockEndpointInfo)(nil).GatewayIPv6))
}

// StaticRoutes mocks base method
func (m *MockEndpointInfo) StaticRoutes() []*types.StaticRoute {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StaticRoutes")
	ret0, _ := ret[0].([]*types.StaticRoute)
	return ret0
}

// StaticRoutes indicates an expected call of StaticRoutes
func (mr *MockEndpointInfoMockRecorder) StaticRoutes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StaticRoutes", reflect.TypeOf((*MockEndpointInfo)(nil).StaticRoutes))
}

// Sandbox mocks base method
func (m *MockEndpointInfo) Sandbox() libnetwork.Sandbox {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sandbox")
	ret0, _ := ret[0].(libnetwork.Sandbox)
	return ret0
}

// Sandbox indicates an expected call of Sandbox
func (mr *MockEndpointInfoMockRecorder) Sandbox() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sandbox", reflect.TypeOf((*MockEndpointInfo)(nil).Sandbox))
}

// LoadBalancer mocks base method
func (m *MockEndpointInfo) LoadBalancer() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadBalancer")
	ret0, _ := ret[0].(bool)
	return ret0
}

// LoadBalancer indicates an expected call of LoadBalancer
func (mr *MockEndpointInfoMockRecorder) LoadBalancer() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadBalancer", reflect.TypeOf((*MockEndpointInfo)(nil).LoadBalancer))
}

// MockInterfaceInfo is a mock of InterfaceInfo interface
type MockInterfaceInfo struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceInfoMockRecorder
}

// MockInterfaceInfoMockRecorder is the mock recorder for MockInterfaceInfo
type MockInterfaceInfoMockRecorder struct {
	mock *MockInterfaceInfo
}

// NewMockInterfaceInfo creates a new mock instance
func NewMockInterfaceInfo(ctrl *gomock.Controller) *MockInterfaceInfo {
	mock := &MockInterfaceInfo{ctrl: ctrl}
	mock.recorder = &MockInterfaceInfoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInterfaceInfo) EXPECT() *MockInterfaceInfoMockRecorder {
	return m.recorder
}

// MacAddress mocks base method
func (m *MockInterfaceInfo) MacAddress() net.HardwareAddr {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MacAddress")
	ret0, _ := ret[0].(net.HardwareAddr)
	return ret0
}

// MacAddress indicates an expected call of MacAddress
func (mr *MockInterfaceInfoMockRecorder) MacAddress() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MacAddress", reflect.TypeOf((*MockInterfaceInfo)(nil).MacAddress))
}

// Address mocks base method
func (m *MockInterfaceInfo) Address() *net.IPNet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Address")
	ret0, _ := ret[0].(*net.IPNet)
	return ret0
}

// Address indicates an expected call of Address
func (mr *MockInterfaceInfoMockRecorder) Address() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Address", reflect.TypeOf((*MockInterfaceInfo)(nil).Address))
}

// AddressIPv6 mocks base method
func (m *MockInterfaceInfo) AddressIPv6() *net.IPNet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddressIPv6")
	ret0, _ := ret[0].(*net.IPNet)
	return ret0
}

// AddressIPv6 indicates an expected call of AddressIPv6
func (mr *MockInterfaceInfoMockRecorder) AddressIPv6() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddressIPv6", reflect.TypeOf((*MockInterfaceInfo)(nil).AddressIPv6))
}

// LinkLocalAddresses mocks base method
func (m *MockInterfaceInfo) LinkLocalAddresses() []*net.IPNet {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LinkLocalAddresses")
	ret0, _ := ret[0].([]*net.IPNet)
	return ret0
}

// LinkLocalAddresses indicates an expected call of LinkLocalAddresses
func (mr *MockInterfaceInfoMockRecorder) LinkLocalAddresses() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LinkLocalAddresses", reflect.TypeOf((*MockInterfaceInfo)(nil).LinkLocalAddresses))
}

// SrcName mocks base method
func (m *MockInterfaceInfo) SrcName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SrcName")
	ret0, _ := ret[0].(string)
	return ret0
}

// SrcName indicates an expected call of SrcName
func (mr *MockInterfaceInfoMockRecorder) SrcName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SrcName", reflect.TypeOf((*MockInterfaceInfo)(nil).SrcName))
}
