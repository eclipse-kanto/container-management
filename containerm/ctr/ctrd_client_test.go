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

package ctr

import (
	"context"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	containerdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	ctrdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctrd"
	ioMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/io"
	loggerMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/logger"
	streamsMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/streams"
	"github.com/eclipse-kanto/container-management/containerm/streams"
	"github.com/eclipse-kanto/container-management/containerm/util"

	statsV1 "github.com/containerd/cgroups/stats/v1"
	"github.com/containerd/containerd"
	containerdtypes "github.com/containerd/containerd/api/types"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/snapshots"
	"github.com/containerd/imgcrypt"
	"github.com/containerd/imgcrypt/images/encryption"
	"github.com/containerd/typeurl"
	"github.com/containers/ocicrypt/config"
	"github.com/golang/mock/gomock"
	"github.com/opencontainers/runtime-spec/specs-go"
)

var testContainerID = "test-container-id"

func TestCtrdClientCreateContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockIoMgr := NewMockcontainerIOManager(mockCtrl)
	mockLogMgr := ctrdMocks.NewMockcontainerLogsManager(mockCtrl)
	mockDecrypctMgr := ctrdMocks.NewMockcontainerDecryptMgr(mockCtrl)
	mockLogDriver := loggerMocks.NewMockLogDriver(mockCtrl)
	mockSpi := ctrdMocks.NewMockcontainerdSpi(mockCtrl)
	mockImage := containerdMocks.NewMockImage(mockCtrl)

	testClient := &containerdClient{
		ioMgr:   mockIoMgr,
		logsMgr: mockLogMgr,
		decMgr:  mockDecrypctMgr,
		spi:     mockSpi,
	}

	testCtr := &types.Container{
		ID: testContainerID,
		Image: types.Image{
			Name:          "test.host/name:latest",
			DecryptConfig: &types.DecryptConfig{},
		},
		HostConfig: &types.HostConfig{
			LogConfig: &types.LogConfiguration{
				ModeConfig: &types.LogModeConfiguration{},
			},
		},
		IOConfig: &types.IOConfig{
			OpenStdin: false,
			Tty:       false,
		},
	}
	ctx := context.Background()

	tests := map[string]struct {
		mockExec func() error
	}{
		"test_error_initialise_io": {
			mockExec: func() error {
				err := log.NewErrorf("failed to initialise IO for container ID = %s", testCtr.ID)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, err)
				mockIoMgr.EXPECT().ClearIO(testCtr.ID).Return(nil)
				mockSpi.EXPECT().RemoveSnapshot(ctx, testCtr.ID).Return(nil)
				mockSpi.EXPECT().UnmountSnapshot(ctx, testCtr.ID, rootFSPathDefault).Return(nil)
				return err
			},
		},
		"test_error_initialize_logger": {
			mockExec: func() error {
				err := log.NewErrorf("failed to initialize logger for container ID = %s", testCtr.ID)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockIoMgr.EXPECT().ClearIO(testCtr.ID).Return(nil)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(nil, err)
				mockSpi.EXPECT().RemoveSnapshot(ctx, testCtr.ID).Return(nil)
				mockSpi.EXPECT().UnmountSnapshot(ctx, testCtr.ID, rootFSPathDefault).Return(nil)
				return err
			},
		},
		"test_error_get_image": {
			mockExec: func() error {
				err := log.NewErrorf("error while trying to get container image with ID = %s for container ID = %s ", testCtr.Image.Name, testCtr.ID)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockIoMgr.EXPECT().ClearIO(testCtr.ID).Return(nil)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				mockDecrypctMgr.EXPECT().GetDecryptConfig(testCtr.Image.DecryptConfig).Return(nil, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(nil, err)
				mockSpi.EXPECT().RemoveSnapshot(ctx, testCtr.ID).Return(nil)
				mockSpi.EXPECT().UnmountSnapshot(ctx, testCtr.ID, rootFSPathDefault).Return(nil)
				return err
			},
		},
		"test_create_container_without_error": {
			mockExec: func() error {
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				dc := &config.DecryptConfig{}
				mockDecrypctMgr.EXPECT().GetDecryptConfig(testCtr.Image.DecryptConfig).Times(2).Return(dc, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockDecrypctMgr.EXPECT().CheckAuthorization(ctx, mockImage, dc).Return(nil)
				mockSpi.EXPECT().PrepareSnapshot(ctx, testCtr.ID, mockImage, matchers.MatchesUnpackOpts(encryption.WithUnpackConfigApplyOpts(encryption.WithDecryptedUnpack(&imgcrypt.Payload{DecryptConfig: *dc})))).Return(nil)
				mockSpi.EXPECT().MountSnapshot(ctx, testCtr.ID, rootFSPathDefault)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertError(t, testCase.mockExec(), testClient.CreateContainer(ctx, testCtr, ""))
		})
	}
}

func TestCtrdClientDestroyContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockIoMgr := NewMockcontainerIOManager(mockCtrl)
	mockSpi := ctrdMocks.NewMockcontainerdSpi(mockCtrl)
	mockTask := containerdMocks.NewMockTask(mockCtrl)
	mockImg := containerdMocks.NewMockImage(mockCtrl)
	mockResMgr := NewMockresourcesWatcher(mockCtrl)

	ctx := context.Background()
	stopOpts := &types.StopOpts{}

	notExistingContainerID := "test-id"
	testContainerImageRef := "test.img/ref:latest"

	tests := map[string]struct {
		testClient *containerdClient
		testCtr    *types.Container
		mockExec   func() error
	}{
		"test_error_not_existing_container": {
			testClient: &containerdClient{
				ctrdCache: &containerInfoCache{
					cache: map[string]*containerInfo{
						testContainerID: {
							c: &types.Container{
								ID: testContainerID,
							},
						},
					},
				},
				ioMgr:         mockIoMgr,
				spi:           mockSpi,
				imagesWatcher: mockResMgr,
			},
			testCtr: &types.Container{
				ID: notExistingContainerID,
				Image: types.Image{
					Name: testContainerImageRef,
				},
			},
			mockExec: func() error {
				err := log.NewErrorf("container with ID = test-id does not exist")
				mockIoMgr.EXPECT().ClearIO(notExistingContainerID).Return(nil)
				mockSpi.EXPECT().RemoveSnapshot(ctx, notExistingContainerID).Return(nil)
				mockSpi.EXPECT().UnmountSnapshot(ctx, notExistingContainerID, rootFSPathDefault).Return(nil)
				mockSpi.EXPECT().GetImage(ctx, testContainerImageRef).Return(nil, errdefs.ErrNotFound)
				return err
			},
		},
		"test_error_kill_task": {
			testClient: &containerdClient{
				ctrdCache: &containerInfoCache{
					cache: map[string]*containerInfo{
						testContainerID: {
							c: &types.Container{
								ID: testContainerID,
							},
							task: mockTask,
						},
					},
				},
				ioMgr:         mockIoMgr,
				spi:           mockSpi,
				imagesWatcher: mockResMgr,
			},
			testCtr: &types.Container{
				ID: testContainerID,
			},
			mockExec: func() error {
				err := log.NewErrorf("Kill Error")
				mockTask.EXPECT().Kill(ctx, util.ToSignal(stopOpts.Signal)).Return(err)
				return err
			},
		},
		"test_destroy_container_without_error": {
			testClient: &containerdClient{
				ctrdCache: &containerInfoCache{
					cache: map[string]*containerInfo{
						testContainerID: {
							c: &types.Container{
								ID: testContainerID,
							},
						},
					},
				},
				ioMgr:         mockIoMgr,
				spi:           mockSpi,
				imagesWatcher: mockResMgr,
				imageExpiry:   24 * time.Hour,
			},
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: testContainerImageRef,
				},
			},
			mockExec: func() error {
				mockIoMgr.EXPECT().ClearIO(testContainerID).Return(nil)
				mockSpi.EXPECT().RemoveSnapshot(ctx, testContainerID).Return(nil)
				mockSpi.EXPECT().UnmountSnapshot(ctx, testContainerID, rootFSPathDefault).Return(nil)
				mockSpi.EXPECT().GetImage(ctx, testContainerImageRef).Return(mockImg, nil)
				mockImg.EXPECT().Metadata().Return(images.Image{CreatedAt: time.Now().Add(-12 * time.Hour)})
				mockImg.EXPECT().Name().Return(testContainerImageRef).Times(1)
				mockResMgr.EXPECT().Watch(testContainerImageRef, gomock.Any(), gomock.Any()).Return(nil)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			expectedError := testCase.mockExec()
			if expectedError == nil {
				testCase.testClient.DestroyContainer(ctx, testCase.testCtr, stopOpts, true)
				testutil.AssertTrue(t, len(testCase.testClient.ctrdCache.cache) == 0)
			} else {
				_, _, actualError := testCase.testClient.DestroyContainer(ctx, testCase.testCtr, stopOpts, true)
				testutil.AssertError(t, expectedError, actualError)
			}
		})
	}
}

func TestCtrdClientStartContainer(t *testing.T) {
	const (
		testCtrTaskPID uint32 = 123
	)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSpi := ctrdMocks.NewMockcontainerdSpi(mockCtrl)
	mockIoMgr := NewMockcontainerIOManager(mockCtrl)
	mockLogMgr := ctrdMocks.NewMockcontainerLogsManager(mockCtrl)
	mockLogDriver := loggerMocks.NewMockLogDriver(mockCtrl)
	mockDecMgr := ctrdMocks.NewMockcontainerDecryptMgr(mockCtrl)
	mockImage := containerdMocks.NewMockImage(mockCtrl)
	mockContainer := containerdMocks.NewMockContainer(mockCtrl)
	mockTask := containerdMocks.NewMockTask(mockCtrl)
	mockIo := containerdMocks.NewMockIO(mockCtrl)

	testClient := &containerdClient{
		ctrdCache: newContainerInfoCache(),
		ioMgr:     mockIoMgr,
		logsMgr:   mockLogMgr,
		decMgr:    mockDecMgr,
		spi:       mockSpi,
	}
	ctx := context.Background()

	tests := map[string]struct {
		testCtr  *types.Container
		mockExec func(testCtr *types.Container) error
	}{
		"test_error_check_existing_container": {
			testCtr: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("error trying to check for existing container ID = %s", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testContainerID).Return(nil, err)
				return err
			},
		},
		"test_error_existing_container": {
			testCtr: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("container with ID = %s already exists", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(mockContainer, nil)
				return err
			},
		},
		"test_error_snapshot_for_container_not_exist": {
			testCtr: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("snapshot for container with ID = %s does not exist", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, err)
				return err
			},
		},
		"test_error_missing_container_image": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("missing image ID = %s for container with ID = %s", testCtr.Image.Name, testContainerID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockDecMgr.EXPECT().GetDecryptConfig(testCtr.Image.DecryptConfig).Return(nil, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(nil, err)
				return err
			},
		},
		"test_error_generating_container_opts": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("missing image ID = %s for container with ID = %s", testCtr.Image.Name, testContainerID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				dc := &config.DecryptConfig{}
				mockDecMgr.EXPECT().GetDecryptConfig(testCtr.Image.DecryptConfig).Return(dc, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockDecMgr.EXPECT().CheckAuthorization(ctx, mockImage, dc).Return(nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockDecMgr.EXPECT().GetDecryptConfig(testCtr.Image.DecryptConfig).Return(nil, err)
				return err
			},
		},
		"test_error_creating_new_container": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						ModeConfig: &types.LogModeConfiguration{},
					},
				},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("error creating new container with ID = %s", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				dc := &config.DecryptConfig{}
				mockDecMgr.EXPECT().GetDecryptConfig(testCtr.Image.DecryptConfig).Times(2).Return(dc, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockDecMgr.EXPECT().CheckAuthorization(ctx, mockImage, dc).Return(nil)
				mockSpi.EXPECT().CreateContainer(ctx, testCtr.ID, gomock.Any() /*for now..*/).Return(nil, err)
				return err
			},
		},
		"test_error_initialise_IO": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						ModeConfig: &types.LogModeConfiguration{},
					},
				},
				IOConfig: &types.IOConfig{
					OpenStdin: false,
					Tty:       false,
				},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("failed to initialise IO for container ID = %s", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				dc := &config.DecryptConfig{}
				mockDecMgr.EXPECT().GetDecryptConfig(testCtr.Image.DecryptConfig).Times(2).Return(dc, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockDecMgr.EXPECT().CheckAuthorization(ctx, mockImage, dc).Return(nil)
				mockSpi.EXPECT().CreateContainer(ctx, testCtr.ID, gomock.Any() /*for now..*/).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(false)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, err)
				mockContainer.EXPECT().Delete(ctx)
				return err
			},
		},
		"test_error_initialize_logger": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						ModeConfig: &types.LogModeConfiguration{},
					},
				},
				IOConfig: &types.IOConfig{
					OpenStdin: false,
					Tty:       false,
				},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("failed to initialize logger for container ID = %s", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				dc := &config.DecryptConfig{}
				mockDecMgr.EXPECT().GetDecryptConfig(testCtr.Image.DecryptConfig).Times(2).Return(dc, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockDecMgr.EXPECT().CheckAuthorization(ctx, mockImage, dc).Return(nil)
				mockSpi.EXPECT().CreateContainer(ctx, testCtr.ID, gomock.Any() /*for now..*/).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(false)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockContainer.EXPECT().Delete(ctx)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(nil, err)
				return err
			},
		},
		"test_error_creating_task": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						ModeConfig: &types.LogModeConfiguration{},
					},
				},
				IOConfig: &types.IOConfig{
					OpenStdin: false,
					Tty:       false,
				},
			},
			mockExec: func(testCtr *types.Container) error {
				err := log.NewErrorf("error creating task for container ID = %s", testCtr.ID)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				dc := &config.DecryptConfig{}
				mockDecMgr.EXPECT().GetDecryptConfig(testCtr.Image.DecryptConfig).Times(2).Return(dc, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockDecMgr.EXPECT().CheckAuthorization(ctx, mockImage, dc).Return(nil)
				mockSpi.EXPECT().CreateContainer(ctx, testCtr.ID, gomock.Any()).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(false)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				cioCreator := func(id string) (cio.IO, error) { return mockIo, nil }
				mockIoMgr.EXPECT().NewCioCreator(testCtr.IOConfig.Tty).Return(cioCreator)
				mockSpi.EXPECT().CreateTask(ctx, mockContainer, gomock.Any()).Return(nil, err)
				mockIoMgr.EXPECT().ResetIO(testCtr.ID)
				mockContainer.EXPECT().Delete(ctx)
				return err
			},
		},
		"test_create_container_without_error": {
			testCtr: &types.Container{
				ID: testContainerID,
				Image: types.Image{
					Name: "test.host/name:latest",
				},
				HostConfig: &types.HostConfig{
					LogConfig: &types.LogConfiguration{
						ModeConfig: &types.LogModeConfiguration{},
					},
				},
				IOConfig: &types.IOConfig{
					OpenStdin: false,
					Tty:       false,
				},
			},
			mockExec: func(testCtr *types.Container) error {
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil)
				mockSpi.EXPECT().GetSnapshotID(testCtr.ID).Return(testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				dc := &config.DecryptConfig{}
				mockDecMgr.EXPECT().GetDecryptConfig(testCtr.Image.DecryptConfig).Times(2).Return(dc, nil)
				mockSpi.EXPECT().GetImage(ctx, testCtr.Image.Name).Return(mockImage, nil)
				mockDecMgr.EXPECT().CheckAuthorization(ctx, mockImage, dc).Return(nil)
				mockSpi.EXPECT().CreateContainer(ctx, testCtr.ID, gomock.Any() /*for now..*/).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(false)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, nil)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				cioCreator := func(id string) (cio.IO, error) { return mockIo, nil }
				mockIoMgr.EXPECT().NewCioCreator(testCtr.IOConfig.Tty).Return(cioCreator)
				mockSpi.EXPECT().CreateTask(ctx, mockContainer, gomock.Any() /*for now..*/).Return(mockTask, nil)
				mockContainer.EXPECT().ID().Return(testCtr.ID)
				resChan := make(<-chan containerd.ExitStatus)
				mockTask.EXPECT().Wait(context.TODO()).Return(resChan, nil)
				mockTask.EXPECT().Start(ctx).Return(nil)
				mockTask.EXPECT().Pid().Return(testCtrTaskPID).Times(2)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			expectedError := testCase.mockExec(testCase.testCtr)
			_, actualError := testClient.StartContainer(ctx, testCase.testCtr, "")
			testutil.AssertError(t, expectedError, actualError)
		})
	}
}

func TestAttachContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockIoMgr := NewMockcontainerIOManager(mockCtrl)
	mockReadCloser := ioMocks.NewMockReadCloser(mockCtrl)
	mockStream := streamsMocks.NewMockStream(mockCtrl)
	mockIO := NewMockIO(mockCtrl)
	attachConfig := &streams.AttachConfig{UseStdin: true, Stdin: mockReadCloser}
	ctx := context.Background()

	testClient := &containerdClient{
		ioMgr: mockIoMgr,
	}

	testCtr := &types.Container{
		ID: "test-id",
		IOConfig: &types.IOConfig{
			OpenStdin: true,
			Tty:       true,
		},
	}

	tests := map[string]struct {
		attachConfig *streams.AttachConfig
		mockExec     func() error
	}{
		"test_containerIO_nil": {
			attachConfig: attachConfig,
			mockExec: func() error {
				mockIoMgr.EXPECT().GetIO(testCtr.ID).Return(nil)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, log.NewError("failed to initialise IO for container ID = test-id"))
				return log.NewError("failed to initialise IO for container ID = test-id")
			},
		},
		"test_container_stdin_true": {
			attachConfig: attachConfig,
			mockExec: func() error {
				errChan := make(chan error)
				go func() {
					errChan <- nil
					close(errChan)
				}()

				mockIoMgr.EXPECT().GetIO(testCtr.ID).Return(mockIO)
				mockIO.EXPECT().Stream().Return(mockStream)
				mockReadCloser.EXPECT().Read(gomock.Any()).DoAndReturn(func(p []byte) (int, error) {
					return -1, io.EOF
				}).AnyTimes()
				mockStream.EXPECT().Attach(ctx, gomock.AssignableToTypeOf(attachConfig)).Return(errChan)

				return nil
			},
		},
		"test_container_stdin_false": {
			attachConfig: &streams.AttachConfig{Stdin: mockReadCloser},
			mockExec: func() error {
				errChan := make(chan error)
				go func() {
					errChan <- nil
					close(errChan)
				}()

				mockIoMgr.EXPECT().GetIO(testCtr.ID).Return(mockIO)
				mockIO.EXPECT().Stream().Return(mockStream)
				mockStream.EXPECT().Attach(ctx, gomock.AssignableToTypeOf(attachConfig)).Return(errChan)

				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertError(t, testCase.mockExec(), testClient.AttachContainer(ctx, testCtr, testCase.attachConfig))
		})
	}
}

func TestPauseContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockTask := containerdMocks.NewMockTask(mockCtrl)

	testClient := &containerdClient{
		ctrdCache: &containerInfoCache{
			cache: map[string]*containerInfo{
				testContainerID: {
					c: &types.Container{
						ID: testContainerID,
					},
					task: mockTask,
				},
			},
		},
	}
	ctx := context.Background()

	tests := map[string]struct {
		arg      *types.Container
		mockExec func(context context.Context, mockTask *containerdMocks.MockTask) error
	}{
		"test_error_missing_container_to_pause": {
			arg: &types.Container{
				ID: "test-container",
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				return log.NewErrorf("missing container to pause")
			},
		},
		"test_pause_container_with_error": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				err := log.NewErrorf("test pause task error")
				mockTask.EXPECT().Pause(context).Return(err)
				return err
			},
		},
		"test_pause_container_without_error": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				mockTask.EXPECT().Pause(context).Return(nil)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertError(t, testCase.mockExec(ctx, mockTask), testClient.PauseContainer(ctx, testCase.arg))
		})
	}
}

func TestUnpauseContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockTask := containerdMocks.NewMockTask(mockCtrl)

	testClient := &containerdClient{
		ctrdCache: &containerInfoCache{
			cache: map[string]*containerInfo{
				testContainerID: {
					c: &types.Container{
						ID: testContainerID,
					},
					task: mockTask,
				},
			},
		},
	}
	ctx := context.Background()

	tests := map[string]struct {
		arg      *types.Container
		mockExec func(context context.Context, mockTask *containerdMocks.MockTask) error
	}{
		"test_error_missing_container_to_unpause": {
			arg: &types.Container{
				ID: "test-container",
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				return log.NewErrorf("missing container to unpause")
			},
		},
		"test_unpause_container_with_error": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				err := log.NewErrorf("test unpause task error")
				mockTask.EXPECT().Resume(context).Return(err)
				return err
			},
		},
		"test_unpause_container_without_error": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) error {
				mockTask.EXPECT().Resume(context).Return(nil)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertError(t, testCase.mockExec(ctx, mockTask), testClient.UnpauseContainer(ctx, testCase.arg))
		})
	}
}

func TestListContainers(t *testing.T) {
	tests := map[string]struct {
		testClient *containerdClient
		want       []*types.Container
	}{
		"test_without_containers": {
			testClient: &containerdClient{
				ctrdCache: &containerInfoCache{},
			},
			want: make([]*types.Container, 0),
		},
		"test_with_containers": {
			testClient: &containerdClient{
				ctrdCache: &containerInfoCache{
					cache: map[string]*containerInfo{
						testContainerID: {
							c: &types.Container{
								ID: testContainerID,
							},
						},
					},
				},
			},
			want: []*types.Container{
				{
					ID: testContainerID,
				},
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			got, _ := testCase.testClient.ListContainers(context.Background())
			testutil.AssertEqual(t, got, testCase.want)
		})
	}
}

func TestGetContainerInfo(t *testing.T) {
	testClient := &containerdClient{
		ctrdCache: &containerInfoCache{
			cache: map[string]*containerInfo{
				testContainerID: {
					c: &types.Container{
						ID: testContainerID,
					},
				},
			},
		},
	}

	tests := map[string]struct {
		arg       string
		want      *types.Container
		wantError error
	}{
		"test_existing_container": {
			arg: testContainerID,
			want: &types.Container{
				ID: testContainerID,
			},
			wantError: nil,
		},
		"test_not_existing_container": {
			arg:       "not-existing-container-id",
			want:      nil,
			wantError: log.NewErrorf("missing container with ID = not-existing-container-id"),
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			got, err := testClient.GetContainerInfo(context.Background(), testCase.arg)
			if testCase.wantError != nil {
				testutil.AssertEqual(t, testCase.want, got)
			} else {
				testutil.AssertError(t, testCase.wantError, err)
			}
		})
	}
}

func TestRestoreContainer(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockSpi := ctrdMocks.NewMockcontainerdSpi(mockCtrl)
	mockIoMgr := NewMockcontainerIOManager(mockCtrl)
	mockLogMgr := ctrdMocks.NewMockcontainerLogsManager(mockCtrl)
	mockLogDriver := loggerMocks.NewMockLogDriver(mockCtrl)
	mockContainer := containerdMocks.NewMockContainer(mockCtrl)
	mockTask := containerdMocks.NewMockTask(mockCtrl)

	testClient := &containerdClient{
		ctrdCache: newContainerInfoCache(),
		ioMgr:     mockIoMgr,
		logsMgr:   mockLogMgr,
		spi:       mockSpi,
	}
	ctx := context.Background()
	testCtr := &types.Container{
		ID: testContainerID,
		HostConfig: &types.HostConfig{
			LogConfig: &types.LogConfiguration{
				ModeConfig: &types.LogModeConfiguration{},
			},
		},
		IOConfig: &types.IOConfig{
			OpenStdin: false,
		},
	}

	tests := map[string]struct {
		mockExec func() error
	}{
		"test_error_snapshot_not_exist": {
			mockExec: func() error {
				err := log.NewErrorf("snapshot for container with ID = %s does not exist", testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, err)
				return err
			},
		},
		"test_error_retrieve_container": {
			mockExec: func() error {
				err := log.NewErrorf("failed to retrieve container ID = %s from containerd while restoring", testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil).Return(nil, err)
				return err
			},
		},
		"test_error_initialising IO": {
			mockExec: func() error {
				err := log.NewErrorf("error while initialising IO for container ID = %s", testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(false)
				mockIoMgr.EXPECT().InitIO(testCtr.ID, testCtr.IOConfig.OpenStdin).Return(nil, err)
				return err
			},
		},
		"test_error_init_log_driver": {
			mockExec: func() error {
				err := log.NewErrorf("failed to initialize logger for container ID = %s", testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(true)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(nil, err)
				return err
			},
		},
		"test_error_load_task": {
			mockExec: func() error {
				err := log.NewErrorf("error loading task for container ID = %s while restoring", testCtr.ID)
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(true)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				mockIoMgr.EXPECT().NewCioAttach(testCtr.ID)
				mockSpi.EXPECT().LoadTask(gomock.Any(), mockContainer, gomock.Any()).Return(nil, err)
				mockContainer.EXPECT().Delete(ctx)
				return err
			},
		},
		"test_restore_container_without_error": {
			mockExec: func() error {
				mockSpi.EXPECT().GetSnapshot(ctx, testCtr.ID).Return(snapshots.Info{}, nil)
				mockSpi.EXPECT().LoadContainer(ctx, testCtr.ID).Return(nil, nil).Return(mockContainer, nil)
				mockIoMgr.EXPECT().ExistsIO(testCtr.ID).Return(true)
				mockLogMgr.EXPECT().GetLogDriver(testCtr).Return(mockLogDriver, nil)
				mockIoMgr.EXPECT().ConfigureIO(testCtr.ID, mockLogDriver, testCtr.HostConfig.LogConfig.ModeConfig).Return(nil)
				mockIoMgr.EXPECT().NewCioAttach(testCtr.ID)
				mockSpi.EXPECT().LoadTask(gomock.Any(), mockContainer, gomock.Any()).Return(mockTask, nil)
				mockContainer.EXPECT().ID().Return(testCtr.ID)
				resChan := make(<-chan containerd.ExitStatus)
				mockTask.EXPECT().Wait(ctx).Return(resChan, nil)
				return nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			testutil.AssertError(t, testCase.mockExec(), testClient.RestoreContainer(ctx, testCtr))
		})
	}
}

func TestReleaseContainerResources(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockIoMgr := NewMockcontainerIOManager(mockCtrl)

	testClient := &containerdClient{
		ioMgr: mockIoMgr,
	}

	testCtr := &types.Container{
		ID: "test-id",
	}

	mockIoMgr.EXPECT().ResetIO(testCtr.ID).MaxTimes(1)

	testClient.ReleaseContainerResources(context.Background(), testCtr)
}

func TestSetContainerExitHooks(t *testing.T) {
	testClient := &containerdClient{
		ctrdCache: &containerInfoCache{},
	}

	arg := func(container *types.Container, code int64, err error, oomKilled bool, cleanup func() error) error {
		return nil
	}

	testClient.SetContainerExitHooks(arg)

	testutil.AssertEqual(t, 1, len(testClient.ctrdCache.containerExitHooks))

	if reflect.ValueOf(testClient.ctrdCache.containerExitHooks[0]).Pointer() != reflect.ValueOf(arg).Pointer() {
		t.Errorf("SetContainerExitHooks() = %v, want %v", reflect.ValueOf(testClient.ctrdCache.containerExitHooks[0]), reflect.ValueOf(arg))
	}
}

func TestCtrdClientDispose(t *testing.T) {
	testCases := map[string]struct {
		imagesExpiryDisabled bool
		mockExec             func(mockCtrdWrapper *ctrdMocks.MockcontainerdSpi, mockResMgr *MockresourcesWatcher) error
	}{
		"test_no_err_expiry_enabled": {
			imagesExpiryDisabled: false,
			mockExec: func(mockCtrdWrapper *ctrdMocks.MockcontainerdSpi, mockResMgr *MockresourcesWatcher) error {
				mockCtrdWrapper.EXPECT().Dispose(gomock.Any()).Return(nil)
				mockResMgr.EXPECT().Dispose()
				return nil
			},
		},
		"test_err_expiry_enabled": {
			imagesExpiryDisabled: false,
			mockExec: func(mockCtrdWrapper *ctrdMocks.MockcontainerdSpi, mockResMgr *MockresourcesWatcher) error {
				err := log.NewError("test error")
				mockCtrdWrapper.EXPECT().Dispose(gomock.Any()).Return(err)
				mockResMgr.EXPECT().Dispose()
				return err
			},
		},
		"test_no_err_expiry_disabled": {
			imagesExpiryDisabled: true,
			mockExec: func(mockCtrdWrapper *ctrdMocks.MockcontainerdSpi, mockResMgr *MockresourcesWatcher) error {
				mockCtrdWrapper.EXPECT().Dispose(gomock.Any()).Return(nil)
				mockResMgr.EXPECT().Dispose().Times(0)
				return nil
			},
		},
		"test_err_expiry_disabled": {
			imagesExpiryDisabled: true,
			mockExec: func(mockCtrdWrapper *ctrdMocks.MockcontainerdSpi, mockResMgr *MockresourcesWatcher) error {
				err := log.NewError("test error")
				mockCtrdWrapper.EXPECT().Dispose(gomock.Any()).Return(err)
				mockResMgr.EXPECT().Dispose().Times(0)
				return err
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			// init mock ctrl
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			// init mocks
			mockSpi := ctrdMocks.NewMockcontainerdSpi(mockCtrl)
			mockResMgr := NewMockresourcesWatcher(mockCtrl)
			// mock exec
			expectedErr := testData.mockExec(mockSpi, mockResMgr)
			// init spi under test
			testClient := &containerdClient{
				spi:                mockSpi,
				ctrdCache:          newContainerInfoCache(),
				imageExpiryDisable: testData.imagesExpiryDisabled,
			}
			if !testData.imagesExpiryDisabled {
				testClient.imagesWatcher = mockResMgr
			}
			// test
			actualErr := testClient.Dispose(context.Background())
			testutil.AssertTrue(t, testClient.ctrdCache.isContainerdDead())
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestCtrdClientUpdateContainer(t *testing.T) {
	var (
		mockContainer *containerdMocks.MockContainer
		mockTask      *containerdMocks.MockTask
	)

	ctx := namespaces.WithNamespace(context.Background(), "test")
	spec, _ := oci.GenerateSpec(ctx, nil, &containers.Container{}) // populates default unix spec
	testCtrID := "test-update-id"
	unlimited := int64(-1)

	var testClient *containerdClient
	tests := map[string]struct {
		ctr       *types.Container
		resources *types.Resources
		mockExec  func() error
	}{
		"test_with_initial_limits": {
			ctr: &types.Container{
				ID: testCtrID,
				HostConfig: &types.HostConfig{
					Resources: &types.Resources{
						Memory:     "200M",
						MemorySwap: "500M",
					},
				},
			},
			resources: &types.Resources{
				Memory: "200M",
			},
			mockExec: func() error {
				limit := int64(200 * 1024 * 1024)
				resources := &specs.LinuxResources{
					Devices: spec.Linux.Resources.Devices,
					Memory: &specs.LinuxMemory{
						Limit: &limit,
						Swap:  &unlimited,
					},
				}
				mockContainer.EXPECT().Spec(ctx).Return(spec, nil)
				mockTask.EXPECT().Update(ctx, matchers.MatchesUpdateTaskOpts(containerd.WithResources(resources))).Return(nil)
				return nil
			},
		},
		"test_with_initially_missing_limits": {
			ctr: &types.Container{
				ID:         testCtrID,
				HostConfig: &types.HostConfig{},
			},
			resources: &types.Resources{
				Memory:            "200M",
				MemoryReservation: "100M",
			},
			mockExec: func() error {
				limit := int64(200 * 1024 * 1024)
				reservation := int64(100 * 1024 * 1024)
				resources := &specs.LinuxResources{
					Devices: spec.Linux.Resources.Devices,
					Memory: &specs.LinuxMemory{
						Limit:       &limit,
						Reservation: &reservation,
					},
				}
				mockContainer.EXPECT().Spec(ctx).Return(spec, nil)
				mockTask.EXPECT().Update(ctx, matchers.MatchesUpdateTaskOpts(containerd.WithResources(resources))).Return(nil)
				return nil
			},
		},
		"test_no_error_nil_resources": {
			ctr: &types.Container{
				ID:         testCtrID,
				HostConfig: &types.HostConfig{},
			},
			mockExec: func() error {
				return nil
			},
		},
		"test_get_spec_err": {
			ctr: &types.Container{
				ID: testCtrID,
			},
			resources: &types.Resources{},
			mockExec: func() error {
				err := log.NewError("test error")
				mockContainer.EXPECT().Spec(ctx).Return(nil, err)
				return err
			},
		},
		"test_missing_ctr_err": {
			ctr: &types.Container{
				ID: testCtrID,
			},
			resources: &types.Resources{},
			mockExec: func() error {
				testClient.ctrdCache.cache[testCtrID] = nil
				return log.NewErrorf("missing container to update with ID = %s", testCtrID)
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {

			// init mock ctrl
			mockCtrl := gomock.NewController(t)

			// init mocks
			mockContainer = containerdMocks.NewMockContainer(mockCtrl)
			mockTask = containerdMocks.NewMockTask(mockCtrl)
			defer mockCtrl.Finish()

			testClient = &containerdClient{
				ctrdCache: newContainerInfoCache(),
			}
			testClient.ctrdCache.cache[testCase.ctr.ID] = &containerInfo{
				container: mockContainer,
				task:      mockTask,
			}

			testutil.AssertError(t, testCase.mockExec(), testClient.UpdateContainer(ctx, testCase.ctr, testCase.resources))
		})
	}
}

func TestGetContainerStats(t *testing.T) {

	tests := map[string]struct {
		arg      *types.Container
		mockExec func(context context.Context, mockTask *containerdMocks.MockTask) (*types.CPUStats, *types.MemoryStats, *types.IOStats, uint64, time.Time, error)
	}{
		"test_container_not_exists": {
			arg: &types.Container{
				ID: "non-existing-test-container-id",
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) (*types.CPUStats, *types.MemoryStats, *types.IOStats, uint64, time.Time, error) {
				return nil, nil, nil, 0, time.Time{}, log.NewErrorf("missing container with ID = non-existing-test-container-id")
			},
		},
		"test_metrics_error": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) (*types.CPUStats, *types.MemoryStats, *types.IOStats, uint64, time.Time, error) {
				err := log.NewErrorf("metrics error")
				mockTask.EXPECT().Metrics(context).Return(nil, err)
				return nil, nil, nil, 0, time.Time{}, err
			},
		},
		"test_metrics_invalid_type": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) (*types.CPUStats, *types.MemoryStats, *types.IOStats, uint64, time.Time, error) {
				invalidMetricsObj := &statsV1.MemoryStat{
					Usage: &statsV1.MemoryEntry{},
				}
				b, mErr := typeurl.MarshalAny(invalidMetricsObj)
				testutil.AssertNil(t, mErr)
				invalidMetrics := &containerdtypes.Metric{
					Data: b,
				}
				err := log.NewErrorf("unexpected metrics type = %T for container with ID = %s", invalidMetricsObj, testContainerID)
				mockTask.EXPECT().Metrics(context).Return(invalidMetrics, nil)
				return nil, nil, nil, 0, time.Time{}, err
			},
		},
		"test_metrics": {
			arg: &types.Container{
				ID: testContainerID,
			},
			mockExec: func(context context.Context, mockTask *containerdMocks.MockTask) (*types.CPUStats, *types.MemoryStats, *types.IOStats, uint64, time.Time, error) {
				ctrdMetrics := &statsV1.Metrics{
					Blkio: &statsV1.BlkIOStat{
						IoServiceBytesRecursive: []*statsV1.BlkIOEntry{
							{
								Op:    "read",
								Value: 1,
							},
							{
								Op:    "read",
								Value: 1,
							},
							{
								Op:    "write",
								Value: 1,
							},
							{
								Op:    "write",
								Value: 11,
							},
						},
					},
					Pids: &statsV1.PidsStat{
						Current: 11,
					},
					CPU: &statsV1.CPUStat{
						Usage: &statsV1.CPUUsage{
							Total: 1,
						},
					},
					Memory: &statsV1.MemoryStat{
						Usage: &statsV1.MemoryEntry{
							Usage: 11,
						},
						TotalInactiveFile: 1,
					},
				}

				eBytes, marshalErr := typeurl.MarshalAny(ctrdMetrics)
				testutil.AssertNil(t, marshalErr)

				ctrdMetricsRaw := &containerdtypes.Metric{
					Data:      eBytes,
					Timestamp: time.Now(),
				}
				mockTask.EXPECT().Metrics(context).Return(ctrdMetricsRaw, nil)
				return &types.CPUStats{Used: ctrdMetrics.CPU.Usage.Total}, &types.MemoryStats{Used: ctrdMetrics.Memory.Usage.Usage - ctrdMetrics.Memory.TotalInactiveFile}, &types.IOStats{Read: 2, Write: 12}, ctrdMetrics.Pids.Current, ctrdMetricsRaw.Timestamp, nil
			},
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockTask := containerdMocks.NewMockTask(mockCtrl)

			testClient := &containerdClient{
				ctrdCache: &containerInfoCache{
					cache: map[string]*containerInfo{
						testContainerID: {
							c: &types.Container{
								ID: testContainerID,
							},
							task: mockTask,
						},
					},
				},
			}

			ctx := context.Background()

			cpuStats, memStats, ioStats, pidStats, timestamp, expectedError := testCase.mockExec(ctx, mockTask)
			cpu, mem, io, pids, tstamp, err := testClient.GetContainerStats(ctx, testCase.arg)
			testutil.AssertError(t, expectedError, err)
			if expectedError != nil {
				testutil.AssertNil(t, cpu)
				testutil.AssertNil(t, mem)
				testutil.AssertNil(t, io)
				testutil.AssertEqual(t, uint64(0), pids)
				testutil.AssertEqual(t, time.Time{}, tstamp)
			} else {
				testutil.AssertNotNil(t, cpu)
				testutil.AssertEqual(t, cpuStats.Used, cpu.Used)
				testutil.AssertNotNil(t, mem)
				testutil.AssertEqual(t, memStats.Used, mem.Used)
				testutil.AssertNotNil(t, io)
				testutil.AssertEqual(t, ioStats.Read, io.Read)
				testutil.AssertEqual(t, ioStats.Write, io.Write)
				testutil.AssertEqual(t, pidStats, pids)
				testutil.AssertEqual(t, timestamp, tstamp)
			}
		})
	}
}
