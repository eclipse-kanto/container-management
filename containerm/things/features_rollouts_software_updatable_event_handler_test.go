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
	"context"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/golang/mock/gomock"
)

func TestSoftwareUpdatableHandleContainerEvents(t *testing.T) {
	const (
		testEventsTimeout = 5 * time.Second
	)
	controller := gomock.NewController(t)

	setupManagerMock(controller)
	setupEventsManagerMock(controller)
	setupThingMock(controller)

	testSu := newSoftwareUpdatable(mockThing, mockContainerManager, mockEventsManager)
	testDependencyDescription := dependencyDescription(testContainer)

	defer func() {
		testSu.dispose()
		controller.Finish()
	}()
	eventChan := make(chan *types.Event)
	errorChan := make(chan error)

	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(eventChan, errorChan)
	testSu.(*softwareUpdatable).handleContainerEvents(context.Background())

	tests := map[string]struct {
		chanEvent                     *types.Event
		mockExecution                 mockContainerEventExpectedUpdateDeps
		initialInstalledDependencies  map[string]*datatypes.DependencyDescription
		expectedInstalledDependencies map[string]*datatypes.DependencyDescription
	}{
		"test_things_container_events_created": {
			chanEvent: &types.Event{
				Type:   types.EventTypeContainers,
				Source: copyTestContainer(testContainer),
				Action: types.EventActionContainersCreated,
			},
			mockExecution: mockContainerEventExpectedNewDep,
			expectedInstalledDependencies: map[string]*datatypes.DependencyDescription{
				generateDependencyDescriptionKey(testDependencyDescription): testDependencyDescription,
			},
		},
		"test_things_container_events_remove": {
			chanEvent: &types.Event{
				Type:   types.EventTypeContainers,
				Source: copyTestContainer(testContainer),
				Action: types.EventActionContainersRemoved,
			},
			mockExecution: mockContainerEventExpectedRemovedDep,
			initialInstalledDependencies: map[string]*datatypes.DependencyDescription{
				generateDependencyDescriptionKey(testDependencyDescription): testDependencyDescription,
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			testWg := &sync.WaitGroup{}

			testSu.(*softwareUpdatable).status.InstalledDependencies = testCase.initialInstalledDependencies

			testCase.mockExecution(t, testCase.expectedInstalledDependencies, testWg)
			eventChan <- testCase.chanEvent

			testutil.AssertWithTimeout(t, testWg, testEventsTimeout)
		})
	}
}

type mockContainerEventExpectedUpdateDeps func(t *testing.T, installedDependencies map[string]*datatypes.DependencyDescription, testWg *sync.WaitGroup)

func mockContainerEventExpectedNewDep(t *testing.T, installedDependencies map[string]*datatypes.DependencyDescription, testWg *sync.WaitGroup) {
	testWg.Add(1)
	mockThing.EXPECT().SetFeatureProperty(SoftwareUpdatableFeatureID, softwareUpdatablePropertyInstalledDependencies, installedDependencies).Do(func(featureId, propertyId, value interface{}) {
		testWg.Done()
	})
}

func mockContainerEventExpectedRemovedDep(t *testing.T, installedDependencies map[string]*datatypes.DependencyDescription, testWg *sync.WaitGroup) {
	testWg.Add(1)
	mockThing.EXPECT().SetFeatureProperty(SoftwareUpdatableFeatureID, softwareUpdatablePropertyInstalledDependencies, installedDependencies).Do(func(featureId, propertyId, value interface{}) {
		testWg.Done()
	})
}
