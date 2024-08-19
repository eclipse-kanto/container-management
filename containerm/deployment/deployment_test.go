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
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	mocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/mgr"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
)

const (
	testContainerName1  = "redis"
	testContainerImage1 = "docker.io/library/redis:latest"
	testContainerName2  = "influxdb"
	testContainerImage2 = "docker.io/library/influxdb:1.8.4"
	baseCtrJSONPath     = "../pkg/testutil/config/container"
	testTimeoutDuration = 5 * time.Second
)

var (
	testContainerMatcher = matchers.MatchesContainerImage(testContainerName1, testContainerImage1)
)

func TestDeployCommon(t *testing.T) {
	var (
		validCtrJSONPath = filepath.Join(baseCtrJSONPath, "valid.json")

		testContext = context.Background()
	)

	tests := map[string]struct {
		ctrPath  string
		mockExec func(*mocks.MockContainerManager) error
	}{
		"test_deploy_containers_list_error": {
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				err := log.NewError("test error")
				mockMgr.EXPECT().List(testContext).Return(nil, err).Times(2)
				return err
			},
		},
		"test_deploy_path_is_file_error": {
			ctrPath: validCtrJSONPath,
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				mockMgr.EXPECT().List(testContext).Return(nil, nil).Times(2)
				return log.NewErrorf("the containers deploy path = %s is not a directory", validCtrJSONPath)
			},
		},
		"test_deploy_path_not_exist": {
			ctrPath: filepath.Join(baseCtrJSONPath, "not/exist"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				mockMgr.EXPECT().List(testContext).Return(nil, nil).Times(2)
				return nil
			},
		},
		"test_deploy_path_is_empty": {
			ctrPath: filepath.Join(baseCtrJSONPath, "empty"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				mockMgr.EXPECT().List(testContext).Return(nil, nil).Times(2)
				return nil
			},
		},
		"test_deploy_ignore_non_json_files": {
			ctrPath: filepath.Join(baseCtrJSONPath, "non_json_files"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				mockMgr.EXPECT().List(testContext).Return(nil, nil).Times(2)
				return nil
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			metaPath := createTmpMetaPath(t)
			defer os.Remove(metaPath)

			// set up
			mockCtrl := gomock.NewController(t)
			mockCtrl.Finish()
			defer mockCtrl.Finish()

			mockMgr := mocks.NewMockContainerManager(mockCtrl)

			deployMgr := &deploymentMgr{
				mode:     InitialDeployMode,
				metaPath: metaPath,
				ctrPath:  testCase.ctrPath,
				ctrMgr:   mockMgr,
			}
			expectedErr := testCase.mockExec(mockMgr)
			actualErr := deployMgr.Deploy(testContext)
			testutil.AssertError(t, expectedErr, actualErr)

			deployMgr.mode = UpdateMode
			actualErr = deployMgr.Deploy(testContext)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestInitialDeploy(t *testing.T) {
	var (
		testContext   = context.Background()
		testWaitGroup = &sync.WaitGroup{}
	)

	tests := map[string]struct {
		metaPath string
		ctrPath  string
		mockExec func(*mocks.MockContainerManager) error
	}{
		"test_initial_deploy_containers_not_a_first_run": {
			metaPath: createTmpMetaPathNonEmpty(t),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				return nil
			},
		},
		"test_initial_deploy_containers_exist": {
			metaPath: createTmpMetaPath(t),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{{}}, nil)
				return nil
			},
		},
		"test_initial_deploy_container_create_error": {
			metaPath: createTmpMetaPath(t),
			ctrPath:  filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				mockMgr.EXPECT().List(testContext).Return(nil, nil)
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				mockMgr.EXPECT().Create(testContext, testCtr).Do(
					func(ctx context.Context, container *types.Container) {
						testWaitGroup.Done()
					}).Return(nil, log.NewError("test error")).Times(1)
				return nil
			},
		},
		"test_initial_deploy_container_start_error": {
			metaPath: createTmpMetaPath(t),
			ctrPath:  filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				mockMgr.EXPECT().List(testContext).Return(nil, nil)
				ctrID := uuid.NewString()
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
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
			metaPath: createTmpMetaPath(t),
			ctrPath:  baseCtrJSONPath,
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				mockMgr.EXPECT().List(testContext).Return(nil, nil)

				ctrID1 := uuid.NewString()
				testCtr1 := newTestContainer(testContainerName1, testContainerImage1)
				mockMgr.EXPECT().Create(testContext, testCtr1).Do(
					func(ctx context.Context, container *types.Container) {
						testCtr1.ID = ctrID1
					}).Return(testCtr1, nil).Times(1)
				mockMgr.EXPECT().Start(testContext, ctrID1).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
				}).Return(nil).Times(1)

				testWaitGroup.Add(1)
				ctrID2 := uuid.NewString()
				testCtr2 := newTestContainer(testContainerName2, testContainerImage2)
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
			defer os.Remove(testCase.metaPath)

			// set up
			mockCtrl := gomock.NewController(t)
			mockCtrl.Finish()
			defer mockCtrl.Finish()

			mockMgr := mocks.NewMockContainerManager(mockCtrl)

			deployMgr := &deploymentMgr{
				mode:     InitialDeployMode,
				metaPath: testCase.metaPath,
				ctrPath:  testCase.ctrPath,
				ctrMgr:   mockMgr,
			}

			expectedErr := testCase.mockExec(mockMgr)
			actualErr := deployMgr.Deploy(testContext)
			testutil.AssertError(t, expectedErr, actualErr)
			testutil.AssertWithTimeout(t, testWaitGroup, testTimeoutDuration)
		})
	}
}

func TestUpdate(t *testing.T) {
	var (
		testContext   = context.Background()
		testWaitGroup = &sync.WaitGroup{}
	)

	tests := map[string]struct {
		ctrPath  string
		mockExec func(*mocks.MockContainerManager) error
	}{
		"test_update_new_container_no_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				mockMgr.EXPECT().List(testContext).Return(nil, nil)
				ctrID := uuid.NewString()
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				util.FillDefaults(testCtr)
				testCtr.ID = ctrID
				mockMgr.EXPECT().Create(testContext, testContainerMatcher).Return(testCtr, nil).Times(1)
				mockMgr.EXPECT().Start(testContext, ctrID).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
				}).Return(nil).Times(1)
				return nil
			},
		},
		"test_update_new_container_create_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				mockMgr.EXPECT().List(testContext).Return(nil, nil)
				mockMgr.EXPECT().Create(testContext, testContainerMatcher).Do(func(ctx context.Context, container *types.Container) {
					testWaitGroup.Done()
				}).Return(nil, log.NewError("test error")).Times(1)
				return nil
			},
		},
		"test_update_new_container_start_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				mockMgr.EXPECT().List(testContext).Return(nil, nil)
				ctrID := uuid.NewString()
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				util.FillDefaults(testCtr)
				testCtr.ID = ctrID
				mockMgr.EXPECT().Create(testContext, testContainerMatcher).Return(testCtr, nil).Times(1)
				mockMgr.EXPECT().Start(testContext, ctrID).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
				}).Return(log.NewError("test error")).Times(1)
				return nil
			},
		},
		"test_update_check_running_container_no_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				testCtr.State = &types.State{Running: true}
				util.FillDefaults(testCtr)
				mockMgr.EXPECT().List(testContext).Do(func(ctx context.Context) {
					testWaitGroup.Done()
				}).Return([]*types.Container{testCtr}, nil)
				return nil
			},
		},
		"test_update_restart_nonrunning_container_no_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				testCtr.State = &types.State{Running: false, Paused: false}
				util.FillDefaults(testCtr)
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Start(testContext, testCtr.ID).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
					testCtr.State.Running = true
				}).Return(nil).Times(1)
				return nil
			},
		},
		"test_update_restart_nonrunning_container_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				testCtr.State = &types.State{Running: false, Paused: false}
				util.FillDefaults(testCtr)
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Start(testContext, testCtr.ID).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
				}).Return(log.NewError("test error")).Times(1)
				return nil
			},
		},
		"test_update_restart_paused_container_no_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				testCtr.State = &types.State{Running: false, Paused: true}
				util.FillDefaults(testCtr)
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Unpause(testContext, testCtr.ID).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
					testCtr.State.Running = true
					testCtr.State.Running = false
				}).Return(nil).Times(1)
				return nil
			},
		},
		"test_update_restart_paused_container_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				testCtr.State = &types.State{Running: false, Paused: true}
				util.FillDefaults(testCtr)
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Unpause(testContext, testCtr.ID).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
				}).Return(log.NewError("test error")).Times(1)
				return nil
			},
		},
		"test_update_modify_container_restart_policy_no_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				testCtr.State = &types.State{Running: true}
				util.FillDefaults(testCtr)
				testUpdateOpts := &types.UpdateOpts{
					RestartPolicy: testCtr.HostConfig.RestartPolicy,
					Resources:     testCtr.HostConfig.Resources,
				}
				testCtr.HostConfig.RestartPolicy = &types.RestartPolicy{Type: types.No}
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Update(testContext, testCtr.ID, testUpdateOpts).Do(func(ctx context.Context, ctrID string, updateOpts *types.UpdateOpts) {
					testWaitGroup.Done()
					testCtr.HostConfig.RestartPolicy = updateOpts.RestartPolicy
					testCtr.HostConfig.Resources = updateOpts.Resources
				}).Return(nil).Times(1)
				return nil
			},
		},
		"test_update_modify_nonrunning_container_restart_policy_no_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				testCtr.State = &types.State{}
				util.FillDefaults(testCtr)
				testUpdateOpts := &types.UpdateOpts{
					RestartPolicy: testCtr.HostConfig.RestartPolicy,
					Resources:     testCtr.HostConfig.Resources,
				}
				testCtr.HostConfig.RestartPolicy = &types.RestartPolicy{Type: types.No}
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Update(testContext, testCtr.ID, testUpdateOpts).Do(func(ctx context.Context, ctrID string, updateOpts *types.UpdateOpts) {
					testCtr.HostConfig.RestartPolicy = updateOpts.RestartPolicy
					testCtr.HostConfig.Resources = updateOpts.Resources
				}).Return(nil).Times(1)
				mockMgr.EXPECT().Start(testContext, testCtr.ID).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
					testCtr.State.Running = true
				}).Return(nil).Times(1)
				return nil
			},
		},
		"test_update_modify_paused_container_restart_policy_no_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				testCtr.State = &types.State{Paused: true}
				util.FillDefaults(testCtr)
				testUpdateOpts := &types.UpdateOpts{
					RestartPolicy: testCtr.HostConfig.RestartPolicy,
					Resources:     testCtr.HostConfig.Resources,
				}
				testCtr.HostConfig.RestartPolicy = &types.RestartPolicy{Type: types.No}
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Update(testContext, testCtr.ID, testUpdateOpts).Do(func(ctx context.Context, ctrID string, updateOpts *types.UpdateOpts) {
					testCtr.HostConfig.RestartPolicy = updateOpts.RestartPolicy
					testCtr.HostConfig.Resources = updateOpts.Resources
				}).Return(nil).Times(1)
				mockMgr.EXPECT().Unpause(testContext, testCtr.ID).Do(func(ctx context.Context, ctrID string) {
					testWaitGroup.Done()
					testCtr.State.Running = true
				}).Return(nil).Times(1)
				return nil
			},
		},
		"test_update_modify_container_restart_policy_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage1)
				testCtr.State = &types.State{Running: true}
				util.FillDefaults(testCtr)
				testUpdateOpts := &types.UpdateOpts{
					RestartPolicy: testCtr.HostConfig.RestartPolicy,
					Resources:     testCtr.HostConfig.Resources,
				}
				testCtr.HostConfig.RestartPolicy = &types.RestartPolicy{Type: types.No}
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Update(testContext, testCtr.ID, testUpdateOpts).Do(func(ctx context.Context, ctrID string, updateOpts *types.UpdateOpts) {
					testWaitGroup.Done()
				}).Return(log.NewError("test error")).Times(1)
				return nil
			},
		},
		"test_update_existing_running_container_no_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage2)
				util.FillDefaults(testCtr)
				testCtr.State = &types.State{Running: true}
				testStopOpts := &types.StopOpts{Force: true, Signal: "SIGTERM"}
				oldID := testCtr.ID
				newContainer := newTestContainer(testContainerName1, testContainerImage1)
				util.FillDefaults(newContainer)
				newID := uuid.NewString()
				newContainer.ID = newID
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Create(testContext, testContainerMatcher).Return(newContainer, nil).Times(1)
				mockMgr.EXPECT().Stop(testContext, oldID, testStopOpts).Return(nil).Times(1)
				mockMgr.EXPECT().Start(testContext, newID).Return(nil).Times(1)
				mockMgr.EXPECT().Remove(testContext, oldID, true, false, nil).Do(func(ctx context.Context, ctrID string, force bool, stopOpts *types.StopOpts) {
					testWaitGroup.Done()
				}).Return(nil).Times(1)
				return nil
			},
		},
		"test_update_existing_nonrunning_container_no_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage2)
				util.FillDefaults(testCtr)
				testCtr.State = &types.State{Running: false}
				oldID := testCtr.ID
				newContainer := newTestContainer(testContainerName1, testContainerImage1)
				util.FillDefaults(newContainer)
				newID := uuid.NewString()
				newContainer.ID = newID
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Create(testContext, testContainerMatcher).Return(newContainer, nil).Times(1)
				mockMgr.EXPECT().Start(testContext, newID).Return(nil).Times(1)
				mockMgr.EXPECT().Remove(testContext, oldID, true, false, nil).Do(func(ctx context.Context, ctrID string, force bool, stopOpts *types.StopOpts) {
					testWaitGroup.Done()
				}).Return(nil).Times(1)
				return nil
			},
		},
		"test_update_existing_running_container_create_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage2)
				util.FillDefaults(testCtr)
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Create(testContext, testContainerMatcher).Do(func(ctx context.Context, container *types.Container) {
					testWaitGroup.Done()
				}).Return(nil, log.NewError("test error")).Times(1)
				return nil
			},
		},
		"test_update_existing_running_container_start_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage2)
				util.FillDefaults(testCtr)
				testCtr.State = &types.State{Running: true}
				testStopOpts := &types.StopOpts{Force: true, Signal: "SIGTERM"}
				oldID := testCtr.ID
				newContainer := newTestContainer(testContainerName1, testContainerImage1)
				util.FillDefaults(newContainer)
				newID := uuid.NewString()
				newContainer.ID = newID
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Create(testContext, testContainerMatcher).Return(newContainer, nil).Times(1)
				mockMgr.EXPECT().Stop(testContext, oldID, testStopOpts).Return(nil).Times(1)
				mockMgr.EXPECT().Start(testContext, newID).Return(log.NewError("test error")).Times(1)
				mockMgr.EXPECT().Start(testContext, oldID).Return(nil).Times(1)
				mockMgr.EXPECT().Remove(testContext, newID, true, false, nil).Do(func(ctx context.Context, ctrID string, force bool, stopOpts *types.StopOpts) {
					testWaitGroup.Done()
				}).Return(nil).Times(1)
				return nil
			},
		},
		"test_update_existing_running_container_stop_start_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage2)
				util.FillDefaults(testCtr)
				testCtr.State = &types.State{Running: true}
				testStopOpts := &types.StopOpts{Force: true, Signal: "SIGTERM"}
				oldID := testCtr.ID
				newContainer := newTestContainer(testContainerName1, testContainerImage1)
				util.FillDefaults(newContainer)
				newID := uuid.NewString()
				newContainer.ID = newID
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Create(testContext, testContainerMatcher).Return(newContainer, nil).Times(1)
				mockMgr.EXPECT().Stop(testContext, oldID, testStopOpts).Return(log.NewError("test error")).Times(1)
				mockMgr.EXPECT().Start(testContext, newID).Return(log.NewError("test error")).Times(1)
				mockMgr.EXPECT().Remove(testContext, newID, true, false, nil).Do(func(ctx context.Context, ctrID string, force bool, stopOpts *types.StopOpts) {
					testWaitGroup.Done()
				}).Return(nil).Times(1)
				return nil
			},
		},
		"test_update_existing_running_container_remove_error": {
			ctrPath: filepath.Join(baseCtrJSONPath, "nested"),
			mockExec: func(mockMgr *mocks.MockContainerManager) error {
				testWaitGroup.Add(1)
				testCtr := newTestContainer(testContainerName1, testContainerImage2)
				util.FillDefaults(testCtr)
				testCtr.State = &types.State{Running: true}
				testStopOpts := &types.StopOpts{Force: true, Signal: "SIGTERM"}
				oldID := testCtr.ID
				newContainer := newTestContainer(testContainerName1, testContainerImage1)
				util.FillDefaults(newContainer)
				newID := uuid.NewString()
				newContainer.ID = newID
				mockMgr.EXPECT().List(testContext).Return([]*types.Container{testCtr}, nil)
				mockMgr.EXPECT().Create(testContext, testContainerMatcher).Return(newContainer, nil).Times(1)
				mockMgr.EXPECT().Stop(testContext, oldID, testStopOpts).Return(nil).Times(1)
				mockMgr.EXPECT().Start(testContext, newID).Return(nil).Times(1)
				mockMgr.EXPECT().Remove(testContext, oldID, true, false, nil).Do(func(ctx context.Context, ctrID string, force bool, stopOpts *types.StopOpts) {
					testWaitGroup.Done()
				}).Return(log.NewError("test error")).Times(1)
				return nil
			},
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			metaPath := createTmpMetaPathNonEmpty(t)
			defer os.Remove(metaPath)

			// set up
			mockCtrl := gomock.NewController(t)
			mockCtrl.Finish()
			defer mockCtrl.Finish()

			mockMgr := mocks.NewMockContainerManager(mockCtrl)

			deployMgr := &deploymentMgr{
				mode:     UpdateMode,
				metaPath: metaPath,
				ctrPath:  testCase.ctrPath,
				ctrMgr:   mockMgr,
			}

			expectedErr := testCase.mockExec(mockMgr)
			actualErr := deployMgr.Deploy(testContext)
			testutil.AssertError(t, expectedErr, actualErr)
			testutil.AssertWithTimeout(t, testWaitGroup, testTimeoutDuration)
		})
	}
}

func TestDispose(t *testing.T) {
	deployMgr := &deploymentMgr{}
	err := deployMgr.Dispose(context.Background())
	testutil.AssertNil(t, err)
}

func createTmpMetaPath(t *testing.T) string {
	path, err := os.MkdirTemp("", "container-management-test-")
	testutil.AssertNil(t, err)
	return path
}

func createTmpMetaPathNonEmpty(t *testing.T) string {
	path := createTmpMetaPath(t)
	err := util.MkDir(filepath.Join(path, "deployment"))
	testutil.AssertNil(t, err)
	return path
}

func newTestContainer(name, image string) *types.Container {
	return &types.Container{Name: name, Image: types.Image{Name: image}}
}
