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

package things

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/things/client"
	"github.com/golang/mock/gomock"
)

func TestProcessContainerThingDefault(t *testing.T) {
	tests := map[string]struct {
		enabledFeatures []string
		mockExec        mockExecutionProcessThing
	}{
		"test_default_config": {
			enabledFeatures: testThingsFeaturesDefaultSet,
			mockExec:        mockDefault,
		},
		"test_factory_only_config": {
			enabledFeatures: []string{ContainerFactoryFeatureID},
			mockExec:        mockFactoryOnly,
		},
		"test_su_only_config": {
			enabledFeatures: []string{SoftwareUpdatableFeatureID},
			mockExec:        mockSUOnly,
		},
		"test_none_config": {
			enabledFeatures: nil,
			mockExec:        mockNone,
		},
	}
	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			controller := gomock.NewController(t)
			defer controller.Finish()
			setupManagerMock(controller)
			setupEventsManagerMock(controller)
			setupThingMock(controller)
			setupThingsContainerManager(controller)

			namespaceID := client.NewNamespacedID("things.containers.service", "test")
			testThingsMgr.containerThingID = namespaceID.String()
			mockThing.EXPECT().GetID().Times(1).Return(namespaceID)

			testThingsMgr.enabledFeatureIds = testCase.enabledFeatures
			testCase.mockExec(t)
			testThingsMgr.processThing(mockThing)
			if testCase.enabledFeatures != nil {
				testutil.AssertEqual(t, len(testCase.enabledFeatures), len(testThingsMgr.managedFeatures))
				for _, fID := range testCase.enabledFeatures {
					testutil.AssertNotNil(t, testThingsMgr.managedFeatures[fID])
				}
			}
		})
	}
}

type mockExecutionProcessThing func(t *testing.T)

func mockDefault(t *testing.T) {
	mockContainerManager.EXPECT().List(gomock.Any()).Times(2).Return(nil, nil)
	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(2).Return(nil, nil)
	mockThing.EXPECT().SetFeature(ContainerFactoryFeatureID, gomock.Any()).Times(1).Return(nil)
	mockThing.EXPECT().SetFeature(SoftwareUpdatableFeatureID, gomock.Any()).Times(1).Return(nil)
}

func mockFactoryOnly(t *testing.T) {
	mockContainerManager.EXPECT().List(gomock.Any()).Times(1).Return(nil, nil)
	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(nil, nil)
	mockThing.EXPECT().SetFeature(ContainerFactoryFeatureID, gomock.Any()).Times(1).Return(nil)
	mockThing.EXPECT().SetFeature(SoftwareUpdatableFeatureID, gomock.Any()).Times(0)
}

func mockSUOnly(t *testing.T) {
	mockContainerManager.EXPECT().List(gomock.Any()).Times(1).Return(nil, nil)
	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(nil, nil)
	mockThing.EXPECT().SetFeature(ContainerFactoryFeatureID, gomock.Any()).Times(0)
	mockThing.EXPECT().SetFeature(SoftwareUpdatableFeatureID, gomock.Any()).Times(1).Return(nil)
}

func mockNone(t *testing.T) {
	mockContainerManager.EXPECT().List(gomock.Any()).Times(0)
	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(0)
	mockThing.EXPECT().SetFeature(ContainerFactoryFeatureID, gomock.Any()).Times(0)
	mockThing.EXPECT().SetFeature(SoftwareUpdatableFeatureID, gomock.Any()).Times(0)
}
