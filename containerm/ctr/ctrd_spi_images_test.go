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
	"github.com/containerd/containerd/namespaces"
	"testing"

	"github.com/containerd/containerd"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	containerdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	ctrdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctrd"
	"github.com/golang/mock/gomock"
)

func TestGetImage(t *testing.T) {
	const (
		testImageRef  = "test.img/ref:latest"
		testNamespace = "test-ns"
	)
	testCases := map[string]struct {
		mockExec func(context.Context, *ctrdMocks.MockcontainerClientWrapper, *containerdMocks.MockImage) (containerd.Image, error)
	}{
		"test_no_err": {
			mockExec: func(ctx context.Context, ctrdWrapper *ctrdMocks.MockcontainerClientWrapper, image *containerdMocks.MockImage) (containerd.Image, error) {
				ctrdWrapper.EXPECT().GetImage(ctx, testImageRef).Times(1).Return(image, nil)
				return image, nil
			},
		},
		"test_err": {
			mockExec: func(ctx context.Context, ctrdWrapper *ctrdMocks.MockcontainerClientWrapper, _ *containerdMocks.MockImage) (containerd.Image, error) {
				err := log.NewError("test get image error")
				ctrdWrapper.EXPECT().GetImage(ctx, testImageRef).Times(1).Return(nil, err)
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
			// init spi under test
			testSpi := &ctrdSpi{
				client:    mockCtrdWrapper,
				namespace: testNamespace,
			}
			ctx := context.Background()
			// mock exec
			expectedImage, expectedErr := testData.mockExec(namespaces.WithNamespace(ctx, testNamespace), mockCtrdWrapper, mockImage)
			// test
			actualImage, actualErr := testSpi.GetImage(ctx, testImageRef)
			testutil.AssertEqual(t, expectedImage, actualImage)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestPullImage(t *testing.T) {
	const (
		testSnapshotterType = "testSnapshotterType"
		testImageRef        = "test.img/ref:latest"
		testNamespace       = "test-ns"
	)

	testCases := map[string]struct {
		mapExec func(context.Context, *ctrdMocks.MockcontainerClientWrapper, *containerdMocks.MockImage) (containerd.Image, error)
	}{
		"test_no_err": {
			mapExec: func(ctx context.Context, ctrdWrapper *ctrdMocks.MockcontainerClientWrapper, image *containerdMocks.MockImage) (containerd.Image, error) {
				ctrdWrapper.EXPECT().Pull(ctx, testImageRef, matchers.MatchesResolverOpts(
					containerd.WithSchema1Conversion,
					containerd.WithPullSnapshotter(testSnapshotterType),
					containerd.WithPullUnpack)).Times(1).Return(image, nil)
				return image, nil
			},
		},
		"test_pull_err": {
			mapExec: func(ctx context.Context, ctrdWrapper *ctrdMocks.MockcontainerClientWrapper, _ *containerdMocks.MockImage) (containerd.Image, error) {
				err := log.NewError("test pull image error")
				ctrdWrapper.EXPECT().Pull(ctx, testImageRef, matchers.MatchesResolverOpts(
					containerd.WithSchema1Conversion,
					containerd.WithPullSnapshotter(testSnapshotterType),
					containerd.WithPullUnpack)).Times(1).Return(nil, err)
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

			// init spi under test
			testSpi := &ctrdSpi{
				client:          mockCtrdWrapper,
				snapshotterType: testSnapshotterType,
				namespace:       testNamespace,
			}
			ctx := context.Background()

			// mock exec
			expectedImage, expectedErr := testData.mapExec(namespaces.WithNamespace(ctx, testNamespace), mockCtrdWrapper, mockImage)

			// test
			actualImage, actualErr := testSpi.PullImage(ctx, testImageRef,
				containerd.WithSchema1Conversion,
				containerd.WithPullSnapshotter(testSnapshotterType),
				containerd.WithPullUnpack)
			testutil.AssertEqual(t, expectedImage, actualImage)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}
func TestUnpackImage(t *testing.T) {
	const (
		testSnapshotterType = "testSnapshotterType"
		testNamespace       = "test-ns"
	)

	testCases := map[string]struct {
		mapExec func(context.Context, *containerdMocks.MockImage) error
	}{
		"test_no_err": {
			mapExec: func(ctx context.Context, imageMock *containerdMocks.MockImage) error {
				imageMock.EXPECT().Unpack(ctx, testSnapshotterType, matchers.MatchesUnpackOpts(
					containerd.WithSnapshotterPlatformCheck())).Times(1).Return(nil)
				return nil
			},
		},
		"test_err": {
			mapExec: func(ctx context.Context, imageMock *containerdMocks.MockImage) error {
				err := log.NewError("test pull image error")
				imageMock.EXPECT().Unpack(ctx, testSnapshotterType, matchers.MatchesUnpackOpts(
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

			// init spi under test
			testSpi := &ctrdSpi{
				snapshotterType: testSnapshotterType,
				namespace:       testNamespace,
			}
			ctx := context.Background()

			// mock exec
			expectedErr := testData.mapExec(namespaces.WithNamespace(ctx, testNamespace), mockImage)

			// test
			actualErr := testSpi.UnpackImage(ctx, mockImage, containerd.WithSnapshotterPlatformCheck())
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}
