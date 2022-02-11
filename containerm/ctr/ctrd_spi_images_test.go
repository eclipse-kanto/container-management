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
	"github.com/containerd/containerd/remotes"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	containerdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctrd"
	"github.com/golang/mock/gomock"
)

const testImageRef = "testImageRef"

func TestGetImage(t *testing.T) {
	testCases := map[string]struct {
		mapExec func(*ctrd.MockcontainerClientWrapper, *containerdMocks.MockImage) (containerd.Image, error)
	}{
		"test_no_err": {
			mapExec: func(ctrdWrapper *ctrd.MockcontainerClientWrapper, image *containerdMocks.MockImage) (containerd.Image, error) {
				ctrdWrapper.EXPECT().GetImage(gomock.Any(), testImageRef).Times(1).Return(image, nil)
				return image, nil
			},
		},
		"test_err": {
			mapExec: func(ctrdWrapper *ctrd.MockcontainerClientWrapper, _ *containerdMocks.MockImage) (containerd.Image, error) {
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
			mockCtrdWrapper := ctrd.NewMockcontainerClientWrapper(mockCtrl)
			mockImage := containerdMocks.NewMockImage(mockCtrl)
			// mock exec
			expectedImage, expectedErr := testData.mapExec(mockCtrdWrapper, mockImage)
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
		withResolver bool
		mapExec      func(*ctrd.MockcontainerClientWrapper, remotes.Resolver, *containerdMocks.MockImage) (containerd.Image, error)
	}{
		"test_with_resolver": {
			withResolver: true,
			mapExec: func(ctrdWrapper *ctrd.MockcontainerClientWrapper, resolver remotes.Resolver, image *containerdMocks.MockImage) (containerd.Image, error) {
				ctrdWrapper.EXPECT().Pull(gomock.Any(), testImageRef, matchers.MatchesResolverOpts(
					containerd.WithSchema1Conversion,
					containerd.WithPullSnapshotter(testSnapshotterType),
					containerd.WithPullUnpack,
					containerd.WithResolver(resolver))).Times(1).Return(image, nil)
				return image, nil
			},
		},
		"test_without_resolver": {
			mapExec: func(ctrdWrapper *ctrd.MockcontainerClientWrapper, _ remotes.Resolver, image *containerdMocks.MockImage) (containerd.Image, error) {
				ctrdWrapper.EXPECT().Pull(gomock.Any(), testImageRef, matchers.MatchesResolverOpts(
					containerd.WithSchema1Conversion,
					containerd.WithPullSnapshotter(testSnapshotterType),
					containerd.WithPullUnpack)).Times(1).Return(image, nil)
				return image, nil
			},
		},
		"test_err": {
			mapExec: func(ctrdWrapper *ctrd.MockcontainerClientWrapper, _ remotes.Resolver, _ *containerdMocks.MockImage) (containerd.Image, error) {
				err := log.NewError("test pull image error")
				ctrdWrapper.EXPECT().Pull(gomock.Any(), testImageRef, gomock.Any()).Times(1).Return(nil, err)
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
			mockCtrdWrapper := ctrd.NewMockcontainerClientWrapper(mockCtrl)
			mockImage := containerdMocks.NewMockImage(mockCtrl)
			var resolver remotes.Resolver
			if testData.withResolver {
				resolver = containerdMocks.NewMockResolver(mockCtrl)
			}
			// mock exec
			expectedImage, expectedErr := testData.mapExec(mockCtrdWrapper, resolver, mockImage)
			// init spi under test
			testSpi := &ctrdSpi{
				client:          mockCtrdWrapper,
				snapshotterType: testSnapshotterType,
				lease: &leases.Lease{
					ID: containerManagementLeaseID,
				},
			}
			// test
			actualImage, actualErr := testSpi.PullImage(context.Background(), testImageRef, resolver)
			testutil.AssertEqual(t, expectedImage, actualImage)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}
