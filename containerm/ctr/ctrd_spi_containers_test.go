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

package ctr

import (
	"context"
	"reflect"
	"testing"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/leases"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	containerdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	ctrdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctrd"
	"github.com/golang/mock/gomock"
)

func TestCreateContainer(t *testing.T) {
	const (
		testExampleContainerID = "test.container.id"
		testLeaseID            = "test.lease.id"
	)

	testNewContainerOpt := func(context.Context, *containerd.Client, *containers.Container) error { return nil }

	testCases := map[string]struct {
		mapExec func(*ctrdMocks.MockcontainerClientWrapper, *containerdMocks.MockContainer) (containerd.Container, error)
	}{
		"test_no_error": {
			mapExec: func(clientWrapper *ctrdMocks.MockcontainerClientWrapper, container *containerdMocks.MockContainer) (containerd.Container, error) {
				clientWrapper.
					EXPECT().
					NewContainer(gomock.Any(), testContainerID, matchers.MatchesNewContainerOpts(testNewContainerOpt)).
					Times(1).
					Return(container, nil)
				return container, nil
			},
		},
		"test_error": {
			mapExec: func(clientWrapper *ctrdMocks.MockcontainerClientWrapper, _ *containerdMocks.MockContainer) (containerd.Container, error) {
				err := log.NewError("example new container error")
				clientWrapper.
					EXPECT().
					NewContainer(gomock.Any(), testContainerID, matchers.MatchesNewContainerOpts(testNewContainerOpt)).
					Times(1).
					Return(nil, err)
				return nil, err
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			clientMock := ctrdMocks.NewMockcontainerClientWrapper(ctrl)
			containerMock := containerdMocks.NewMockContainer(ctrl)
			expectedContainer, expectedErr := testCase.mapExec(clientMock, containerMock)

			testSpi := &ctrdSpi{
				client: clientMock,
				lease: &leases.Lease{
					ID: testLeaseID,
				},
			}
			actualContainer, actualErr := testSpi.CreateContainer(context.Background(), testContainerID, testNewContainerOpt)

			testutil.AssertEqual(t, expectedContainer, actualContainer)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestLoadContainer(t *testing.T) {
	const (
		testExampleContainerID = "test.container.id"
		testLeaseID            = "test.lease.id"
	)

	testCases := map[string]struct {
		mapExec func(*ctrdMocks.MockcontainerClientWrapper, *containerdMocks.MockContainer) (containerd.Container, error)
	}{
		"test_no_error": {
			mapExec: func(clientWrapper *ctrdMocks.MockcontainerClientWrapper, container *containerdMocks.MockContainer) (containerd.Container, error) {
				clientWrapper.
					EXPECT().
					LoadContainer(gomock.Any(), testContainerID).
					Times(1).
					Return(container, nil)
				return container, nil
			},
		},
		"test_error": {
			mapExec: func(clientWrapper *ctrdMocks.MockcontainerClientWrapper, _ *containerdMocks.MockContainer) (containerd.Container, error) {
				err := log.NewError("example new container error")
				clientWrapper.
					EXPECT().
					LoadContainer(gomock.Any(), testContainerID).
					Times(1).
					Return(nil, err)
				return nil, err
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			clientMock := ctrdMocks.NewMockcontainerClientWrapper(ctrl)
			containerMock := containerdMocks.NewMockContainer(ctrl)
			expectedContainer, expectedErr := testCase.mapExec(clientMock, containerMock)

			testSpi := &ctrdSpi{
				client: clientMock,
				lease: &leases.Lease{
					ID: testLeaseID,
				},
			}
			actualContainer, actualErr := testSpi.LoadContainer(context.Background(), testContainerID)

			testutil.AssertEqual(t, expectedContainer, actualContainer)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestCreateTask(t *testing.T) {
	const (
		testExampleContainerID = "test.container.id"
		testLeaseID            = "test.lease.id"
	)

	testTaskOpt := func(context.Context, *containerd.Client, *containerd.TaskInfo) error { return nil }
	testCreatorFunc := cio.NewCreator()

	testCases := map[string]struct {
		mapExec func(containerWrapper *containerdMocks.MockContainer, task *containerdMocks.MockTask) (containerd.Task, error)
	}{
		"test_no_error": {
			mapExec: func(containerWrapper *containerdMocks.MockContainer, task *containerdMocks.MockTask) (containerd.Task, error) {
				containerWrapper.
					EXPECT().
					NewTask(gomock.Any(), gomock.Any(), matchers.MatchesNewTaskOpts(testTaskOpt)).
					Do(func(ctx context.Context, creatorFunc cio.Creator, opts ...containerd.NewTaskOpts) {
						expected := reflect.ValueOf(testCreatorFunc).Pointer()
						actual := reflect.ValueOf(creatorFunc).Pointer()
						testutil.AssertEqual(t, expected, actual)
					}).
					Times(1).
					Return(task, nil)
				return task, nil
			},
		},
		"test_error": {
			mapExec: func(containerWrapper *containerdMocks.MockContainer, task *containerdMocks.MockTask) (containerd.Task, error) {
				err := log.NewError("example create task error")
				containerWrapper.
					EXPECT().
					NewTask(gomock.Any(), gomock.Any(), matchers.MatchesNewTaskOpts(testTaskOpt)).
					Do(func(ctx context.Context, creatorFunc cio.Creator, opts ...containerd.NewTaskOpts) {
						expected := reflect.ValueOf(testCreatorFunc).Pointer()
						actual := reflect.ValueOf(creatorFunc).Pointer()
						testutil.AssertEqual(t, expected, actual)
					}).
					Times(1).
					Return(nil, err)
				return nil, err
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			taskMock := containerdMocks.NewMockTask(ctrl)
			containerMock := containerdMocks.NewMockContainer(ctrl)
			clientMock := ctrdMocks.NewMockcontainerClientWrapper(ctrl)
			expectedTask, expectedErr := testCase.mapExec(containerMock, taskMock)

			testSpi := &ctrdSpi{
				client: clientMock,
				lease: &leases.Lease{
					ID: testLeaseID,
				},
			}
			actualTask, actualErr := testSpi.CreateTask(context.Background(), containerMock, testCreatorFunc, testTaskOpt)

			testutil.AssertEqual(t, expectedTask, actualTask)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestLoadTask(t *testing.T) {
	const (
		testExampleContainerID = "test.container.id"
		testLeaseID            = "test.lease.id"
	)

	testAttachFunc := cio.NewAttach()

	testCases := map[string]struct {
		mapExec func(containerWrapper *containerdMocks.MockContainer, task *containerdMocks.MockTask) (containerd.Task, error)
	}{
		"test_no_error": {
			mapExec: func(containerWrapper *containerdMocks.MockContainer, task *containerdMocks.MockTask) (containerd.Task, error) {
				containerWrapper.
					EXPECT().
					Task(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, cioReattachFunc cio.Attach) {
						expected := reflect.ValueOf(testAttachFunc).Pointer()
						actual := reflect.ValueOf(cioReattachFunc).Pointer()
						testutil.AssertEqual(t, expected, actual)
					}).
					Times(1).
					Return(task, nil)
				return task, nil
			},
		},
		"test_error": {
			mapExec: func(containerWrapper *containerdMocks.MockContainer, task *containerdMocks.MockTask) (containerd.Task, error) {
				err := log.NewError("example load task error")
				containerWrapper.
					EXPECT().
					Task(gomock.Any(), gomock.Any()).
					Do(func(ctx context.Context, cioReattachFunc cio.Attach) {
						expected := reflect.ValueOf(testAttachFunc).Pointer()
						actual := reflect.ValueOf(cioReattachFunc).Pointer()
						testutil.AssertEqual(t, expected, actual)
					}).
					Times(1).
					Return(nil, err)
				return nil, err
			},
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			taskMock := containerdMocks.NewMockTask(ctrl)
			containerMock := containerdMocks.NewMockContainer(ctrl)
			clientMock := ctrdMocks.NewMockcontainerClientWrapper(ctrl)
			expectedTask, expectedErr := testCase.mapExec(containerMock, taskMock)

			testSpi := &ctrdSpi{
				client: clientMock,
				lease: &leases.Lease{
					ID: testLeaseID,
				},
			}
			actualTask, actualErr := testSpi.LoadTask(context.Background(), containerMock, testAttachFunc)

			testutil.AssertEqual(t, expectedTask, actualTask)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}
