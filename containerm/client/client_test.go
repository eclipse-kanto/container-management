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

package client

import (
	"context"
	"errors"
	"io"
	"testing"

	pbcontainers "github.com/eclipse-kanto/container-management/containerm/api/services/containers"
	"github.com/eclipse-kanto/container-management/containerm/api/services/sysinfo"
	"github.com/eclipse-kanto/container-management/containerm/api/types/containers"
	typesSysInfo "github.com/eclipse-kanto/container-management/containerm/api/types/sysinfo"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	mockscontainerspb "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/api/services/containers"
	mockssysinfopb "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/api/services/sysinfo"
	sysinfotypes "github.com/eclipse-kanto/container-management/containerm/sysinfo/types"
	"github.com/eclipse-kanto/container-management/containerm/util/protobuf"
	"github.com/golang/mock/gomock"
	"github.com/golang/protobuf/ptypes/empty"
)

const (
	containerID       = "test-ctr"
	containerImageID  = "host/group/image:tag"
	containerName     = "test-ctr-name"
	containerID2      = "test-ctr-2"
	containerImageID2 = "host/group/image2:tag"
	containerName2    = "test-ctr-name-2"
)

var (
	mockContainersClient *mockscontainerspb.MockContainersClient
	mockAttchClient      *mockscontainerspb.MockContainers_AttachClient
	mockSysInfoClient    *mockssysinfopb.MockSystemInfoClient

	testClient Client

	testCtx context.Context

	testBytes = []byte{0x01, 0x02}
)

func setup(controller *gomock.Controller) {
	mockContainersClient = mockscontainerspb.NewMockContainersClient(controller)
	mockAttchClient = mockscontainerspb.NewMockContainers_AttachClient(controller)
	mockSysInfoClient = mockssysinfopb.NewMockSystemInfoClient(controller)
	testClient = &client{
		grpcContainersClient: mockContainersClient,
		grpcSystemInfoClient: mockSysInfoClient,
	}
	testCtx = context.Background()
}

type testCreateArgs struct {
	ctx context.Context
	ctr *types.Container
}
type mockExecCreate func(args testCreateArgs) error

func TestCreate(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testCreateArgs
		mockExecution mockExecCreate
	}{
		"test_create_no_errs": {
			args: testCreateArgs{
				ctx: testCtx,
				ctr: &types.Container{
					ID:    containerID,
					Image: types.Image{Name: containerImageID},
				},
			},
			mockExecution: mockExecCreateNoErrors,
		},
		"test_create_errs": {
			args: testCreateArgs{
				ctx: testCtx,
				ctr: &types.Container{
					ID:    containerID,
					Image: types.Image{Name: containerImageID},
				},
			},
			mockExecution: mockExecCreateErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(testCase.args)

			_, resultErr := testClient.Create(testCase.args.ctx, testCase.args.ctr)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}

}

type testGetArgs struct {
	ctx   context.Context
	ctrID string
}
type mockExecGet func(args testGetArgs) (*types.Container, error)

func TestGet(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testGetArgs
		mockExecution mockExecGet
	}{
		"test_get_no_errs": {
			args: testGetArgs{
				ctx:   testCtx,
				ctrID: containerID,
			},
			mockExecution: mockExecGetNoErrors,
		},
		"test_get_errs": {
			args: testGetArgs{
				ctx:   testCtx,
				ctrID: containerID,
			},
			mockExecution: mockExecGetErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedCtr, expectedRunErr := testCase.mockExecution(testCase.args)

			ctr, resultErr := testClient.Get(testCase.args.ctx, testCase.args.ctrID)
			// assert container
			testutil.AssertEqual(t, expectedCtr, ctr)
			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testAttachArgs struct {
	ctx   context.Context
	id    string
	stdin bool
}
type mockExecAttach func(args testAttachArgs) (io.Writer, io.ReadCloser, error)

func TestAttach(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testAttachArgs
		mockExecution mockExecAttach
	}{
		"test_attach_no_errs": {
			args: testAttachArgs{
				ctx:   testCtx,
				id:    containerID,
				stdin: false,
			},
			mockExecution: mockExecAttachNoErrors,
		},
		"test_attach_err": {
			args: testAttachArgs{
				ctx:   testCtx,
				id:    containerID,
				stdin: false,
			},
			mockExecution: mockExecAttachError,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedWriter, expectedReader, expectedRunErr := testCase.mockExecution(testCase.args)

			writer, reader, resultErr := testClient.Attach(testCase.args.ctx, testCase.args.id, testCase.args.stdin)

			// assert container
			testutil.AssertEqual(t, expectedWriter, writer)
			testutil.AssertEqual(t, expectedReader, reader)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testListArgs struct {
	ctx     context.Context
	filters []Filter
}
type mockExecList func(args testListArgs) ([]*types.Container, error)

func TestList(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)
	tests := map[string]struct {
		args          testListArgs
		mockExecution mockExecList
	}{
		"test_list_no_filter_no_err": {
			args: testListArgs{
				ctx: testCtx,
			},
			mockExecution: mockExecListNoFilterNoErrors,
		},
		"test_list_filter_no_err": {
			args: testListArgs{
				ctx:     testCtx,
				filters: []Filter{WithName(containerName2)},
			},
			mockExecution: mockExecListFilterNoErrors,
		},
		"test_list_no_containers_no_err": {
			args: testListArgs{
				ctx: testCtx,
			},
			mockExecution: mockExecListNoContainersNoErrors,
		},
		"test_list_err": {
			args: testListArgs{
				ctx: testCtx,
			},
			mockExecution: mockExecListErrors,
		},
	}

	//execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedCtrs, expectedRunErr := testCase.mockExecution(testCase.args)

			ctrs, resultErr := testClient.List(testCase.args.ctx, testCase.args.filters...)

			// assert container
			testutil.AssertEqual(t, expectedCtrs, ctrs)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testStartArgs struct {
	ctx context.Context
	id  string
}
type mockExecStart func(args testStartArgs) error

func TestStart(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testStartArgs
		mockExecution mockExecStart
	}{
		"test_start_no_errs": {
			args: testStartArgs{
				ctx: testCtx,
				id:  containerID,
			},
			mockExecution: mockExecStartNoErrors,
		},
		"test_start_errs": {
			args: testStartArgs{
				ctx: testCtx,
				id:  containerID,
			},
			mockExecution: mockExecStartErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(testCase.args)

			resultErr := testClient.Start(testCase.args.ctx, testCase.args.id)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testStopArgs struct {
	ctx      context.Context
	id       string
	stopOpts *types.StopOpts
}
type mockExecStop func(args testStopArgs) error

func TestStop(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testStopArgs
		mockExecution mockExecStop
	}{
		"test_stop_no_errs": {
			args: testStopArgs{
				ctx: testCtx,
				id:  containerID,
				stopOpts: &types.StopOpts{
					Timeout: 5,
					Force:   false,
				},
			},
			mockExecution: mockExecStopNoErrors,
		},
		"test_stop_errs": {
			args: testStopArgs{
				ctx: testCtx,
				id:  containerID,
				stopOpts: &types.StopOpts{
					Timeout: 5,
					Force:   false,
				},
			},
			mockExecution: mockExecStopErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(testCase.args)

			resultErr := testClient.Stop(testCase.args.ctx, testCase.args.id, testCase.args.stopOpts)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testRestartArgs struct {
	ctx     context.Context
	id      string
	timeout int64
}
type mockExecRestart func(args testRestartArgs) error

func TestRestart(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testRestartArgs
		mockExecution mockExecRestart
	}{
		"test_restart_no_errs": {
			args: testRestartArgs{
				ctx:     testCtx,
				id:      containerID,
				timeout: 5,
			},
			mockExecution: mockExecRestartNoErrors,
		},
		"test_restart_errs": {
			args: testRestartArgs{
				ctx:     testCtx,
				id:      containerID,
				timeout: 5,
			},
			mockExecution: mockExecRestartErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(testCase.args)

			resultErr := testClient.Restart(testCase.args.ctx, testCase.args.id, testCase.args.timeout)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testPauseArgs struct {
	ctx context.Context
	id  string
}
type mockExecPause func(args testPauseArgs) error

func TestPause(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testPauseArgs
		mockExecution mockExecPause
	}{
		"test_pause_no_errs": {
			args: testPauseArgs{
				ctx: testCtx,
				id:  containerID,
			},
			mockExecution: mockExecPauseNoErrors,
		},
		"test_pause_errs": {
			args: testPauseArgs{
				ctx: testCtx,
				id:  containerID,
			},
			mockExecution: mockExecPauseErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(testCase.args)

			resultErr := testClient.Pause(testCase.args.ctx, testCase.args.id)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testResumeArgs struct {
	ctx context.Context
	id  string
}
type mockExecResume func(args testResumeArgs) error

func TestResume(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testResumeArgs
		mockExecution mockExecResume
	}{
		"test_resume_no_errs": {
			args: testResumeArgs{
				ctx: testCtx,
				id:  containerID,
			},
			mockExecution: mockExecResumeNoErrors,
		},
		"test_resume_errs": {
			args: testResumeArgs{
				ctx: testCtx,
				id:  containerID,
			},
			mockExecution: mockExecResumeErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(testCase.args)

			resultErr := testClient.Resume(testCase.args.ctx, testCase.args.id)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testRenameArgs struct {
	ctx  context.Context
	id   string
	name string
}
type mockExecRename func(args testRenameArgs) error

func TestRename(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testRenameArgs
		mockExecution mockExecRename
	}{
		"test_rename_no_errs": {
			args: testRenameArgs{
				ctx:  testCtx,
				id:   containerID,
				name: containerName2,
			},
			mockExecution: mockExecRenameNoErrors,
		},
		"test_rename_errs": {
			args: testRenameArgs{
				ctx:  testCtx,
				id:   containerID,
				name: containerName2,
			},
			mockExecution: mockExecRenameErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(testCase.args)

			resultErr := testClient.Rename(testCase.args.ctx, testCase.args.id, testCase.args.name)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testUpdateArgs struct {
	ctx  context.Context
	id   string
	opts *types.UpdateOpts
}
type mockExecUpdate func(args testUpdateArgs) error

func TestUpdate(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testUpdateArgs
		mockExecution mockExecUpdate
	}{
		"test_update_no_errs": {
			args: testUpdateArgs{
				ctx:  testCtx,
				id:   containerID,
				opts: &types.UpdateOpts{},
			},
			mockExecution: mockExecUpdateNoErrors,
		},
		"test_update_errs": {
			args: testUpdateArgs{
				ctx:  testCtx,
				id:   containerID,
				opts: &types.UpdateOpts{},
			},
			mockExecution: mockExecUpdateErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(testCase.args)

			resultErr := testClient.Update(testCase.args.ctx, testCase.args.id, testCase.args.opts)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testRemoveArgs struct {
	ctx      context.Context
	id       string
	force    bool
	stopOpts *types.StopOpts
}
type mockExecRemove func(args testRemoveArgs) error

func TestRemove(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testRemoveArgs
		mockExecution mockExecRemove
	}{
		"test_remove_no_errs": {
			args: testRemoveArgs{
				ctx:   testCtx,
				force: true,
			},
			mockExecution: mockExecRemoveNoErrors,
		},
		"test_remove_errs": {
			args: testRemoveArgs{
				ctx:   testCtx,
				force: true,
			},
			mockExecution: mockExecRemoveErrors,
		},
		"test_remove_stop_opts": {
			args: testRemoveArgs{
				ctx:      testCtx,
				force:    true,
				stopOpts: &types.StopOpts{Timeout: 10},
			},
			mockExecution: mockExecRemoveStopOpts,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedRunErr := testCase.mockExecution(testCase.args)

			resultErr := testClient.Remove(testCase.args.ctx, testCase.args.id, testCase.args.force, testCase.args.stopOpts)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

type testProjectInfoArgs struct {
	ctx context.Context
}
type mockExecProjectInfo func(args testProjectInfoArgs) (sysinfotypes.ProjectInfo, error)

func TestProjectInfo(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)

	tests := map[string]struct {
		args          testProjectInfoArgs
		mockExecution mockExecProjectInfo
	}{
		"test_project_info_no_errs": {
			args: testProjectInfoArgs{
				ctx: testCtx,
			},
			mockExecution: mockExecProjectInfoNoErrors,
		},
		"test_project_info_errs": {
			args: testProjectInfoArgs{
				ctx: testCtx,
			},
			mockExecution: mockExecProjectInfoErrors,
		},
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expProjectInfo, expectedRunErr := testCase.mockExecution(testCase.args)

			resultProjectInfo, resultErr := testClient.ProjectInfo(testCase.args.ctx)

			testutil.AssertEqual(t, expProjectInfo, resultProjectInfo)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// Tests for client_io_util
type testWriteArgs struct {
	data   []byte
	writer *Writer
}
type mockExecWrite func(args testWriteArgs) (int, error)

func TestWriteNoErr(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	setup(controller)
	writer := &Writer{
		ctx:         testCtx,
		writeClient: mockAttchClient,
		containerID: containerID,
		stdIn:       false,
		offset:      0,
		err:         nil,
	}

	tests := map[string]struct {
		args          testWriteArgs
		mockExecution mockExecWrite
	}{
		"test_write_no_errs": {
			args: testWriteArgs{
				data:   testBytes,
				writer: writer,
			},
			mockExecution: mockExecWriteNoErrors,
		},
		// Causes unpredicted test behavior when, as it changes the internal states of the writers
		/*		"test_write_errs": {
				args: testWriteArgs{
					data: testBytes,
					writer: writer,
				},
				mockExecution: mockExecWriteErrors,
			},*/
	}

	// execute tests
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			expectedWriteResult, expectedRunErr := testCase.mockExecution(testCase.args)

			writeResult, resultErr := writer.Write(testBytes)

			// assert container
			testutil.AssertEqual(t, expectedWriteResult, writeResult)

			testutil.AssertError(t, expectedRunErr, resultErr)
		})
	}
}

// Mock executions -------------------------------------------------------------
// Create -------------------------------------------------------------
func mockExecCreateNoErrors(args testCreateArgs) error {
	pbResponse := &pbcontainers.CreateContainerResponse{}
	mockContainersClient.EXPECT().Create(args.ctx, gomock.Eq(&pbcontainers.CreateContainerRequest{Container: protobuf.ToProtoContainer(args.ctr)})).Times(1).Return(pbResponse, nil)
	return nil
}

func mockExecCreateErrors(args testCreateArgs) error {
	err := errors.New("failed to create container")
	mockContainersClient.EXPECT().Create(args.ctx, gomock.Eq(&pbcontainers.CreateContainerRequest{Container: protobuf.ToProtoContainer(args.ctr)})).Times(1).Return(nil, err)
	return err
}

// Get -------------------------------------------------------------
func mockExecGetNoErrors(args testGetArgs) (*types.Container, error) {
	ctr := types.Container{
		ID:    containerID,
		Name:  containerName,
		Image: types.Image{Name: containerImageID},
	}
	pbResponse := &pbcontainers.GetContainerResponse{
		Container: protobuf.ToProtoContainer(&ctr),
	}
	mockContainersClient.EXPECT().Get(args.ctx, gomock.Eq(&pbcontainers.GetContainerRequest{
		Id: args.ctrID,
	})).Times(1).Return(pbResponse, nil)
	return &ctr, nil
}

func mockExecGetErrors(args testGetArgs) (*types.Container, error) {
	err := errors.New("failed to get contianer")
	mockContainersClient.EXPECT().Get(args.ctx, gomock.Eq(&pbcontainers.GetContainerRequest{
		Id: args.ctrID,
	})).Times(1).Return(nil, err)
	return nil, err
}

// List -------------------------------------------------------------
func mockExecListNoFilterNoErrors(args testListArgs) ([]*types.Container, error) {
	ctr := types.Container{
		ID:    containerID,
		Name:  containerName,
		Image: types.Image{Name: containerImageID},
	}
	ctr2 := types.Container{
		ID:    containerID2,
		Name:  containerName2,
		Image: types.Image{Name: containerImageID2},
	}
	pbResponse := &pbcontainers.ListContainersResponse{
		Containers: []*containers.Container{protobuf.ToProtoContainer(&ctr), protobuf.ToProtoContainer(&ctr2)},
	}
	mockContainersClient.EXPECT().List(args.ctx, gomock.Eq(&pbcontainers.ListContainersRequest{})).Times(1).Return(pbResponse, nil)
	return []*types.Container{&ctr, &ctr2}, nil
}

func mockExecListFilterNoErrors(args testListArgs) ([]*types.Container, error) {
	ctr := types.Container{
		ID:    containerID,
		Name:  containerName,
		Image: types.Image{Name: containerImageID},
	}
	ctr2 := types.Container{
		ID:    containerID2,
		Name:  containerName2,
		Image: types.Image{Name: containerImageID2},
	}
	pbResponse := &pbcontainers.ListContainersResponse{
		Containers: []*containers.Container{protobuf.ToProtoContainer(&ctr), protobuf.ToProtoContainer(&ctr2)},
	}
	mockContainersClient.EXPECT().List(args.ctx, gomock.Eq(&pbcontainers.ListContainersRequest{})).Times(1).Return(pbResponse, nil)
	return []*types.Container{&ctr2}, nil
}

func mockExecListNoContainersNoErrors(args testListArgs) ([]*types.Container, error) {
	pbResponse := &pbcontainers.ListContainersResponse{
		Containers: nil,
	}
	mockContainersClient.EXPECT().List(args.ctx, gomock.Eq(&pbcontainers.ListContainersRequest{})).Times(1).Return(pbResponse, nil)
	return nil, nil
}

func mockExecListErrors(args testListArgs) ([]*types.Container, error) {
	err := errors.New("failed to list containers")
	mockContainersClient.EXPECT().List(args.ctx, gomock.Eq(&pbcontainers.ListContainersRequest{})).Times(1).Return(nil, err)
	return nil, err
}

// Start -------------------------------------------------------------
func mockExecStartNoErrors(args testStartArgs) error {
	mockContainersClient.EXPECT().Start(args.ctx, gomock.Eq(&pbcontainers.StartContainerRequest{Id: args.id})).Times(1).Return(nil, nil)
	return nil
}

func mockExecStartErrors(args testStartArgs) error {
	err := errors.New("failed to start")
	mockContainersClient.EXPECT().Start(args.ctx, gomock.Eq(&pbcontainers.StartContainerRequest{Id: args.id})).Times(1).Return(nil, err)
	return err
}

// Stop -------------------------------------------------------------
func mockExecStopNoErrors(args testStopArgs) error {
	mockContainersClient.EXPECT().Stop(args.ctx, gomock.Eq(&pbcontainers.StopContainerRequest{
		Id: args.id,
		StopOptions: &containers.StopOptions{
			Timeout: args.stopOpts.Timeout,
			Force:   args.stopOpts.Force,
		},
	})).Times(1).Return(nil, nil)
	return nil
}

func mockExecStopErrors(args testStopArgs) error {
	err := errors.New("failed to stop")
	mockContainersClient.EXPECT().Stop(args.ctx, gomock.Eq(&pbcontainers.StopContainerRequest{
		Id: args.id,
		StopOptions: &containers.StopOptions{
			Timeout: args.stopOpts.Timeout,
			Force:   args.stopOpts.Force,
		},
	})).Times(1).Return(nil, err)
	return err
}

// Attach -------------------------------------------------------------
func mockExecAttachNoErrors(args testAttachArgs) (io.Writer, io.ReadCloser, error) {
	mockContainersClient.EXPECT().Attach(args.ctx).Times(1).Return(mockAttchClient, nil)
	mockAttchClient.EXPECT().Send(gomock.Eq(&pbcontainers.AttachContainerRequest{
		Id:          args.id,
		StdIn:       args.stdin,
		DataToWrite: nil,
	})).Times(1).Return(nil)
	return &Writer{
			ctx:         args.ctx,
			writeClient: mockAttchClient,
			containerID: args.id,
			stdIn:       args.stdin,
		}, &Reader{
			ctx:         args.ctx,
			containerID: args.id,
			stdIn:       args.stdin,
			readClient:  mockAttchClient,
		}, nil
}
func mockExecAttachError(args testAttachArgs) (io.Writer, io.ReadCloser, error) {
	err := errors.New("failed to attach")
	mockContainersClient.EXPECT().Attach(args.ctx).Times(1).Return(nil, err)
	mockAttchClient.EXPECT().Send(gomock.Any()).Times(0)

	return nil, nil, err
}

// Restart -------------------------------------------------------------
func mockExecRestartNoErrors(args testRestartArgs) error {
	mockContainersClient.EXPECT().Restart(args.ctx, gomock.Eq(&pbcontainers.RestartContainerRequest{
		Id: args.id,
	})).Times(1).Return(nil, nil)
	return nil
}

func mockExecRestartErrors(args testRestartArgs) error {
	err := errors.New("failed to restart")
	mockContainersClient.EXPECT().Restart(args.ctx, gomock.Eq(&pbcontainers.RestartContainerRequest{
		Id: args.id,
	})).Times(1).Return(nil, err)
	return err
}

// Pause -------------------------------------------------------------
func mockExecPauseNoErrors(args testPauseArgs) error {
	mockContainersClient.EXPECT().Pause(args.ctx, gomock.Eq(&pbcontainers.PauseContainerRequest{
		Id: args.id,
	})).Times(1).Return(nil, nil)
	return nil
}

func mockExecPauseErrors(args testPauseArgs) error {
	err := errors.New("failed to pause")
	mockContainersClient.EXPECT().Pause(args.ctx, gomock.Eq(&pbcontainers.PauseContainerRequest{
		Id: args.id,
	})).Times(1).Return(nil, err)
	return err
}

// Resume -------------------------------------------------------------
func mockExecResumeNoErrors(args testResumeArgs) error {
	mockContainersClient.EXPECT().Unpause(args.ctx, gomock.Eq(&pbcontainers.UnpauseContainerRequest{
		Id: args.id,
	})).Times(1).Return(nil, nil)
	return nil
}

func mockExecResumeErrors(args testResumeArgs) error {
	err := errors.New("failed to resume")
	mockContainersClient.EXPECT().Unpause(args.ctx, gomock.Eq(&pbcontainers.UnpauseContainerRequest{
		Id: args.id,
	})).Times(1).Return(nil, err)
	return err
}

// Rename -------------------------------------------------------------
func mockExecRenameNoErrors(args testRenameArgs) error {
	mockContainersClient.EXPECT().Rename(args.ctx, gomock.Eq(&pbcontainers.RenameContainerRequest{
		Id:   args.id,
		Name: args.name,
	})).Times(1).Return(nil, nil)
	return nil
}

func mockExecRenameErrors(args testRenameArgs) error {
	err := errors.New("failed to rename")
	mockContainersClient.EXPECT().Rename(args.ctx, gomock.Eq(&pbcontainers.RenameContainerRequest{
		Id:   args.id,
		Name: args.name,
	})).Times(1).Return(nil, err)
	return err
}

// Update -------------------------------------------------------------
func mockExecUpdateNoErrors(args testUpdateArgs) error {
	mockContainersClient.EXPECT().Update(args.ctx, gomock.Eq(&pbcontainers.UpdateContainerRequest{
		Id:            args.id,
		UpdateOptions: protobuf.ToProtoUpdateOptions(args.opts),
	})).Times(1).Return(nil, nil)
	return nil
}

func mockExecUpdateErrors(args testUpdateArgs) error {
	err := errors.New("failed to update")
	mockContainersClient.EXPECT().Update(args.ctx, gomock.Eq(&pbcontainers.UpdateContainerRequest{
		Id:            args.id,
		UpdateOptions: protobuf.ToProtoUpdateOptions(args.opts),
	})).Times(1).Return(nil, err)
	return err
}

// Remove -------------------------------------------------------------
func mockExecRemoveNoErrors(args testRemoveArgs) error {
	mockContainersClient.EXPECT().Remove(args.ctx, gomock.Eq(&pbcontainers.RemoveContainerRequest{
		Id:    args.id,
		Force: args.force,
	})).Times(1).Return(nil, nil)
	return nil
}

func mockExecRemoveStopOpts(args testRemoveArgs) error {
	mockContainersClient.EXPECT().Remove(args.ctx, gomock.Eq(&pbcontainers.RemoveContainerRequest{
		Id:    args.id,
		Force: args.force,
		StopOptions: &containers.StopOptions{
			Timeout: args.stopOpts.Timeout,
			Force:   args.stopOpts.Force,
		},
	})).Times(1).Return(nil, nil)
	return nil
}

func mockExecRemoveErrors(args testRemoveArgs) error {
	err := errors.New("failed to remove")
	mockContainersClient.EXPECT().Remove(args.ctx, gomock.Eq(&pbcontainers.RemoveContainerRequest{
		Id:    args.id,
		Force: args.force,
	})).Times(1).Return(nil, err)
	return err
}

// ProjectInfo -------------------------------------------------------------
func mockExecProjectInfoNoErrors(args testProjectInfoArgs) (sysinfotypes.ProjectInfo, error) {
	pbresponse := &sysinfo.ProjectInfoResponse{
		ProjectInfo: &typesSysInfo.ProjectInfo{
			ProjectVersion: "v1.0.0",
		},
	}
	mockSysInfoClient.EXPECT().ProjectInfo(args.ctx, gomock.Eq(&empty.Empty{})).Times(1).Return(pbresponse, nil)
	return protobuf.ToInternalProjectInfo(pbresponse.ProjectInfo), nil
}

func mockExecProjectInfoErrors(args testProjectInfoArgs) (sysinfotypes.ProjectInfo, error) {
	err := errors.New("failed to get project info")
	mockSysInfoClient.EXPECT().ProjectInfo(args.ctx, gomock.Eq(&empty.Empty{})).Times(1).Return(nil, err)
	return sysinfotypes.ProjectInfo{}, err
}

// Write ------------------------------------------------------------

func mockExecWriteNoErrors(args testWriteArgs) (int, error) {
	n := 0
	for n < len(args.data) {
		mockAttchClient.EXPECT().Send(gomock.Any()).Times(1).Return(nil)
		bufSize := len(args.data) - n
		if bufSize > MaxBufSize {
			bufSize = MaxBufSize
		}
		n += bufSize
	}
	return n, nil
}

func mockExecWriteErrors(args testWriteArgs) (int, error) {
	err := errors.New("failed to write data")
	mockAttchClient.EXPECT().Send(gomock.Any()).Times(1).Return(err)
	return 0, err
}
