// Copyright (c) 2023 Contributors to the Eclipse Foundation
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

package deployment

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/mgr"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

func TestInitialDeploy(t *testing.T) {
	const testTimeoutDuration = 5 * time.Second

	testContext := context.Background()
	testWaitGroup := &sync.WaitGroup{}

	tests := map[string]struct {
		initialDeployPath string
		mockExec          func(*mocks.MockContainerManager) error
	}{
		"test_initial_deploy_containers_exist": {
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{{}}, nil)
				return nil
			},
		},
		"test_initial_deploy_containers_list_error": {
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				err := log.NewError("test error")
				mockMgr.EXPECT().List(testContext).Return(nil, err)
				return err
			},
		},
		"test_initial_deploy_path_not_exist": {
			initialDeployPath: "../pkg/testutil/config/container/not/exist",
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				mockMgr.EXPECT().List(testContext).Return(nil, nil)
				return nil
			},
		},
		"test_initial_deploy_path_is_empty": {
			initialDeployPath: "../pkg/testutil/config/container/empty",
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				mockMgr.EXPECT().List(testContext).Return(nil, nil)
				return nil
			},
		},
		"test_initial_deploy_container_create_error": {
			initialDeployPath: "../pkg/testutil/config/container/nested",
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				mockMgr.EXPECT().List(testContext).Return(nil, nil)
				testCtr := &types.Container{Image: types.Image{Name: "docker.io/library/redis:latest"}}
				mockMgr.EXPECT().Create(testContext, testCtr).Do(
					func(ctx context.Context, container *types.Container) {
						testWaitGroup.Done()
					}).Return(nil, log.NewError("test error")).Times(1)
				return nil
			},
		},
		"test_initial_deploy_container_start_error": {
			initialDeployPath: "../pkg/testutil/config/container/nested",
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				mockMgr.EXPECT().List(testContext).Return(nil, nil)
				ctrID := uuid.NewString()
				testCtr := &types.Container{Image: types.Image{Name: "docker.io/library/redis:latest"}}
				mockMgr.EXPECT().Create(testContext, testCtr).Do(
					func(ctx context.Context, container *types.Container) {
						testCtr.ID = ctrID
					}).Return(testCtr, nil).Times(1)
				mockMgr.EXPECT().Start(testContext, ctrID).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
				}).Return(log.NewError("test error")).Times(1)
				return nil
			},
		},
		"test_initial_deploy_multiple_containers": {
			initialDeployPath: "../pkg/testutil/config/container",
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				mockMgr.EXPECT().List(testContext).Return(nil, nil)

				ctrID1 := uuid.NewString()
				testCtr1 := &types.Container{Image: types.Image{Name: "docker.io/library/redis:latest"}}
				mockMgr.EXPECT().Create(testContext, testCtr1).Do(
					func(ctx context.Context, container *types.Container) {
						testCtr1.ID = ctrID1
					}).Return(testCtr1, nil).Times(1)
				mockMgr.EXPECT().Start(testContext, ctrID1).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
				}).Return(nil).Times(1)

				testWaitGroup.Add(1)
				ctrID2 := uuid.NewString()
				testCtr2 := &types.Container{Image: types.Image{Name: "docker.io/library/influxdb:latest"}}
				mockMgr.EXPECT().Create(testContext, testCtr2).Do(
					func(ctx context.Context, container *types.Container) {
						testCtr2.ID = ctrID2
					}).Return(testCtr2, nil).Times(1)
				mockMgr.EXPECT().Start(testContext, ctrID2).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
				}).Return(nil).Times(1)
				return nil
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			// set up
			mockCtrl := gomock.NewController(t)
			mockCtrl.Finish()
			defer mockCtrl.Finish()

			mockMgr := mocks.NewMockContainerManager(mockCtrl)

			deployMgr, err := newDeploymentMgr(testCase.initialDeployPath, mockMgr)
			testutil.AssertNil(t, err)

			expectedErr := testCase.mockExec(mockMgr)
			actualErr := deployMgr.InitialDeploy(testContext)
			testutil.AssertEqual(t, expectedErr, actualErr)
			testutil.AssertWithTimeout(t, testWaitGroup, testTimeoutDuration)
		})
	}
}

func TestDispose(t *testing.T) {
	deployMgr, err := newDeploymentMgr("", nil)
	testutil.AssertNil(t, err)

	err = deployMgr.Dispose(context.Background())
	testutil.AssertNil(t, err)
}
