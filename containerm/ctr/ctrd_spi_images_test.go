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
	"testing"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/leases"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	containerdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	ctrdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctrd"
	"github.com/golang/mock/gomock"
)

const testImageRef = "testImageRef"

func TestGetImage(t *testing.T) {
	testCases := map[string]struct {
		mockExec func(*ctrdMocks.MockcontainerClientWrapper, *containerdMocks.MockImage) (containerd.Image, error)
	}{
		"test_no_err": {
			mockExec: func(ctrdWrapper *ctrdMocks.MockcontainerClientWrapper, image *containerdMocks.MockImage) (containerd.Image, error) {
				ctrdWrapper.EXPECT().GetImage(gomock.Any(), testImageRef).Times(1).Return(image, nil)
				return image, nil
			},
		},
		"test_err": {
			mockExec: func(ctrdWrapper *ctrdMocks.MockcontainerClientWrapper, _ *containerdMocks.MockImage) (containerd.Image, error) {
				err := log.NewError("test get image error")
				ctrdWrapper.EXPECT().GetImage(gomock.Any(), testImageRef).Times(1).Return(nil, err)
				return nil, err
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			// init mock ctrl
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			// init mocks
			mockCtrdWrapper := ctrdMocks.NewMockcontainerClientWrapper(mockCtrl)
			mockImage := containerdMocks.NewMockImage(mockCtrl)
			// mock exec
			expectedImage, expectedErr := testData.mockExec(mockCtrdWrapper, mockImage)
			// init spi under test
			testSpi := &ctrdSpi{
				client: mockCtrdWrapper,
				lease: &leases.Lease{
					ID: containerManagementLeaseID,
				},
			}
			// test
			actualImage, actualErr := testSpi.GetImage(context.Background(), testImageRef)
			testutil.AssertEqual(t, expectedImage, actualImage)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestPullImage(t *testing.T) {
	const testSnapshotterType = "testSnapshotterType"

	testCases := map[string]struct {
		mapExec func(*ctrdMocks.MockcontainerClientWrapper, *containerdMocks.MockImage, *containerdMocks.MockManager) (containerd.Image, error)
	}{
		"test_no_err": {
			mapExec: func(ctrdWrapper *ctrdMocks.MockcontainerClientWrapper, image *containerdMocks.MockImage, leaseManager *containerdMocks.MockManager) (containerd.Image, error) {
				leaseManager.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(leases.WithID(testImageRef))).Return(leases.Lease{ID: testImageRef}, nil)
				ctrdWrapper.EXPECT().Pull(gomock.Any(), testImageRef, matchers.MatchesResolverOpts(
					containerd.WithSchema1Conversion,
					containerd.WithPullSnapshotter(testSnapshotterType),
					containerd.WithPullUnpack)).Times(1).Return(image, nil)
				return image, nil
			},
		},
		"test_pull_err": {
			mapExec: func(ctrdWrapper *ctrdMocks.MockcontainerClientWrapper, _ *containerdMocks.MockImage, leaseManager *containerdMocks.MockManager) (containerd.Image, error) {
				leaseManager.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(leases.WithID(testImageRef))).Return(leases.Lease{ID: testImageRef}, nil)
				err := log.NewError("test pull image error")
				ctrdWrapper.EXPECT().Pull(gomock.Any(), testImageRef, matchers.MatchesResolverOpts(
					containerd.WithSchema1Conversion,
					containerd.WithPullSnapshotter(testSnapshotterType),
					containerd.WithPullUnpack)).Times(1).Return(nil, err)
				return nil, err
			},
		},
		"test_create_lease_err": {
			mapExec: func(_ *ctrdMocks.MockcontainerClientWrapper, _ *containerdMocks.MockImage, leaseManager *containerdMocks.MockManager) (containerd.Image, error) {
				err := log.NewError("test create lease error")
				leaseManager.EXPECT().Create(gomock.Any(), gomock.AssignableToTypeOf(leases.WithID(testImageRef))).Return(leases.Lease{}, err)
				return nil, err
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			// init mock ctrl
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			// init mocks
			mockCtrdWrapper := ctrdMocks.NewMockcontainerClientWrapper(mockCtrl)
			mockImage := containerdMocks.NewMockImage(mockCtrl)
			mockLeaseService := containerdMocks.NewMockManager(mockCtrl)

			// mock exec
			expectedImage, expectedErr := testData.mapExec(mockCtrdWrapper, mockImage, mockLeaseService)
			// init spi under test
			testSpi := &ctrdSpi{
				client:          mockCtrdWrapper,
				snapshotterType: testSnapshotterType,
				lease: &leases.Lease{
					ID: containerManagementLeaseID,
				},
				leaseService: mockLeaseService,
			}
			// test
			actualImage, actualErr := testSpi.PullImage(context.Background(), testImageRef,
				containerd.WithSchema1Conversion,
				containerd.WithPullSnapshotter(testSnapshotterType),
				containerd.WithPullUnpack)
			testutil.AssertEqual(t, expectedImage, actualImage)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}
func TestUnpackImage(t *testing.T) {
	const testSnapshotterType = "testSnapshotterType"

	testCases := map[string]struct {
		mapExec func(*containerdMocks.MockImage) error
	}{
		"test_no_err": {
			mapExec: func(imageMock *containerdMocks.MockImage) error {
				imageMock.EXPECT().Unpack(gomock.Any(), testSnapshotterType, matchers.MatchesUnpackOpts(
					containerd.WithSnapshotterPlatformCheck())).Times(1).Return(nil)
				return nil
			},
		},
		"test_err": {
			mapExec: func(imageMock *containerdMocks.MockImage) error {
				err := log.NewError("test pull image error")
				imageMock.EXPECT().Unpack(gomock.Any(), testSnapshotterType, matchers.MatchesUnpackOpts(
					containerd.WithSnapshotterPlatformCheck())).Times(1).Return(err)
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
			mockImage := containerdMocks.NewMockImage(mockCtrl)

			// mock exec
			expectedErr := testData.mapExec(mockImage)

			// init spi under test
			testSpi := &ctrdSpi{
				snapshotterType: testSnapshotterType,
			}
			// test
			actualErr := testSpi.UnpackImage(context.Background(), mockImage, containerd.WithSnapshotterPlatformCheck())
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}
