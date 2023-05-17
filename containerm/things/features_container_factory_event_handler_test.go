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
	"github.com/golang/mock/gomock"
)

func TestContainerFactoryHandleContainerEvents(t *testing.T) {
	const (
		testEventsTimeout = 5 * time.Second
	)

	controller := gomock.NewController(t)

	setupManagerMock(controller)
	setupEventsManagerMock(controller)
	setupContainerFactoryStorageMock(controller)
	setupThingMock(controller)

	testCtrFactory := &containerFactoryFeature{
		mgr:        mockContainerManager,
		storageMgr: mockContainerStorage,
		eventsMgr:  mockEventsManager,
		rootThing:  mockThing,
	}
	defer func() {
		testCtrFactory.dispose()
		controller.Finish()
	}()
	eventChan := make(chan *types.Event)
	errorChan := make(chan error)

	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(eventChan, errorChan)
	testCtrFactory.handleContainerEvents(context.Background())

	tests := map[string]struct {
		chanEvent     *types.Event
		mockExecution mockContainerEventExpectedSetProperty
	}{
		"test_things_container_events_created": {
			chanEvent: &types.Event{
				Type: types.EventTypeContainers,
				Source: types.Container{
					ID:   testContainer.ID,
					Name: testContainer.Name,
					State: &types.State{
						Status: types.Created,
					},
				},
				Action: types.EventActionContainersCreated,
			},
			mockExecution: mockContainerEventExpectedCreated,
		},
		"test_things_container_events_remove": {
			chanEvent: &types.Event{
				Type: types.EventTypeContainers,
				Source: types.Container{
					ID:   testContainer.ID,
					Name: testContainer.Name,
					State: &types.State{
						Status: types.Dead,
					},
				},
				Action: types.EventActionContainersRemoved,
			},
			mockExecution: mockContainerEventExpectedRemoved,
		},
		"test_things_container_events_exited": {
			chanEvent: &types.Event{
				Type: types.EventTypeContainers,
				Source: types.Container{
					ID:   testContainer.ID,
					Name: testContainer.Name,
					State: &types.State{
						Status: types.Exited,
					},
				},
				Action: types.EventActionContainersExited,
			},
			mockExecution: mockContainerEventExpectedStatusExited,
		},
		"test_things_container_events_paused": {
			chanEvent: &types.Event{
				Type: types.EventTypeContainers,
				Source: types.Container{
					ID:   testContainer.ID,
					Name: testContainer.Name,
					State: &types.State{
						Status: types.Paused,
					},
				},
				Action: types.EventActionContainersPaused,
			},
			mockExecution: mockContainerEventExpectedStatusPaused,
		},
		"test_things_container_events_stopped": {
			chanEvent: &types.Event{
				Type: types.EventTypeContainers,
				Source: types.Container{
					ID:   testContainer.ID,
					Name: testContainer.Name,
					State: &types.State{
						Status: types.Stopped,
					},
				},
				Action: types.EventActionContainersStopped,
			},
			mockExecution: mockContainerEventExpectedStatusStopped,
		},
		"test_things_container_events_running": {
			chanEvent: &types.Event{
				Type: types.EventTypeContainers,
				Source: types.Container{
					ID:   testContainer.ID,
					Name: testContainer.Name,
					State: &types.State{
						Status: types.Running,
					},
				},
				Action: types.EventActionContainersRunning,
			},
			mockExecution: mockContainerEventExpectedStatusRunning,
		},
		"test_things_container_events_resumed": {
			chanEvent: &types.Event{
				Type: types.EventTypeContainers,
				Source: types.Container{
					ID:   testContainer.ID,
					Name: testContainer.Name,
					State: &types.State{
						Status: types.Running,
					},
				},
				Action: types.EventActionContainersResumed,
			},
			mockExecution: mockContainerEventExpectedStatusResumed,
		},
		"test_things_container_events_unknown": {
			chanEvent: &types.Event{
				Type:   types.EventTypeContainers,
				Source: copyTestContainer(testContainer),
				Action: types.EventActionContainersResumed,
			},
			mockExecution: mockContainerEventExpectedDefault,
		},
		"test_things_container_events_renamed": {
			chanEvent: &types.Event{
				Type:   types.EventTypeContainers,
				Source: copyTestContainer(testContainer),
				Action: types.EventActionContainersRenamed,
			},
			mockExecution: mockContainerEventExpectedRenamed,
		},
		"test_things_container_events_updated": {
			chanEvent: &types.Event{
				Type: types.EventTypeContainers,
				Source: types.Container{
					ID:   testContainer.ID,
					Name: testContainer.Name,
					HostConfig: &types.HostConfig{
						RestartPolicy: &types.RestartPolicy{
							Type: types.Always,
						},
						Resources: &types.Resources{
							Memory: testMemory,
						},
					},
				},
				Action: types.EventActionContainersUpdated,
			},
			mockExecution: mockContainerEventExpectedUpdated,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			testWg := &sync.WaitGroup{}
			_ = testCase.mockExecution(t, testWg)
			eventChan <- testCase.chanEvent
			testutil.AssertWithTimeout(t, testWg, testEventsTimeout)
		})
	}
}

type mockContainerEventExpectedSetProperty func(t *testing.T, testWg *sync.WaitGroup) error

func mockContainerEventExpectedCreated(t *testing.T, testWg *sync.WaitGroup) error {
	testWg.Add(3)
	mockThing.EXPECT().SetFeature(testContainerFeatureID, gomock.Any()).Times(1).Return(nil).Do(func(id, feature interface{}) { testWg.Done() })
	mockThing.EXPECT().SetFeatureProperty(testContainerFeatureID, containerFeaturePropertyStatus+"/state", fromAPIContainerState(&types.State{Status: types.Created})).Do(func(featureId, propertyId, value interface{}) { testWg.Done() })
	mockContainerStorage.EXPECT().StoreContainerInfo(testContainer.ID).Do(func(ctrId interface{}) { testWg.Done() })
	return nil
}

func mockContainerEventExpectedRemoved(t *testing.T, testWg *sync.WaitGroup) error {
	testWg.Add(3)
	mockThing.EXPECT().SetFeatureProperty(testContainerFeatureID, containerFeaturePropertyStatus+"/state", fromAPIContainerState(&types.State{Status: types.Dead})).Do(func(featureId, propertyId, value interface{}) { testWg.Done() })
	mockThing.EXPECT().RemoveFeature(testContainerFeatureID).Times(1).Return(nil).Do(func(id interface{}) { testWg.Done() })
	mockContainerStorage.EXPECT().DeleteContainerInfo(testContainer.ID).Do(func(ctrId interface{}) { testWg.Done() })
	return nil
}

func mockContainerEventExpectedStatusExited(t *testing.T, testWg *sync.WaitGroup) error {
	testWg.Add(1)
	mockThing.EXPECT().SetFeatureProperty(testContainerFeatureID, "status/state", fromAPIContainerState(&types.State{Status: types.Exited})).Times(1).Return(nil).Do(func(featureId, propertyId, value interface{}) { testWg.Done() })
	return nil
}

func mockContainerEventExpectedStatusPaused(t *testing.T, testWg *sync.WaitGroup) error {
	testWg.Add(1)
	mockThing.EXPECT().SetFeatureProperty(testContainerFeatureID, "status/state", fromAPIContainerState(&types.State{Status: types.Paused})).Times(1).Return(nil).Do(func(featureId, propertyId, value interface{}) { testWg.Done() })
	return nil
}

func mockContainerEventExpectedStatusStopped(t *testing.T, testWg *sync.WaitGroup) error {
	testWg.Add(1)
	mockThing.EXPECT().SetFeatureProperty(testContainerFeatureID, "status/state", fromAPIContainerState(&types.State{Status: types.Stopped})).Times(1).Return(nil).Do(func(featureId, propertyId, value interface{}) { testWg.Done() })
	return nil
}

func mockContainerEventExpectedStatusRunning(t *testing.T, testWg *sync.WaitGroup) error {
	testWg.Add(1)
	mockThing.EXPECT().SetFeatureProperty(testContainerFeatureID, "status/state", fromAPIContainerState(&types.State{Status: types.Running})).Times(1).Return(nil).Do(func(featureId, propertyId, value interface{}) { testWg.Done() })
	return nil
}

func mockContainerEventExpectedStatusResumed(t *testing.T, testWg *sync.WaitGroup) error {
	testWg.Add(1)
	mockThing.EXPECT().SetFeatureProperty(testContainerFeatureID, "status/state", fromAPIContainerState(&types.State{Status: types.Running})).Times(1).Return(nil).Do(func(featureId, propertyId, value interface{}) { testWg.Done() })
	return nil
}

func mockContainerEventExpectedDefault(t *testing.T, testWg *sync.WaitGroup) error {
	testWg.Add(1)
	mockThing.EXPECT().SetFeatureProperty(testContainerFeatureID, "status/state", gomock.Any()).Times(1).Return(nil).Do(func(featureId, propertyId, value interface{}) { testWg.Done() })

	return nil
}

func mockContainerEventExpectedRenamed(t *testing.T, testWg *sync.WaitGroup) error {
	testWg.Add(1)
	mockThing.EXPECT().SetFeatureProperty(testContainerFeatureID, "status/name", testContainer.Name).Times(1).Return(nil).Do(func(featureId, propertyId, value interface{}) { testWg.Done() })
	return nil
}

func mockContainerEventExpectedUpdated(t *testing.T, testWg *sync.WaitGroup) error {
	testWg.Add(1)
	cfg := &configuration{
		RestartPolicy: &restartPolicy{RpType: always},
		Resources:     &resources{Memory: testMemory},
	}
	mockThing.EXPECT().SetFeatureProperty(testContainerFeatureID, "status/config", cfg).Times(1).Return(nil).Do(func(featureId, propertyId, value interface{}) { testWg.Done() })
	return nil
}
