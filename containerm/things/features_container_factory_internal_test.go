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

package things

import (
	"reflect"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/golang/mock/gomock"
)

/* Feature Matcher definition (move to the tests base if needed)
-Using the standart Eq() matcher is not enough, as functions (as struct fields in this case) are not comparable even by DeepEqual. */
type featureEq struct {
	feature model.Feature
}

func (m featureEq) Matches(arg interface{}) bool {
	featureArg := arg.(model.Feature)
	result := true
	result = result && reflect.DeepEqual(featureArg.GetDefinition(), m.feature.GetDefinition())
	result = result && reflect.DeepEqual(featureArg.GetID(), m.feature.GetID())
	result = result && reflect.DeepEqual(featureArg.GetProperties(), m.feature.GetProperties())

	return result
}

// In order ot satisfy the Matcher interface
func (m featureEq) String() string {

	return ""
}

func FeatureEq(feature model.Feature) gomock.Matcher {
	return featureEq{feature: feature}
}

func TestContainerFactoryInternalCreateFeature(t *testing.T) {
	tests := map[string]struct {
		testContainer *types.Container
		mockExec      mockExecContainerFactoryInternal
	}{
		"test_create_feature_no_error": {
			testContainer: testContainer,
			mockExec:      mockExecContainerFactoryInternalCreate,
		},
		"test_create_feature_error": {
			testContainer: testContainer,
			mockExec:      mockExecContainerFactoryInternalCreateErr,
		},
	}
	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)

			setupManagerMock(controller)
			setupEventsManagerMock(controller)
			setupContainerFactoryStorageMock(controller)
			setupThingMock(controller)

			testFactory := newContainerFactoryFeature(mockContainerManager, mockEventsManager, mockThing, mockContainerStorage)
			err := testCase.mockExec(t, testCase.testContainer)
			result := testFactory.(*containerFactoryFeature).createContainerFeature(testCase.testContainer)
			testutil.AssertError(t, err, result)

		})
	}
}

type mockExecContainerFactoryInternal func(t *testing.T, container *types.Container) error

func mockExecContainerFactoryInternalCreate(t *testing.T, container *types.Container) error {
	expectedFeature := newContainerFeature(container.Image.Name, container.Name, container, mockContainerManager)
	dittoFeature := expectedFeature.createFeature()
	mockThing.EXPECT().SetFeature(dittoFeature.GetID(), FeatureEq(dittoFeature)).Return(nil)
	return nil
}
func mockExecContainerFactoryInternalCreateErr(t *testing.T, container *types.Container) error {
	expectedFeature := newContainerFeature(container.Image.Name, container.Name, container, mockContainerManager)
	dittoFeature := expectedFeature.createFeature()
	err := log.NewError("test error")
	mockThing.EXPECT().SetFeature(dittoFeature.GetID(), FeatureEq(dittoFeature)).Return(err)
	return err
}

func TestContainerFactoryInternalRemoveFeature(t *testing.T) {
	tests := map[string]struct {
		testContainer *types.Container
		mockExec      mockExecContainerFactoryInternal
	}{
		"test_remove_feature_no_error": {
			testContainer: testContainer,
			mockExec:      mockExecContainerFactoryInternalRemove,
		},
		"test_remove_feature_error": {
			testContainer: testContainer,
			mockExec:      mockExecContainerFactoryInternalRemoveErr,
		},
	}
	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)

			setupManagerMock(controller)
			setupEventsManagerMock(controller)
			setupContainerFactoryStorageMock(controller)
			setupThingMock(controller)

			testFactory := newContainerFactoryFeature(mockContainerManager, mockEventsManager, mockThing, mockContainerStorage)
			err := testCase.mockExec(t, testCase.testContainer)
			result := testFactory.(*containerFactoryFeature).removeContainerFeature(testCase.testContainer.ID)
			testutil.AssertError(t, err, result)

		})
	}
}

func mockExecContainerFactoryInternalRemove(t *testing.T, container *types.Container) error {
	mockThing.EXPECT().RemoveFeature(generateContainerFeatureID(container.ID)).Return(nil)
	return nil
}
func mockExecContainerFactoryInternalRemoveErr(t *testing.T, container *types.Container) error {
	err := log.NewError("test error")
	mockThing.EXPECT().RemoveFeature(generateContainerFeatureID(container.ID)).Return(err)
	return err
}

func TestContainerFactoryInternalUpdateFeature(t *testing.T) {
	tests := map[string]struct {
		testContainer *types.Container
		mockExec      mockExecContainerFactoryInternal
	}{
		"test_update_feature_no_error": {
			testContainer: testContainer,
			mockExec:      mockExecContainerFactoryInternalUpdate,
		},
		"test_update_feature_error": {
			testContainer: testContainer,
			mockExec:      mockExecContainerFactoryInternalUpdateErr,
		},
	}
	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)

			setupManagerMock(controller)
			setupEventsManagerMock(controller)
			setupContainerFactoryStorageMock(controller)
			setupThingMock(controller)

			testFactory := newContainerFactoryFeature(mockContainerManager, mockEventsManager, mockThing, mockContainerStorage)
			err := testCase.mockExec(t, testCase.testContainer)
			result := testFactory.(*containerFactoryFeature).
				updateContainerFeature(testCase.testContainer.ID, containerFeaturePropertyPathStatusState, fromAPIContainerState(testCase.testContainer.State))
			testutil.AssertError(t, err, result)

		})
	}
}

func mockExecContainerFactoryInternalUpdate(t *testing.T, container *types.Container) error {
	mockThing.EXPECT().SetFeatureProperty(generateContainerFeatureID(container.ID), containerFeaturePropertyStatus+"/state", fromAPIContainerState(container.State)).Return(nil)
	return nil
}
func mockExecContainerFactoryInternalUpdateErr(t *testing.T, container *types.Container) error {
	err := log.NewError("test error")
	mockThing.EXPECT().SetFeatureProperty(generateContainerFeatureID(container.ID), containerFeaturePropertyStatus+"/state", fromAPIContainerState(container.State)).Return(err)
	return err
}

func TestContainerFactoryInternalProcessContainers(t *testing.T) {
	tests := map[string]struct {
		testContainers    []*types.Container
		restoredInfo      map[string]string
		expectedStoreInfo map[string]string
		mockExec          mockExecContainerFactoryInternalProcessCtrs
	}{
		"test_process_ctrs_equal": {
			testContainers:    []*types.Container{testContainer},
			restoredInfo:      map[string]string{testContainer.ID: generateContainerFeatureID(testContainer.ID)},
			expectedStoreInfo: map[string]string{testContainer.ID: generateContainerFeatureID(testContainer.ID)},
			mockExec:          mockExecContainerFactoryInternalProcessCtrsEqual,
		},
		"test_process_ctrs_remove": {
			testContainers:    []*types.Container{},
			restoredInfo:      map[string]string{testContainer.ID: generateContainerFeatureID(testContainer.ID)},
			expectedStoreInfo: nil,
			mockExec:          mockExecContainerFactoryInternalProcessCtrsRemove,
		},
	}
	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)

			setupManagerMock(controller)
			setupEventsManagerMock(controller)
			setupContainerFactoryStorageMock(controller)
			setupThingMock(controller)

			testFactory := newContainerFactoryFeature(mockContainerManager, mockEventsManager, mockThing, mockContainerStorage)
			testCase.mockExec(t, testCase.testContainers, testCase.restoredInfo, testCase.expectedStoreInfo)

			testFactory.(*containerFactoryFeature).processContainers(testCase.testContainers)
		})
	}
}

type mockExecContainerFactoryInternalProcessCtrs func(t *testing.T, ctrs []*types.Container, restoredInfo map[string]string, expectedStoreInfo map[string]string)

func mockExecContainerFactoryInternalProcessCtrsEqual(t *testing.T, ctrs []*types.Container, restoredInfo map[string]string, expectedStoreInfo map[string]string) {
	mockContainerStorage.EXPECT().Restore().Return(restoredInfo, nil)
	for _, ctr := range ctrs {
		mockExecContainerFactoryInternalCreate(t, ctr)
	}
	mockContainerStorage.EXPECT().UpdateContainersInfo(gomock.Any()).Do(func(arg map[string]string) {
		// assert equal
		testutil.AssertEqual(t, expectedStoreInfo, arg)
	}).Times(1)
}

func mockExecContainerFactoryInternalProcessCtrsRemove(t *testing.T, ctrs []*types.Container, restoredInfo map[string]string, expectedStoreInfo map[string]string) {
	mockContainerStorage.EXPECT().Restore().Return(restoredInfo, nil)
	for _, ctrFeatureID := range restoredInfo {
		mockThing.EXPECT().RemoveFeature(ctrFeatureID).Return(nil)
	}
	mockContainerStorage.EXPECT().UpdateContainersInfo(gomock.Any()).Do(func(arg map[string]string) {
		// assert equal
		testutil.AssertEqual(t, expectedStoreInfo, arg)
	}).Times(1)
}
