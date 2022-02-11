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
	"errors"
	"fmt"
	"testing"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	containerdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	"github.com/golang/mock/gomock"
)

const (
	testType   = "test_type"
	testRootFs = "test_root_fs"
)

var (
	testSnapshotID = fmt.Sprintf(snapshotIDTemplate, testContainerID)
)

func TestGetSnapshot(t *testing.T) {
	testCases := map[string]struct {
		mapExec func(mockSnapshotter *containerdMocks.MockSnapshotter) (snapshots.Info, error)
	}{
		"test_no_err": {
			mapExec: func(mockSnapshotter *containerdMocks.MockSnapshotter) (snapshots.Info, error) {
				info := snapshots.Info{Name: "testSnapshotName"}
				mockSnapshotter.EXPECT().Stat(gomock.Any(), testSnapshotID).Return(info, nil)
				return info, nil
			},
		},
		"test_err": {
			mapExec: func(mockSnapshotter *containerdMocks.MockSnapshotter) (snapshots.Info, error) {
				info := snapshots.Info{}
				err := errors.New("test error")
				mockSnapshotter.EXPECT().Stat(gomock.Any(), testSnapshotID).Return(info, err)
				return info, err
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			// init mock ctrl
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			// init mocks
			mockSnapshotter := containerdMocks.NewMockSnapshotter(mockCtrl)
			// mock exec
			expectedInfo, expectedErr := testData.mapExec(mockSnapshotter)
			// init spi under test
			testSpi := &ctrdSpi{
				snapshotService: mockSnapshotter,
				lease: &leases.Lease{
					ID: containerManagementLeaseID,
				},
			}
			// test
			actualInfo, actualErr := testSpi.GetSnapshot(context.Background(), testContainerID)
			testutil.AssertEqual(t, expectedInfo, actualInfo)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestGetSnapshotID(t *testing.T) {
	testSpi := &ctrdSpi{}
	testutil.AssertEqual(t, "test-container-id-snapshot", testSpi.GetSnapshotID(testContainerID))
}

func TestPrepareSnapshot(t *testing.T) {
	testCases := map[string]struct {
		mapExec func(*containerdMocks.MockImage, *containerdMocks.MockSnapshotter) error
	}{
		"test_image_error_rootfs": {
			mapExec: func(mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test image RootFS error")
				mockImage.EXPECT().RootFS(gomock.Any()).Return(nil, err)
				return err
			},
		},
		"test_error_prepare": {
			mapExec: func(mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test prepare error")
				mockImage.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(gomock.Any(), testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), err)
				return err
			},
		},
		"test_no_error_prepare": {
			mapExec: func(mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := (error)(nil)
				mockImage.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(gomock.Any(), testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), err)
				return err
			},
		},
		"test_error_is_unpacked": {
			mapExec: func(mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test isUnpacked error")

				mockImage.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(gomock.Any(), testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), errdefs.ErrNotFound)
				mockImage.EXPECT().IsUnpacked(gomock.Any(), testType).Return(false, err)
				mockImage.EXPECT().Name().Times(2)
				return err
			},
		},
		"test_error_unpack": {
			mapExec: func(mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test unpack error")

				mockImage.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(gomock.Any(), testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), errdefs.ErrNotFound)
				mockImage.EXPECT().IsUnpacked(gomock.Any(), testType).Return(false, nil)
				mockImage.EXPECT().Name().Times(3)
				mockImage.EXPECT().Unpack(gomock.Any(), testType).Return(err)
				return err
			},
		},
		"test_unpack_success_prepare_fail": {
			mapExec: func(mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test prepare after unpack error")

				mockImage.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(gomock.Any(), testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), errdefs.ErrNotFound)
				mockImage.EXPECT().IsUnpacked(gomock.Any(), testType).Return(false, nil)
				mockImage.EXPECT().Unpack(gomock.Any(), testType).Return(nil)
				mockSnapshotter.EXPECT().Prepare(gomock.Any(), testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), err)
				mockImage.EXPECT().Name().Times(2)
				return err
			},
		},
		"test_unpack_prepare_success": {
			mapExec: func(mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				mockImage.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(gomock.Any(), testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), errdefs.ErrNotFound)
				mockImage.EXPECT().IsUnpacked(gomock.Any(), testType).Return(false, nil)
				mockImage.EXPECT().Unpack(gomock.Any(), testType).Return(nil)
				mockSnapshotter.EXPECT().Prepare(gomock.Any(), testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), nil)
				mockImage.EXPECT().Name().Times(2)
				return nil
			},
		},
		"test_is_packed": {
			mapExec: func(mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				mockImage.EXPECT().RootFS(gomock.Any()).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(gomock.Any(), testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), errdefs.ErrNotFound)
				mockImage.EXPECT().IsUnpacked(gomock.Any(), testType).Return(true, nil)
				mockImage.EXPECT().Name()
				return nil
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockImage := containerdMocks.NewMockImage(mockCtrl)
			mockSnapshotter := containerdMocks.NewMockSnapshotter(mockCtrl)
			expectedErr := testData.mapExec(mockImage, mockSnapshotter)

			testSpi := &ctrdSpi{
				snapshotService: mockSnapshotter,
				snapshotterType: testType,
				lease: &leases.Lease{
					ID: containerManagementLeaseID,
				},
			}

			actualErr := testSpi.PrepareSnapshot(context.Background(), testContainerID, mockImage)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestMountSnapshot(t *testing.T) {
	testCases := map[string]struct {
		mapExec func(*containerdMocks.MockSnapshotter) error
	}{
		"test_error_mounts": {
			mapExec: func(mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test mounts error")
				mockSnapshotter.EXPECT().Mounts(gomock.Any(), testSnapshotID).Return(nil, err)
				return err
			},
		},
		"test_mounts_size_error": {
			mapExec: func(mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("failed to get mounts for snapshot for container test-container-id: not equals 1")
				mockSnapshotter.EXPECT().Mounts(gomock.Any(), testSnapshotID).Return(make([]mount.Mount, 2), nil)
				return err
			},
		},
	}
	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockSnapshotter := containerdMocks.NewMockSnapshotter(mockCtrl)
			expectedErr := testData.mapExec(mockSnapshotter)

			testSpi := &ctrdSpi{
				snapshotService: mockSnapshotter,
				snapshotterType: testType,
				lease: &leases.Lease{
					ID: containerManagementLeaseID,
				},
			}

			actualErr := testSpi.MountSnapshot(context.Background(), testContainerID, testRootFs)
			testutil.AssertError(t, expectedErr, actualErr)

		})
	}
}

func TestRemoveSnapshot(t *testing.T) {
	testCases := map[string]struct {
		mapExec func(*containerdMocks.MockSnapshotter) error
	}{
		"test_error_remove": {
			mapExec: func(mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test error remove")
				mockSnapshotter.EXPECT().Remove(gomock.Any(), testSnapshotID).Return(err)
				return err
			},
		},
		"test_error_not_found_remove": {
			mapExec: func(mockSnapshotter *containerdMocks.MockSnapshotter) error {
				mockSnapshotter.EXPECT().Remove(gomock.Any(), testSnapshotID).Return(errdefs.ErrNotFound)
				return nil
			},
		},
		"test_remove_success": {
			mapExec: func(mockSnapshotter *containerdMocks.MockSnapshotter) error {
				mockSnapshotter.EXPECT().Remove(gomock.Any(), testSnapshotID).Return(nil)
				return nil
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockSnapshotter := containerdMocks.NewMockSnapshotter(mockCtrl)
			expectedErr := testData.mapExec(mockSnapshotter)

			testSpi := &ctrdSpi{
				snapshotService: mockSnapshotter,
				snapshotterType: testType,
				lease: &leases.Lease{
					ID: containerManagementLeaseID,
				},
			}

			actualErr := testSpi.RemoveSnapshot(context.Background(), testContainerID)
			testutil.AssertError(t, expectedErr, actualErr)

		})
	}
}
