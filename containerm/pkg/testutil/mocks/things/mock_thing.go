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
// Source: things/api/model/thing.go

// Package mocks is a generated GoMock package.
package mocks

import (
	model "github.com/eclipse-kanto/container-management/things/api/model"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockThing is a mock of Thing interface
type MockThing struct {
	ctrl     *gomock.Controller
	recorder *MockThingMockRecorder
}

// MockThingMockRecorder is the mock recorder for MockThing
type MockThingMockRecorder struct {
	mock *MockThing
}

// NewMockThing creates a new mock instance
func NewMockThing(ctrl *gomock.Controller) *MockThing {
	mock := &MockThing{ctrl: ctrl}
	mock.recorder = &MockThingMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockThing) EXPECT() *MockThingMockRecorder {
	return m.recorder
}

// GetNamespace mocks base method
func (m *MockThing) GetNamespace() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNamespace")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetNamespace indicates an expected call of GetNamespace
func (mr *MockThingMockRecorder) GetNamespace() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNamespace", reflect.TypeOf((*MockThing)(nil).GetNamespace))
}

// GetID mocks base method
func (m *MockThing) GetID() model.NamespacedID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetID")
	ret0, _ := ret[0].(model.NamespacedID)
	return ret0
}

// GetID indicates an expected call of GetID
func (mr *MockThingMockRecorder) GetID() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetID", reflect.TypeOf((*MockThing)(nil).GetID))
}

// GetPolicy mocks base method
func (m *MockThing) GetPolicy() model.NamespacedID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPolicy")
	ret0, _ := ret[0].(model.NamespacedID)
	return ret0
}

// GetPolicy indicates an expected call of GetPolicy
func (mr *MockThingMockRecorder) GetPolicy() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPolicy", reflect.TypeOf((*MockThing)(nil).GetPolicy))
}

// GetDefinition mocks base method
func (m *MockThing) GetDefinition() model.DefinitionID {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDefinition")
	ret0, _ := ret[0].(model.DefinitionID)
	return ret0
}

// GetDefinition indicates an expected call of GetDefinition
func (mr *MockThingMockRecorder) GetDefinition() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDefinition", reflect.TypeOf((*MockThing)(nil).GetDefinition))
}

// SetDefinition mocks base method
func (m *MockThing) SetDefinition(DefinitionID model.DefinitionID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetDefinition", DefinitionID)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetDefinition indicates an expected call of SetDefinition
func (mr *MockThingMockRecorder) SetDefinition(DefinitionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetDefinition", reflect.TypeOf((*MockThing)(nil).SetDefinition), DefinitionID)
}

// RemoveDefinition mocks base method
func (m *MockThing) RemoveDefinition() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveDefinition")
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveDefinition indicates an expected call of RemoveDefinition
func (mr *MockThingMockRecorder) RemoveDefinition() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveDefinition", reflect.TypeOf((*MockThing)(nil).RemoveDefinition))
}

// GetAttributes mocks base method
func (m *MockThing) GetAttributes() map[string]interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAttributes")
	ret0, _ := ret[0].(map[string]interface{})
	return ret0
}

// GetAttributes indicates an expected call of GetAttributes
func (mr *MockThingMockRecorder) GetAttributes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAttributes", reflect.TypeOf((*MockThing)(nil).GetAttributes))
}

// GetAttribute mocks base method
func (m *MockThing) GetAttribute(id string) interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAttribute", id)
	ret0, _ := ret[0].(interface{})
	return ret0
}

// GetAttribute indicates an expected call of GetAttribute
func (mr *MockThingMockRecorder) GetAttribute(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAttribute", reflect.TypeOf((*MockThing)(nil).GetAttribute), id)
}

// SetAttributes mocks base method
func (m *MockThing) SetAttributes(attributes map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetAttributes", attributes)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetAttributes indicates an expected call of SetAttributes
func (mr *MockThingMockRecorder) SetAttributes(attributes interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAttributes", reflect.TypeOf((*MockThing)(nil).SetAttributes), attributes)
}

// SetAttribute mocks base method
func (m *MockThing) SetAttribute(id string, value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetAttribute", id, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetAttribute indicates an expected call of SetAttribute
func (mr *MockThingMockRecorder) SetAttribute(id, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAttribute", reflect.TypeOf((*MockThing)(nil).SetAttribute), id, value)
}

// RemoveAttributes mocks base method
func (m *MockThing) RemoveAttributes() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveAttributes")
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveAttributes indicates an expected call of RemoveAttributes
func (mr *MockThingMockRecorder) RemoveAttributes() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveAttributes", reflect.TypeOf((*MockThing)(nil).RemoveAttributes))
}

// RemoveAttribute mocks base method
func (m *MockThing) RemoveAttribute(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveAttribute", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveAttribute indicates an expected call of RemoveAttribute
func (mr *MockThingMockRecorder) RemoveAttribute(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveAttribute", reflect.TypeOf((*MockThing)(nil).RemoveAttribute), id)
}

// GetFeatures mocks base method
func (m *MockThing) GetFeatures() map[string]model.Feature {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFeatures")
	ret0, _ := ret[0].(map[string]model.Feature)
	return ret0
}

// GetFeatures indicates an expected call of GetFeatures
func (mr *MockThingMockRecorder) GetFeatures() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFeatures", reflect.TypeOf((*MockThing)(nil).GetFeatures))
}

// GetFeature mocks base method
func (m *MockThing) GetFeature(id string) model.Feature {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetFeature", id)
	ret0, _ := ret[0].(model.Feature)
	return ret0
}

// GetFeature indicates an expected call of GetFeature
func (mr *MockThingMockRecorder) GetFeature(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetFeature", reflect.TypeOf((*MockThing)(nil).GetFeature), id)
}

// SetFeatures mocks base method
func (m *MockThing) SetFeatures(features map[string]model.Feature) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetFeatures", features)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetFeatures indicates an expected call of SetFeatures
func (mr *MockThingMockRecorder) SetFeatures(features interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFeatures", reflect.TypeOf((*MockThing)(nil).SetFeatures), features)
}

// SetFeature mocks base method
func (m *MockThing) SetFeature(id string, feature model.Feature) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetFeature", id, feature)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetFeature indicates an expected call of SetFeature
func (mr *MockThingMockRecorder) SetFeature(id, feature interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFeature", reflect.TypeOf((*MockThing)(nil).SetFeature), id, feature)
}

// RemoveFeatures mocks base method
func (m *MockThing) RemoveFeatures() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveFeatures")
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveFeatures indicates an expected call of RemoveFeatures
func (mr *MockThingMockRecorder) RemoveFeatures() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveFeatures", reflect.TypeOf((*MockThing)(nil).RemoveFeatures))
}

// RemoveFeature mocks base method
func (m *MockThing) RemoveFeature(id string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveFeature", id)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveFeature indicates an expected call of RemoveFeature
func (mr *MockThingMockRecorder) RemoveFeature(id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveFeature", reflect.TypeOf((*MockThing)(nil).RemoveFeature), id)
}

// SetFeatureDefinition mocks base method
func (m *MockThing) SetFeatureDefinition(featureId string, DefinitionID []model.DefinitionID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetFeatureDefinition", featureId, DefinitionID)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetFeatureDefinition indicates an expected call of SetFeatureDefinition
func (mr *MockThingMockRecorder) SetFeatureDefinition(featureId, DefinitionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFeatureDefinition", reflect.TypeOf((*MockThing)(nil).SetFeatureDefinition), featureId, DefinitionID)
}

// RemoveFeatureDefinition mocks base method
func (m *MockThing) RemoveFeatureDefinition(featureId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveFeatureDefinition", featureId)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveFeatureDefinition indicates an expected call of RemoveFeatureDefinition
func (mr *MockThingMockRecorder) RemoveFeatureDefinition(featureId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveFeatureDefinition", reflect.TypeOf((*MockThing)(nil).RemoveFeatureDefinition), featureId)
}

// SetFeatureProperties mocks base method
func (m *MockThing) SetFeatureProperties(featureId string, properties map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetFeatureProperties", featureId, properties)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetFeatureProperties indicates an expected call of SetFeatureProperties
func (mr *MockThingMockRecorder) SetFeatureProperties(featureId, properties interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFeatureProperties", reflect.TypeOf((*MockThing)(nil).SetFeatureProperties), featureId, properties)
}

// SetFeatureProperty mocks base method
func (m *MockThing) SetFeatureProperty(featureId, propertyId string, value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetFeatureProperty", featureId, propertyId, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetFeatureProperty indicates an expected call of SetFeatureProperty
func (mr *MockThingMockRecorder) SetFeatureProperty(featureId, propertyId, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetFeatureProperty", reflect.TypeOf((*MockThing)(nil).SetFeatureProperty), featureId, propertyId, value)
}

// RemoveFeatureProperties mocks base method
func (m *MockThing) RemoveFeatureProperties(featureId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveFeatureProperties", featureId)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveFeatureProperties indicates an expected call of RemoveFeatureProperties
func (mr *MockThingMockRecorder) RemoveFeatureProperties(featureId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveFeatureProperties", reflect.TypeOf((*MockThing)(nil).RemoveFeatureProperties), featureId)
}

// RemoveFeatureProperty mocks base method
func (m *MockThing) RemoveFeatureProperty(featureId, propertyId string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveFeatureProperty", featureId, propertyId)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveFeatureProperty indicates an expected call of RemoveFeatureProperty
func (mr *MockThingMockRecorder) RemoveFeatureProperty(featureId, propertyId interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveFeatureProperty", reflect.TypeOf((*MockThing)(nil).RemoveFeatureProperty), featureId, propertyId)
}

// SendMessage mocks base method
func (m *MockThing) SendMessage(action string, value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMessage", action, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMessage indicates an expected call of SendMessage
func (mr *MockThingMockRecorder) SendMessage(action, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMessage", reflect.TypeOf((*MockThing)(nil).SendMessage), action, value)
}

// SendFeatureMessage mocks base method
func (m *MockThing) SendFeatureMessage(featureId, action string, value interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendFeatureMessage", featureId, action, value)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendFeatureMessage indicates an expected call of SendFeatureMessage
func (mr *MockThingMockRecorder) SendFeatureMessage(featureId, action, value interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendFeatureMessage", reflect.TypeOf((*MockThing)(nil).SendFeatureMessage), featureId, action, value)
}

// GetRevision mocks base method
func (m *MockThing) GetRevision() int64 {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRevision")
	ret0, _ := ret[0].(int64)
	return ret0
}

// GetRevision indicates an expected call of GetRevision
func (mr *MockThingMockRecorder) GetRevision() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRevision", reflect.TypeOf((*MockThing)(nil).GetRevision))
}

// GetLastModified mocks base method
func (m *MockThing) GetLastModified() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLastModified")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetLastModified indicates an expected call of GetLastModified
func (mr *MockThingMockRecorder) GetLastModified() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLastModified", reflect.TypeOf((*MockThing)(nil).GetLastModified))
}
