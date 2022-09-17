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
	"errors"
	"fmt"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/namespaces"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"testing"

	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	containerdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	"github.com/golang/mock/gomock"
)

func TestGetSnapshot(t *testing.T) {
	const (
		testCtrID     = "test-container-id"
		testLeaseID   = "test.lease"
		testNamespace = "test-ns"
	)
	testSnapshotID := fmt.Sprintf(snapshotIDTemplate, testCtrID)

	testCases := map[string]struct {
		mockExec func(ctx context.Context, mockSnapshotter *containerdMocks.MockSnapshotter) (snapshots.Info, error)
	}{
		"test_no_err": {
			mockExec: func(ctx context.Context, mockSnapshotter *containerdMocks.MockSnapshotter) (snapshots.Info, error) {
				info := snapshots.Info{Name: "testSnapshotName"}
				mockSnapshotter.EXPECT().Stat(ctx, testSnapshotID).Return(info, nil)
				return info, nil
			},
		},
		"test_err": {
			mockExec: func(ctx context.Context, mockSnapshotter *containerdMocks.MockSnapshotter) (snapshots.Info, error) {
				info := snapshots.Info{}
				err := errors.New("test error")
				mockSnapshotter.EXPECT().Stat(ctx, testSnapshotID).Return(info, err)
				return info, err
			},
		},
	}
	prepareContext := func(ctx context.Context) context.Context {
		resCtx := namespaces.WithNamespace(ctx, testNamespace)
		resCtx = leases.WithLease(resCtx, testLeaseID)
		return resCtx
	}
	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			// init mock ctrl
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			// init mocks
			mockSnapshotter := containerdMocks.NewMockSnapshotter(mockCtrl)

			ctx := context.Background()
			// init spi under test
			testSpi := &ctrdSpi{
				snapshotService: mockSnapshotter,
				lease:           &leases.Lease{ID: testLeaseID},
				namespace:       testNamespace,
			}
			// mock exec
			expectedInfo, expectedErr := testData.mockExec(prepareContext(ctx), mockSnapshotter)

			// test
			actualInfo, actualErr := testSpi.GetSnapshot(ctx, testCtrID)
			testutil.AssertEqual(t, expectedInfo, actualInfo)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestGetSnapshotID(t *testing.T) {
	const testCtrID = "test-container-id"
	expectedSnapshotID := fmt.Sprintf(snapshotIDTemplate, testCtrID)

	testSpi := &ctrdSpi{}
	testutil.AssertEqual(t, expectedSnapshotID, testSpi.GetSnapshotID(testCtrID))
}

func TestPrepareSnapshot(t *testing.T) {
	const (
		testType      = "test_type"
		testCtrID     = "test-container-id"
		testLeaseID   = "test.lease"
		testNamespace = "test-ns"
	)
	testSnapshotID := fmt.Sprintf(snapshotIDTemplate, testCtrID)

	prepareContext := func(ctx context.Context, namespace, leaseID string) context.Context {
		resCtx := namespaces.WithNamespace(ctx, namespace)
		if leaseID != "" {
			resCtx = leases.WithLease(resCtx, leaseID)
		}
		return resCtx
	}

	testCases := map[string]struct {
		ctx      context.Context
		mockExec func(context.Context, *containerdMocks.MockImage, *containerdMocks.MockSnapshotter) error
	}{
		"test_image_error_rootfs": {
			mockExec: func(ctx context.Context, mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				testCtx := prepareContext(ctx, testNamespace, testLeaseID)
				err := errors.New("test image RootFS error")
				mockImage.EXPECT().RootFS(testCtx).Return(nil, err)
				return err
			},
		},
		"test_error_prepare": {
			mockExec: func(ctx context.Context, mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test prepare error")
				testCtx := prepareContext(ctx, testNamespace, testLeaseID)
				mockImage.EXPECT().RootFS(testCtx).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(testCtx, testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), err)
				return err
			},
		},
		"test_no_error_prepare": {
			mockExec: func(ctx context.Context, mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				testCtx := prepareContext(ctx, testNamespace, testLeaseID)
				mockImage.EXPECT().RootFS(testCtx).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(testCtx, testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), nil)
				return nil
			},
		},
		"test_error_is_unpacked": {
			mockExec: func(ctx context.Context, mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test isUnpacked error")
				testCtx := prepareContext(ctx, testNamespace, testLeaseID)

				mockImage.EXPECT().RootFS(testCtx).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(testCtx, testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), errdefs.ErrNotFound)
				mockImage.EXPECT().IsUnpacked(testCtx, testType).Return(false, err)
				mockImage.EXPECT().Name().Times(2)
				return err
			},
		},
		"test_error_unpack": {
			mockExec: func(ctx context.Context, mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test unpack error")
				testCtx := prepareContext(ctx, testNamespace, testLeaseID)
				testCtxNoLease := prepareContext(prepareContext(ctx, testNamespace, ""), testNamespace, "")

				mockImage.EXPECT().RootFS(testCtx).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(testCtx, testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), errdefs.ErrNotFound)
				mockImage.EXPECT().IsUnpacked(testCtx, testType).Return(false, nil)
				mockImage.EXPECT().Name().Times(3)
				mockImage.EXPECT().Unpack(testCtxNoLease, testType).Return(err)
				return err
			},
		},
		"test_unpack_success_prepare_fail": {
			mockExec: func(ctx context.Context, mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test prepare after unpack error")
				testCtx := prepareContext(ctx, testNamespace, testLeaseID)
				testCtxNoLease := prepareContext(prepareContext(ctx, testNamespace, ""), testNamespace, "")

				mockImage.EXPECT().RootFS(testCtx).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(testCtx, testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), errdefs.ErrNotFound)
				mockImage.EXPECT().IsUnpacked(testCtx, testType).Return(false, nil)
				mockImage.EXPECT().Unpack(testCtxNoLease, testType).Return(nil)
				mockSnapshotter.EXPECT().Prepare(testCtx, testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), err)
				mockImage.EXPECT().Name().Times(2)
				return err
			},
		},
		"test_unpack_prepare_success": {
			mockExec: func(ctx context.Context, mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				testCtx := prepareContext(ctx, testNamespace, testLeaseID)
				testCtxNoLease := prepareContext(prepareContext(ctx, testNamespace, ""), testNamespace, "")

				mockImage.EXPECT().RootFS(testCtx).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(testCtx, testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), errdefs.ErrNotFound)
				mockImage.EXPECT().IsUnpacked(testCtx, testType).Return(false, nil)
				mockImage.EXPECT().Unpack(testCtxNoLease, testType).Return(nil)
				mockSnapshotter.EXPECT().Prepare(testCtx, testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), nil)
				mockImage.EXPECT().Name().Times(2)
				return nil
			},
		},
		"test_is_packed": {
			mockExec: func(ctx context.Context, mockImage *containerdMocks.MockImage, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				testCtx := prepareContext(ctx, testNamespace, testLeaseID)

				mockImage.EXPECT().RootFS(testCtx).Return(nil, nil)
				mockSnapshotter.EXPECT().Prepare(testCtx, testSnapshotID, gomock.Any()).Return(make([]mount.Mount, 0), errdefs.ErrNotFound)
				mockImage.EXPECT().IsUnpacked(testCtx, testType).Return(true, nil)
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

			testSpi := &ctrdSpi{
				snapshotService: mockSnapshotter,
				snapshotterType: testType,
				lease:           &leases.Lease{ID: testLeaseID},
				namespace:       testNamespace,
			}
			ctx := context.Background()
			expectedErr := testData.mockExec(ctx, mockImage, mockSnapshotter)

			actualErr := testSpi.PrepareSnapshot(ctx, testCtrID, mockImage)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestMountSnapshot(t *testing.T) {
	const (
		testType      = "test_type"
		testRootFs    = "test_root_fs"
		testCtrID     = "test-container-id"
		testLeaseID   = "test.lease"
		testNamespace = "test-ns"
	)
	testSnapshotID := fmt.Sprintf(snapshotIDTemplate, testCtrID)

	testCases := map[string]struct {
		mockExec func(context.Context, *containerdMocks.MockSnapshotter) error
	}{
		"test_error_mounts": {
			mockExec: func(ctx context.Context, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test mounts error")
				mockSnapshotter.EXPECT().Mounts(ctx, testSnapshotID).Return(nil, err)
				return err
			},
		},
		"test_mounts_size_error": {
			mockExec: func(ctx context.Context, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("failed to get mounts for snapshot for container test-container-id: not equals 1")
				mockSnapshotter.EXPECT().Mounts(ctx, testSnapshotID).Return(make([]mount.Mount, 2), nil)
				return err
			},
		},
	}
	prepareContext := func(ctx context.Context) context.Context {
		resCtx := namespaces.WithNamespace(ctx, testNamespace)
		resCtx = leases.WithLease(resCtx, testLeaseID)
		return resCtx
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockSnapshotter := containerdMocks.NewMockSnapshotter(mockCtrl)
			testSpi := &ctrdSpi{
				snapshotService: mockSnapshotter,
				snapshotterType: testType,
				lease:           &leases.Lease{ID: testLeaseID},
				namespace:       testNamespace,
			}
			ctx := context.Background()

			expectedErr := testData.mockExec(prepareContext(ctx), mockSnapshotter)

			actualErr := testSpi.MountSnapshot(ctx, testCtrID, testRootFs)
			testutil.AssertError(t, expectedErr, actualErr)

		})
	}
}

func TestRemoveSnapshot(t *testing.T) {
	const (
		testType      = "test_type"
		testCtrID     = "test-container-id"
		testLeaseID   = "test.lease"
		testNamespace = "test-ns"
	)
	testSnapshotID := fmt.Sprintf(snapshotIDTemplate, testCtrID)

	testCases := map[string]struct {
		mockExec func(context.Context, *containerdMocks.MockSnapshotter) error
	}{
		"test_error_remove": {
			mockExec: func(ctx context.Context, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				err := errors.New("test error remove")
				mockSnapshotter.EXPECT().Remove(ctx, testSnapshotID).Return(err)
				return err
			},
		},
		"test_error_not_found_remove": {
			mockExec: func(ctx context.Context, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				mockSnapshotter.EXPECT().Remove(ctx, testSnapshotID).Return(errdefs.ErrNotFound)
				return nil
			},
		},
		"test_remove_success": {
			mockExec: func(ctx context.Context, mockSnapshotter *containerdMocks.MockSnapshotter) error {
				mockSnapshotter.EXPECT().Remove(ctx, testSnapshotID).Return(nil)
				return nil
			},
		},
	}
	prepareContext := func(ctx context.Context) context.Context {
		resCtx := namespaces.WithNamespace(ctx, testNamespace)
		resCtx = leases.WithLease(resCtx, testLeaseID)
		return resCtx
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockSnapshotter := containerdMocks.NewMockSnapshotter(mockCtrl)

			testSpi := &ctrdSpi{
				snapshotService: mockSnapshotter,
				snapshotterType: testType,
				lease:           &leases.Lease{ID: testLeaseID},
				namespace:       testNamespace,
			}
			ctx := context.Background()

			expectedErr := testData.mockExec(prepareContext(ctx), mockSnapshotter)

			actualErr := testSpi.RemoveSnapshot(ctx, testCtrID)
			testutil.AssertError(t, expectedErr, actualErr)

		})
	}
}

func TestListSnapshots(t *testing.T) {
	const (
		testType      = "test_type"
		testFilter    = "name=test-snapshot"
		testLeaseID   = "test.lease"
		testNamespace = "test-ns"
	)

	testCases := map[string]struct {
		mockExec func(context.Context, *containerdMocks.MockSnapshotter) ([]snapshots.Info, error)
	}{
		"test_walk_error": {
			mockExec: func(ctx context.Context, snapshotter *containerdMocks.MockSnapshotter) ([]snapshots.Info, error) {
				err := log.NewError("test error")
				snapshotter.EXPECT().Walk(ctx, gomock.Any(), testFilter).Return(err)
				return nil, err
			},
		},
		"test_no_error": {
			mockExec: func(ctx context.Context, snapshotter *containerdMocks.MockSnapshotter) ([]snapshots.Info, error) {
				testSnapshotInfo := snapshots.Info{
					Name: "test-snapshot",
				}
				snapshotter.EXPECT().Walk(ctx, gomock.Any(), testFilter).Do(
					func(ctx context.Context, fn snapshots.WalkFunc, filters ...string) error {
						_ = fn(ctx, testSnapshotInfo)
						return nil
					},
				)
				return []snapshots.Info{testSnapshotInfo}, nil
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()

			mockSnapshotter := containerdMocks.NewMockSnapshotter(mockCtrl)

			testSpi := &ctrdSpi{
				snapshotService: mockSnapshotter,
				snapshotterType: testType,
				lease:           &leases.Lease{ID: testLeaseID},
				namespace:       testNamespace,
			}
			ctx := context.Background()

			expectedSnapshots, expectedErr := testData.mockExec(namespaces.WithNamespace(ctx, testNamespace), mockSnapshotter)

			actualSnapshots, actualErr := testSpi.ListSnapshots(ctx, testFilter)
			testutil.AssertError(t, expectedErr, actualErr)
			testutil.AssertEqual(t, expectedSnapshots, actualSnapshots)

		})
	}
}
