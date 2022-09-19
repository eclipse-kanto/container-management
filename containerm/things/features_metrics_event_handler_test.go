// Copyright (c) 2022 Contributors to the Eclipse Foundation
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

func TestMetricsHandleContainerEvents(t *testing.T) {
	const testEventsTimeout = 5 * time.Second

	controller := gomock.NewController(t)

	setupManagerMock(controller)
	setupEventsManagerMock(controller)
	setupThingMock(controller)

	testMetrics := &metricsFeature{
		mgr:       mockContainerManager,
		eventsMgr: mockEventsManager,
		rootThing: mockThing,
	}
	defer func() {
		testMetrics.dispose()
		controller.Finish()
	}()
	eventChan := make(chan *types.Event)
	errorChan := make(chan error)

	mockEventsManager.EXPECT().Subscribe(gomock.Any()).Times(1).Return(eventChan, errorChan)
	testMetrics.handleContainerEvents(context.Background())
	testMetrics.request = new(Request)
	testMetrics.previousCPU = make(map[string]*types.CPUStats)
	event := &types.Event{
		Type: types.EventTypeContainers,
		Source: types.Container{
			ID:   testContainer.ID,
			Name: testContainer.Name,
			State: &types.State{
				Status: types.Running,
			},
		},
		Action: types.EventActionContainersRunning,
	}

	testWg := &sync.WaitGroup{}
	testWg.Add(1)
	mockContainerManager.EXPECT().Metrics(gomock.Any(), testContainer.ID).Return(&types.Metrics{CPU: &types.CPUStats{Used: 10000, Total: 100000}}, nil).Times(1).Do(func(_, _ interface{}) { testWg.Done() })
	eventChan <- event
	testutil.AssertWithTimeout(t, testWg, testEventsTimeout)
}
