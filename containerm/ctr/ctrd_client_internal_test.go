// Copyright (c) 2022 Contributors to the Eclipse Foundation
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
	"crypto/sha256"
	"fmt"
	"github.com/containerd/containerd"
	eventstypes "github.com/containerd/containerd/api/events"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/runtime"
	"github.com/containerd/containerd/snapshots"
	"github.com/containerd/imgcrypt"
	"github.com/containerd/imgcrypt/images/encryption"
	"github.com/containerd/typeurl"
	"github.com/containers/ocicrypt/config"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	mocksContainerd "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	mocksCtrd "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctrd"
	mocksIo "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/io"
	mocksLogger "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/logger"
	"github.com/eclipse-kanto/container-management/containerm/util"
	protoTypes "github.com/gogo/protobuf/types"
	"github.com/golang/mock/gomock"
	"github.com/opencontainers/go-digest"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestClientInternalGenerateNewContainerOpts(t *testing.T) {
	const (
		containerID       = "test_container_id"
		snapshotID        = containerID + "snapshot"
		containerImageRef = "some.repo/image:tag"
		rootExec          = "some/exec/path"
	)
	container := &types.Container{
		ID: containerID,
		Image: types.Image{
			Name:          containerImageRef,
			DecryptConfig: &types.DecryptConfig{},
		},
		HostConfig: &types.HostConfig{},
	}
	testCases := map[string]struct {
		mockExec func(imageMock *mocksContainerd.MockImage, spiMock *mocksCtrd.MockcontainerdSpi, decrytpMgrMock *mocksCtrd.MockcontainerDecryptMgr) ([]containerd.NewContainerOpts, error)
	}{
		"test_no_error": {
			mockExec: func(imageMock *mocksContainerd.MockImage, spiMock *mocksCtrd.MockcontainerdSpi, decrytpMgrMock *mocksCtrd.MockcontainerDecryptMgr) ([]containerd.NewContainerOpts, error) {
				spiMock.EXPECT().GetSnapshotID(container.ID)
				dc := &config.DecryptConfig{}
				decrytpMgrMock.EXPECT().GetDecryptConfig(container.Image.DecryptConfig).Return(dc, nil)
				res := WithSnapshotOpts(snapshotID, containerd.DefaultSnapshotter) // what these With* return must be tested for each dedicated static func
				res = append(res,
					WithRuntimeOpts(container, rootExec),
					WithSpecOpts(container, imageMock, rootExec),
					encryption.WithAuthorizationCheck(dc),
				)
				return res, nil
			},
		},
		"test_error": {
			mockExec: func(imageMock *mocksContainerd.MockImage, spiMock *mocksCtrd.MockcontainerdSpi, decrytpMgrMock *mocksCtrd.MockcontainerDecryptMgr) ([]containerd.NewContainerOpts, error) {
				spiMock.EXPECT().GetSnapshotID(container.ID)
				err := log.NewError("test error")
				decrytpMgrMock.EXPECT().GetDecryptConfig(container.Image.DecryptConfig).Return(nil, err)
				return nil, err
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			decryptMgrMock := mocksCtrd.NewMockcontainerDecryptMgr(ctrl)
			ctrdClient := &containerdClient{
				spi:      spiMock,
				decMgr:   decryptMgrMock,
				rootExec: rootExec,
			}
			imageMock := mocksContainerd.NewMockImage(ctrl)
			expectedOpts, expectedErr := testCaseData.mockExec(imageMock, spiMock, decryptMgrMock)
			actualOpts, actualErr := ctrdClient.generateNewContainerOpts(container, imageMock)
			testutil.AssertError(t, expectedErr, actualErr)
			testutil.AssertTrue(t, matchers.MatchesNewContainerOpts(expectedOpts...).Matches(actualOpts))
		})
	}
}

func TestClientInternalGenerateRemoteOpts(t *testing.T) {
	const containerImageRef = "some.repo/image:tag"
	testImageInfo := types.Image{
		Name: containerImageRef,
	}
	testCases := map[string]struct {
		mockExec func(resolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) []containerd.RemoteOpt
	}{
		"test_with_resolver": {
			mockExec: func(resolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) []containerd.RemoteOpt {
				ctrdResolverMock := mocksContainerd.NewMockResolver(ctrl)
				resolverMock.EXPECT().ResolveImageRegistry(util.GetImageHost(testImageInfo.Name)).Return(ctrdResolverMock)
				return []containerd.RemoteOpt{
					containerd.WithSchema1Conversion,
					containerd.WithResolver(ctrdResolverMock),
				}
			},
		},
		"test_without_resolver": {
			mockExec: func(resolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) []containerd.RemoteOpt {
				resolverMock.EXPECT().ResolveImageRegistry(util.GetImageHost(testImageInfo.Name)).Return(nil)
				return []containerd.RemoteOpt{
					containerd.WithSchema1Conversion,
				}
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			regResolverMock := mocksCtrd.NewMockcontainerImageRegistriesResolver(ctrl)
			ctrdClient := &containerdClient{
				registriesResolver: regResolverMock,
			}
			expectedOpts := testCaseData.mockExec(regResolverMock, ctrl)
			actualOpts := ctrdClient.generateRemoteOpts(testImageInfo)
			testutil.AssertTrue(t, matchers.MatchesResolverOpts(expectedOpts...).Matches(actualOpts))
		})
	}
}

func TestClientInternalGenerateUnpackOpts(t *testing.T) {
	const containerImageRef = "some.repo/image:tag"
	testImageInfo := types.Image{
		Name:          containerImageRef,
		DecryptConfig: &types.DecryptConfig{},
	}
	testCases := map[string]struct {
		mockExec func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr) ([]containerd.UnpackOpt, error)
	}{
		"test_no_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr) ([]containerd.UnpackOpt, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				return []containerd.UnpackOpt{
					encryption.WithUnpackConfigApplyOpts(encryption.WithDecryptedUnpack(&imgcrypt.Payload{DecryptConfig: *dc})),
				}, nil
			},
		},
		"test_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr) ([]containerd.UnpackOpt, error) {
				err := log.NewError("test error")
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(nil, err)
				return nil, err
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			decryptMgrMock := mocksCtrd.NewMockcontainerDecryptMgr(ctrl)
			ctrdClient := &containerdClient{
				decMgr: decryptMgrMock,
			}
			expectedOpts, expectedErr := testCaseData.mockExec(decryptMgrMock)
			actualOpts, actualErr := ctrdClient.generateUnpackOpts(testImageInfo)
			testutil.AssertError(t, expectedErr, actualErr)
			testutil.AssertTrue(t, matchers.MatchesUnpackOpts(expectedOpts...).Matches(actualOpts))
		})
	}
}

func TestClientInternalGetImage(t *testing.T) {
	const containerImageRef = "some.repo/image:tag"
	testImageInfo := types.Image{
		Name:          containerImageRef,
		DecryptConfig: &types.DecryptConfig{},
	}
	testCases := map[string]struct {
		mockExec func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, ctrl *gomock.Controller) (containerd.Image, error)
	}{
		"test_get_decrypt_cfg_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, ctrl *gomock.Controller) (containerd.Image, error) {
				err := log.NewError("test error")
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(nil, err)
				return nil, err
			},
		},
		"test_spi_get_image_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				err := log.NewError("test error")
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(nil, err)
				return nil, err
			},
		},
		"test_check_auth_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				imageMock := mocksContainerd.NewMockImage(ctrl)
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(imageMock, nil)
				err := log.NewError("test error")
				decryptMgrMock.EXPECT().CheckAuthorization(gomock.Any(), imageMock, dc).Return(err)
				return nil, err
			},
		},
		"test_no_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				imageMock := mocksContainerd.NewMockImage(ctrl)
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(imageMock, nil)
				decryptMgrMock.EXPECT().CheckAuthorization(gomock.Any(), imageMock, dc).Return(nil)
				return imageMock, nil
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			decryptMgrMock := mocksCtrd.NewMockcontainerDecryptMgr(ctrl)
			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			ctrdClient := &containerdClient{
				decMgr: decryptMgrMock,
				spi:    spiMock,
			}
			expectedImage, expectedErr := testCaseData.mockExec(decryptMgrMock, spiMock, ctrl)
			actualImage, actualErr := ctrdClient.getImage(context.TODO(), testImageInfo)
			testutil.AssertError(t, expectedErr, actualErr)
			testutil.AssertEqual(t, expectedImage, actualImage)
		})
	}
}

func TestClientInternalPullImage(t *testing.T) {
	const containerImageRef = "some.repo/image:tag"
	testImageInfo := types.Image{
		Name:          containerImageRef,
		DecryptConfig: &types.DecryptConfig{},
	}
	testCases := map[string]struct {
		mockExec func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, regsResolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) (containerd.Image, error)
	}{
		"test_get_decrypt_cfg_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, regsResolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) (containerd.Image, error) {
				err := log.NewError("test error")
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(nil, err)
				return nil, err
			},
		},
		"test_spi_get_image_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, regsResolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				err := log.NewError("test error")
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(nil, err)
				return nil, err
			},
		},
		"test_spi_get_image_available_check_auth_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, regsResolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				imageMock := mocksContainerd.NewMockImage(ctrl)
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(imageMock, nil)
				err := log.NewError("test error")
				decryptMgrMock.EXPECT().CheckAuthorization(gomock.Any(), imageMock, dc).Return(err)
				return nil, err
			},
		},
		"test_spi_get_image_available_no_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, regsResolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				imageMock := mocksContainerd.NewMockImage(ctrl)
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(imageMock, nil)
				decryptMgrMock.EXPECT().CheckAuthorization(gomock.Any(), imageMock, dc).Return(nil)
				return imageMock, nil
			},
		},
		"test_spi_get_image_not_available_pull_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, regsResolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(nil, errdefs.ErrNotFound)
				regsResolverMock.EXPECT().ResolveImageRegistry(util.GetImageHost(testImageInfo.Name)).Return(nil)
				err := log.NewError("test error")
				spiMock.EXPECT().PullImage(gomock.Any(), testImageInfo.Name, matchers.MatchesResolverOpts(containerd.WithSchema1Conversion)).Return(nil, err)
				return nil, err
			},
		},
		"test_spi_get_image_not_available_check_auth_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, regsResolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(nil, errdefs.ErrNotFound)
				regsResolverMock.EXPECT().ResolveImageRegistry(util.GetImageHost(testImageInfo.Name)).Return(nil)
				imageMock := mocksContainerd.NewMockImage(ctrl)
				spiMock.EXPECT().PullImage(gomock.Any(), testImageInfo.Name, matchers.MatchesResolverOpts(containerd.WithSchema1Conversion)).Return(imageMock, nil)
				err := log.NewError("test error")
				decryptMgrMock.EXPECT().CheckAuthorization(gomock.Any(), imageMock, dc).Return(err)
				return nil, err
			},
		},
		"test_spi_get_image_not_available_gen_unpack_opts_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, regsResolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(nil, errdefs.ErrNotFound)
				regsResolverMock.EXPECT().ResolveImageRegistry(util.GetImageHost(testImageInfo.Name)).Return(nil)
				imageMock := mocksContainerd.NewMockImage(ctrl)
				spiMock.EXPECT().PullImage(gomock.Any(), testImageInfo.Name, matchers.MatchesResolverOpts(containerd.WithSchema1Conversion)).Return(imageMock, nil)
				decryptMgrMock.EXPECT().CheckAuthorization(gomock.Any(), imageMock, dc).Return(nil)
				err := log.NewError("test error")
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(nil, err)
				return nil, err
			},
		},
		"test_spi_get_image_not_available_unpack_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, regsResolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil).Times(2)
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(nil, errdefs.ErrNotFound)
				regsResolverMock.EXPECT().ResolveImageRegistry(util.GetImageHost(testImageInfo.Name)).Return(nil)
				imageMock := mocksContainerd.NewMockImage(ctrl)
				spiMock.EXPECT().PullImage(gomock.Any(), testImageInfo.Name, matchers.MatchesResolverOpts(containerd.WithSchema1Conversion)).Return(imageMock, nil)
				decryptMgrMock.EXPECT().CheckAuthorization(gomock.Any(), imageMock, dc).Return(nil)
				err := log.NewError("test error")
				spiMock.EXPECT().UnpackImage(gomock.Any(), imageMock, matchers.MatchesUnpackOpts(encryption.WithUnpackConfigApplyOpts(encryption.WithDecryptedUnpack(&imgcrypt.Payload{DecryptConfig: *dc})))).Return(err)
				return nil, err
			},
		},
		"test_spi_get_image_not_available_no_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, regsResolverMock *mocksCtrd.MockcontainerImageRegistriesResolver, ctrl *gomock.Controller) (containerd.Image, error) {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil).Times(2)
				spiMock.EXPECT().GetImage(gomock.Any(), testImageInfo.Name).Return(nil, errdefs.ErrNotFound)
				regsResolverMock.EXPECT().ResolveImageRegistry(util.GetImageHost(testImageInfo.Name)).Return(nil)
				imageMock := mocksContainerd.NewMockImage(ctrl)
				spiMock.EXPECT().PullImage(gomock.Any(), testImageInfo.Name, matchers.MatchesResolverOpts(containerd.WithSchema1Conversion)).Return(imageMock, nil)
				decryptMgrMock.EXPECT().CheckAuthorization(gomock.Any(), imageMock, dc).Return(nil)
				spiMock.EXPECT().UnpackImage(gomock.Any(), imageMock, matchers.MatchesUnpackOpts(encryption.WithUnpackConfigApplyOpts(encryption.WithDecryptedUnpack(&imgcrypt.Payload{DecryptConfig: *dc})))).Return(nil)
				return imageMock, nil
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			decryptMgrMock := mocksCtrd.NewMockcontainerDecryptMgr(ctrl)
			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			registriesResolverMock := mocksCtrd.NewMockcontainerImageRegistriesResolver(ctrl)
			ctrdClient := &containerdClient{
				decMgr:             decryptMgrMock,
				spi:                spiMock,
				registriesResolver: registriesResolverMock,
			}
			expectedImage, expectedErr := testCaseData.mockExec(decryptMgrMock, spiMock, registriesResolverMock, ctrl)
			actualImage, actualErr := ctrdClient.pullImage(context.TODO(), testImageInfo)
			testutil.AssertError(t, expectedErr, actualErr)
			testutil.AssertEqual(t, expectedImage, actualImage)
		})
	}
}

func TestClientInternalCreateSnapshot(t *testing.T) {
	const (
		containerID       = "test-container-id"
		containerName     = "test-container-name"
		containerImageRef = "some.repo/image:tag"
	)
	testImageInfo := types.Image{
		Name:          containerImageRef,
		DecryptConfig: &types.DecryptConfig{},
	}
	testCases := map[string]struct {
		mockExec func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error
	}{
		"test_gen_unpack_opts_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				err := log.NewError("test error")
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(nil, err)
				imageMock.EXPECT().Name().Return(containerName)
				return err
			},
		},
		"test_prepare_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				err := log.NewError("test error")
				spiMock.EXPECT().PrepareSnapshot(gomock.Any(), containerID, imageMock, matchers.MatchesUnpackOpts(encryption.WithUnpackConfigApplyOpts(encryption.WithDecryptedUnpack(&imgcrypt.Payload{DecryptConfig: *dc})))).Return(err)
				imageMock.EXPECT().Name().Return(containerName)
				return err
			},
		},
		"test_mount_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				spiMock.EXPECT().PrepareSnapshot(gomock.Any(), containerID, imageMock, matchers.MatchesUnpackOpts(encryption.WithUnpackConfigApplyOpts(encryption.WithDecryptedUnpack(&imgcrypt.Payload{DecryptConfig: *dc})))).Return(nil)
				err := log.NewError("test error")
				spiMock.EXPECT().MountSnapshot(gomock.Any(), containerID, rootFSPathDefault).Return(err)
				imageMock.EXPECT().Name().Return(containerName)
				return err
			},
		},
		"test_no_error": {
			mockExec: func(decryptMgrMock *mocksCtrd.MockcontainerDecryptMgr, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				dc := &config.DecryptConfig{}
				decryptMgrMock.EXPECT().GetDecryptConfig(testImageInfo.DecryptConfig).Return(dc, nil)
				spiMock.EXPECT().PrepareSnapshot(gomock.Any(), containerID, imageMock, matchers.MatchesUnpackOpts(encryption.WithUnpackConfigApplyOpts(encryption.WithDecryptedUnpack(&imgcrypt.Payload{DecryptConfig: *dc})))).Return(nil)
				spiMock.EXPECT().MountSnapshot(gomock.Any(), containerID, rootFSPathDefault).Return(nil)
				return nil
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			decryptMgrMock := mocksCtrd.NewMockcontainerDecryptMgr(ctrl)
			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			imageMock := mocksContainerd.NewMockImage(ctrl)
			ctrdClient := &containerdClient{
				decMgr: decryptMgrMock,
				spi:    spiMock,
			}
			expectedErr := testCaseData.mockExec(decryptMgrMock, spiMock, imageMock)
			actualErr := ctrdClient.createSnapshot(context.TODO(), containerID, imageMock, testImageInfo)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestClientInternalClearSnapshot(t *testing.T) {
	const (
		containerID = "test-container-id"
	)
	testCases := map[string]struct {
		mockExec func(spiMock *mocksCtrd.MockcontainerdSpi)
	}{
		"test_remove_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi) {
				err := log.NewError("test error")
				spiMock.EXPECT().RemoveSnapshot(gomock.Any(), containerID).Return(err)
				spiMock.EXPECT().UnmountSnapshot(gomock.Any(), containerID, rootFSPathDefault).Return(nil)
			},
		},
		"test_unmount_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi) {
				spiMock.EXPECT().RemoveSnapshot(gomock.Any(), containerID).Return(nil)
				err := log.NewError("test error")
				spiMock.EXPECT().UnmountSnapshot(gomock.Any(), containerID, rootFSPathDefault).Return(err)
			},
		},
		"test_all_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi) {
				err := log.NewError("test error")
				spiMock.EXPECT().RemoveSnapshot(gomock.Any(), containerID).Return(err)
				spiMock.EXPECT().UnmountSnapshot(gomock.Any(), containerID, rootFSPathDefault).Return(err)
			},
		},
		"test_no_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi) {
				spiMock.EXPECT().RemoveSnapshot(gomock.Any(), containerID).Return(nil)
				spiMock.EXPECT().UnmountSnapshot(gomock.Any(), containerID, rootFSPathDefault).Return(nil)
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			ctrdClient := &containerdClient{
				spi: spiMock,
			}
			testCaseData.mockExec(spiMock)
			ctrdClient.clearSnapshot(context.TODO(), containerID)
		})
	}
}

func TestClientInternalCreateTask(t *testing.T) {
	const (
		containerID          = "test-container-id"
		checkpointDir        = "some/dir"
		taskPid       uint32 = 123
	)
	testCioCreator := func(id string) (cio.IO, error) {
		return nil, nil
	}
	testCtrIOCfg := &types.IOConfig{Tty: true}
	testStatusChan := make(chan containerd.ExitStatus)

	testCases := map[string]struct {
		mockExec func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error)
	}{
		"test_create_task_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error) {
				ioMgrMock.EXPECT().NewCioCreator(testCtrIOCfg.Tty).Return(testCioCreator)
				err := log.NewError("test error")
				spiMock.EXPECT().CreateTask(gomock.Any(), mockContainer, matchers.MatchesCioCreator(testCioCreator)).Return(nil, err)
				return nil, err
			},
		},
		"test_wait_task_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error) {
				ioMgrMock.EXPECT().NewCioCreator(testCtrIOCfg.Tty).Return(testCioCreator)
				taskMock := mocksContainerd.NewMockTask(ctrl)
				spiMock.EXPECT().CreateTask(gomock.Any(), mockContainer, matchers.MatchesCioCreator(testCioCreator)).Return(taskMock, nil)
				err := log.NewError("test error")
				taskMock.EXPECT().Wait(gomock.Any()).Return(nil, err)
				taskMock.EXPECT().Delete(gomock.Any()).Return(nil, nil)
				return nil, err
			},
		},
		"test_wait_task_delete_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error) {
				ioMgrMock.EXPECT().NewCioCreator(testCtrIOCfg.Tty).Return(testCioCreator)
				taskMock := mocksContainerd.NewMockTask(ctrl)
				spiMock.EXPECT().CreateTask(gomock.Any(), mockContainer, matchers.MatchesCioCreator(testCioCreator)).Return(taskMock, nil)
				err := log.NewError("test error")
				taskMock.EXPECT().Wait(gomock.Any()).Return(nil, err)
				taskMock.EXPECT().Delete(gomock.Any()).Return(nil, err)
				return nil, err
			},
		},
		"test_no_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error) {
				ioMgrMock.EXPECT().NewCioCreator(testCtrIOCfg.Tty).Return(testCioCreator)
				taskMock := mocksContainerd.NewMockTask(ctrl)
				spiMock.EXPECT().CreateTask(gomock.Any(), mockContainer, matchers.MatchesCioCreator(testCioCreator)).Return(taskMock, nil)
				taskMock.EXPECT().Wait(gomock.Any()).Return(testStatusChan, nil)
				taskMock.EXPECT().Pid().Return(taskPid)
				return &containerInfo{
					container:     mockContainer,
					task:          taskMock,
					statusChannel: testStatusChan,
					resultChannel: make(chan exitInfo),
				}, nil
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			ioMgrMock := NewMockcontainerIOManager(ctrl)
			ctrdClient := &containerdClient{
				spi:   spiMock,
				ioMgr: ioMgrMock,
			}
			containerMock := mocksContainerd.NewMockContainer(ctrl)
			expectedCtrInfo, expectedErr := testCaseData.mockExec(spiMock, ioMgrMock, containerMock, ctrl)
			actualCtrInfo, actualErr := ctrdClient.createTask(context.TODO(), testCtrIOCfg, containerID, checkpointDir, containerMock)

			testutil.AssertError(t, expectedErr, actualErr)
			if expectedCtrInfo != nil {
				testutil.AssertNotNil(t, expectedCtrInfo)
				testutil.AssertEqual(t, expectedCtrInfo.task, actualCtrInfo.task)
				testutil.AssertEqual(t, expectedCtrInfo.container, actualCtrInfo.container)
				testutil.AssertEqual(t, expectedCtrInfo.statusChannel, actualCtrInfo.statusChannel)
				testutil.AssertNotNil(t, actualCtrInfo.resultChannel)
			} else {
				testutil.AssertNil(t, actualCtrInfo)
			}
		})
	}
}

func TestClientInternalLoadTask(t *testing.T) {
	const (
		containerID   = "test-container-id"
		checkpointDir = "some/dir"
	)
	testCioAttach := func(*cio.FIFOSet) (cio.IO, error) {
		return nil, nil
	}
	testStatusChan := make(chan containerd.ExitStatus)

	testCases := map[string]struct {
		mockExec func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error)
	}{
		"test_timeout_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error) {
				ioMgrMock.EXPECT().NewCioAttach(containerID).Return(testCioAttach).Times(3)
				spiMock.EXPECT().LoadTask(gomock.Any(), mockContainer, matchers.MatchesCioAttach(testCioAttach)).
					Do(func(ctx context.Context, container containerd.Container, cioAttach cio.Attach) (containerd.Task, error) {
						time.Sleep(10 * time.Second)
						return nil, nil
					}).Times(3)
				return nil, log.NewErrorf("failed to connect to shim for container id = %s", containerID)
			},
		},
		"test_create_task_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error) {
				ioMgrMock.EXPECT().NewCioAttach(containerID).Return(testCioAttach)
				err := log.NewError("test error")
				spiMock.EXPECT().LoadTask(gomock.Any(), mockContainer, matchers.MatchesCioAttach(testCioAttach)).Return(nil, err)
				return nil, err
			},
		},
		"test_task_not_found_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error) {
				ioMgrMock.EXPECT().NewCioAttach(containerID).Return(testCioAttach)
				err := log.NewErrorf("task for containerd container id = %s not found - container is also deleted", containerID)
				spiMock.EXPECT().LoadTask(gomock.Any(), mockContainer, matchers.MatchesCioAttach(testCioAttach)).Return(nil, errdefs.ErrNotFound)
				return nil, err
			},
		},
		"test_task_wait_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error) {
				ioMgrMock.EXPECT().NewCioAttach(containerID).Return(testCioAttach)
				err := log.NewErrorf("test error")
				taskMock := mocksContainerd.NewMockTask(ctrl)
				spiMock.EXPECT().LoadTask(gomock.Any(), mockContainer, matchers.MatchesCioAttach(testCioAttach)).Return(taskMock, nil)
				taskMock.EXPECT().Wait(gomock.Any()).Return(nil, err)
				return nil, err
			},
		},
		"test_no_error": {
			mockExec: func(spiMock *mocksCtrd.MockcontainerdSpi, ioMgrMock *MockcontainerIOManager, mockContainer *mocksContainerd.MockContainer, ctrl *gomock.Controller) (*containerInfo, error) {
				ioMgrMock.EXPECT().NewCioAttach(containerID).Return(testCioAttach)
				taskMock := mocksContainerd.NewMockTask(ctrl)
				spiMock.EXPECT().LoadTask(gomock.Any(), mockContainer, matchers.MatchesCioAttach(testCioAttach)).Return(taskMock, nil)
				taskMock.EXPECT().Wait(gomock.Any()).Return(testStatusChan, nil)
				return &containerInfo{
					container:     mockContainer,
					task:          taskMock,
					statusChannel: testStatusChan,
					resultChannel: make(chan exitInfo),
				}, nil
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			ioMgrMock := NewMockcontainerIOManager(ctrl)
			ctrdClient := &containerdClient{
				spi:   spiMock,
				ioMgr: ioMgrMock,
			}
			containerMock := mocksContainerd.NewMockContainer(ctrl)
			expectedCtrInfo, expectedErr := testCaseData.mockExec(spiMock, ioMgrMock, containerMock, ctrl)
			actualCtrInfo, actualErr := ctrdClient.loadTask(context.TODO(), containerID, checkpointDir, containerMock)

			testutil.AssertError(t, expectedErr, actualErr)
			if expectedCtrInfo != nil {
				testutil.AssertNotNil(t, expectedCtrInfo)
				testutil.AssertEqual(t, expectedCtrInfo.task, actualCtrInfo.task)
				testutil.AssertEqual(t, expectedCtrInfo.container, actualCtrInfo.container)
				testutil.AssertEqual(t, expectedCtrInfo.statusChannel, actualCtrInfo.statusChannel)
				testutil.AssertNotNil(t, actualCtrInfo.resultChannel)
			} else {
				testutil.AssertNil(t, actualCtrInfo)
			}
		})
	}
}

func TestClientInternalInitLogDriver(t *testing.T) {
	const containerID = "test-container-id"
	container := &types.Container{
		ID: containerID,
		HostConfig: &types.HostConfig{
			LogConfig: &types.LogConfiguration{
				ModeConfig: &types.LogModeConfiguration{},
			},
		},
	}
	testCases := map[string]struct {
		mockExec func(logsMgrMock *mocksCtrd.MockcontainerLogsManager, ioMgrMock *MockcontainerIOManager, ctrl *gomock.Controller) error
	}{
		"test_get_log_driver_error": {
			mockExec: func(logsMgrMock *mocksCtrd.MockcontainerLogsManager, ioMgrMock *MockcontainerIOManager, ctrl *gomock.Controller) error {
				err := log.NewError("test error")
				logsMgrMock.EXPECT().GetLogDriver(container).Return(nil, err)
				return err
			},
		},
		"test_configure_io_error": {
			mockExec: func(logsMgrMock *mocksCtrd.MockcontainerLogsManager, ioMgrMock *MockcontainerIOManager, ctrl *gomock.Controller) error {
				logDriverMock := mocksLogger.NewMockLogDriver(ctrl)
				logsMgrMock.EXPECT().GetLogDriver(container).Return(logDriverMock, nil)
				err := log.NewError("test error")
				ioMgrMock.EXPECT().ConfigureIO(container.ID, logDriverMock, container.HostConfig.LogConfig.ModeConfig).Return(err)
				return err
			},
		},
		"test_no_error": {
			mockExec: func(logsMgrMock *mocksCtrd.MockcontainerLogsManager, ioMgrMock *MockcontainerIOManager, ctrl *gomock.Controller) error {
				logDriverMock := mocksLogger.NewMockLogDriver(ctrl)
				logsMgrMock.EXPECT().GetLogDriver(container).Return(logDriverMock, nil)
				ioMgrMock.EXPECT().ConfigureIO(container.ID, logDriverMock, container.HostConfig.LogConfig.ModeConfig).Return(nil)
				return nil
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ioMgrMock := NewMockcontainerIOManager(ctrl)
			logsMgrMock := mocksCtrd.NewMockcontainerLogsManager(ctrl)
			ctrdClient := &containerdClient{
				ioMgr:   ioMgrMock,
				logsMgr: logsMgrMock,
			}
			expectedErr := testCaseData.mockExec(logsMgrMock, ioMgrMock, ctrl)
			actualErr := ctrdClient.initLogDriver(container)
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}

func TestClientInternalKillTask(t *testing.T) {
	const containerID = "test-container-id"
	container := &types.Container{
		ID: containerID,
	}
	testCases := map[string]struct {
		stopOpts *types.StopOpts
		mockExec func(taskMock *mocksContainerd.MockTask, resultChan chan exitInfo) (int64, error)
	}{
		"test_sigkill": {
			stopOpts: &types.StopOpts{
				Signal: "SIGKILL",
			},
			mockExec: func(taskMock *mocksContainerd.MockTask, resultChan chan exitInfo) (int64, error) {
				err := log.NewError("test error")
				taskMock.EXPECT().Kill(gomock.Any(), syscall.SIGKILL, matchers.MatchesTaskKillOpts(containerd.WithKillAll)).Return(err)
				return -1, err
			},
		},
		"test_sigterm_kill_error": {
			stopOpts: &types.StopOpts{
				Signal: "SIGTERM",
			},
			mockExec: func(taskMock *mocksContainerd.MockTask, resultChan chan exitInfo) (int64, error) {
				err := log.NewError("test error")
				taskMock.EXPECT().Kill(gomock.Any(), syscall.SIGTERM).Return(err)
				return -1, err
			},
		},
		"test_custom_signal_no_error": {
			stopOpts: &types.StopOpts{
				Signal:  "123",
				Timeout: 30,
			},
			mockExec: func(taskMock *mocksContainerd.MockTask, resultChan chan exitInfo) (int64, error) {
				taskMock.EXPECT().Kill(gomock.Any(), syscall.Signal(123)).Return(nil)
				resultChan <- exitInfo{exitCode: 0, exitTime: time.Now(), exitError: nil}
				return 0, nil
			},
		},
		"test_sigterm_kill_not_forced_timeout": {
			stopOpts: &types.StopOpts{
				Signal:  "SIGTERM",
				Timeout: 1,
				Force:   false,
			},
			mockExec: func(taskMock *mocksContainerd.MockTask, resultChan chan exitInfo) (int64, error) {
				taskMock.EXPECT().Kill(gomock.Any(), syscall.SIGTERM).
					Do(func(context.Context, syscall.Signal, ...containerd.KillOpts) error {
						time.Sleep(3 * time.Second)
						return nil
					})
				err := log.NewErrorf("could not stop container with ID = %s with %s", container.ID, "SIGTERM")
				return -1, err
			},
		},
		"test_sigterm_kill_forced_timeout": {
			stopOpts: &types.StopOpts{
				Signal:  "SIGTERM",
				Timeout: 1,
				Force:   true,
			},
			mockExec: func(taskMock *mocksContainerd.MockTask, resultChan chan exitInfo) (int64, error) {
				taskMock.EXPECT().Kill(gomock.Any(), syscall.SIGTERM).
					Do(func(context.Context, syscall.Signal, ...containerd.KillOpts) error {
						time.Sleep(3 * time.Second)
						return nil
					})
				err := log.NewError("test error")
				taskMock.EXPECT().Kill(gomock.Any(), syscall.SIGKILL, matchers.MatchesTaskKillOpts(containerd.WithKillAll)).Return(err)
				return -1, err
			},
		},
	}

	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			taskMock := mocksContainerd.NewMockTask(ctrl)
			ctrInfo := &containerInfo{
				c:             container,
				task:          taskMock,
				resultChannel: make(chan exitInfo, 1),
			}
			ctrdClient := &containerdClient{}

			expectedCode, expectedErr := testCaseData.mockExec(taskMock, ctrInfo.resultChannel)
			actualCode, _, actualErr := ctrdClient.killTask(context.TODO(), ctrInfo, testCaseData.stopOpts)

			testutil.AssertError(t, expectedErr, actualErr)
			testutil.AssertEqual(t, expectedCode, actualCode)
		})
	}
}

func TestClientInternalKillTaskForced(t *testing.T) {
	const (
		containerID                        = "test-container-id"
		generalSigKillTimeoutConfiguration = 3 * time.Second
	)
	container := &types.Container{
		ID: containerID,
	}
	testCases := map[string]struct {
		mockExec func(taskMock *mocksContainerd.MockTask, resultChan chan exitInfo) (int64, error)
	}{
		"test_kill_error": {
			mockExec: func(taskMock *mocksContainerd.MockTask, resultChan chan exitInfo) (int64, error) {
				err := log.NewError("test error")
				taskMock.EXPECT().Kill(gomock.Any(), syscall.SIGKILL, matchers.MatchesTaskKillOpts(containerd.WithKillAll)).Return(err)
				return -1, err
			},
		},
		"test_kill_no_error": {
			mockExec: func(taskMock *mocksContainerd.MockTask, resultChan chan exitInfo) (int64, error) {
				taskMock.EXPECT().Kill(gomock.Any(), syscall.SIGKILL, matchers.MatchesTaskKillOpts(containerd.WithKillAll)).
					Return(nil)
				resultChan <- exitInfo{
					exitCode:  0,
					exitError: nil,
					exitTime:  time.Now(),
				}
				return 0, nil
			},
		},
		"test_sigterm_kill_timeout": {
			mockExec: func(taskMock *mocksContainerd.MockTask, resultChan chan exitInfo) (int64, error) {
				taskMock.EXPECT().Kill(gomock.Any(), syscall.SIGKILL, matchers.MatchesTaskKillOpts(containerd.WithKillAll)).Return(nil)
				return -1, log.NewErrorf("could not stop container with ID = %s with SIGKILL", container.ID)
			},
		},
	}

	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			taskMock := mocksContainerd.NewMockTask(ctrl)
			ctrInfo := &containerInfo{
				c:             container,
				task:          taskMock,
				resultChannel: make(chan exitInfo, 1),
			}
			ctrdClient := &containerdClient{}

			expectedCode, expectedErr := testCaseData.mockExec(taskMock, ctrInfo.resultChannel)
			actualCode, _, actualErr := ctrdClient.killTaskForced(context.TODO(), ctrInfo, generalSigKillTimeoutConfiguration)

			testutil.AssertError(t, expectedErr, actualErr)
			testutil.AssertEqual(t, expectedCode, actualCode)
		})
	}
}

func TestClientInternalProcessEvents(t *testing.T) {
	const (
		namespace        = "test-namespace"
		eventsFilter     = "namespace==" + namespace + ",topic~=tasks/oom.*"
		containerID      = "test-container-id"
		testCasesTimeout = 5 * time.Second
	)
	testCases := map[string]struct {
		event    *events.Envelope
		mockExec func(ioMock *mocksIo.MockWriteCloser, spiMock *mocksCtrd.MockcontainerdSpi) (*sync.WaitGroup, bool)
	}{
		"test_event_wrong_topic": {
			mockExec: func(ioMock *mocksIo.MockWriteCloser, spiMock *mocksCtrd.MockcontainerdSpi) (*sync.WaitGroup, bool) {
				eventsChan := make(chan *events.Envelope, 1)
				errChan := make(chan error, 1)

				spiMock.EXPECT().Subscribe(gomock.Any(), eventsFilter).Return(eventsChan, errChan)

				wg := &sync.WaitGroup{}
				wg.Add(1)

				event := &events.Envelope{
					Topic:     runtime.TaskStartEventTopic,
					Namespace: namespace,
				}
				eventsChan <- event

				ioMock.EXPECT().Write(matchers.ContainsString(fmt.Sprintf("skip envelope with topic %s:", event.Topic))).
					Do(func(p []byte) (int, error) {
						wg.Done()
						return 0, nil
					})
				return wg, false
			},
		},
		"test_event_wrong_namespace": {
			mockExec: func(ioMock *mocksIo.MockWriteCloser, spiMock *mocksCtrd.MockcontainerdSpi) (*sync.WaitGroup, bool) {
				eventsChan := make(chan *events.Envelope, 1)
				errChan := make(chan error, 1)

				spiMock.EXPECT().Subscribe(gomock.Any(), eventsFilter).Return(eventsChan, errChan)

				wg := &sync.WaitGroup{}
				wg.Add(1)

				event := &events.Envelope{
					Topic:     runtime.TaskOOMEventTopic,
					Namespace: "some-random-ns",
				}
				eventsChan <- event

				ioMock.EXPECT().Write(matchers.ContainsString(fmt.Sprintf("skip envelope with topic %s:", event.Topic))).
					Do(func(p []byte) (int, error) {
						wg.Done()
						return 0, nil
					})
				return wg, false
			},
		},
		"test_event_unmarshal_error": {
			mockExec: func(ioMock *mocksIo.MockWriteCloser, spiMock *mocksCtrd.MockcontainerdSpi) (*sync.WaitGroup, bool) {
				eventsChan := make(chan *events.Envelope, 1)
				errChan := make(chan error, 1)

				spiMock.EXPECT().Subscribe(gomock.Any(), eventsFilter).Return(eventsChan, errChan)

				wg := &sync.WaitGroup{}
				wg.Add(1)
				event := &events.Envelope{
					Namespace: namespace,
					Topic:     runtime.TaskOOMEventTopic,
					Event:     &protoTypes.Any{TypeUrl: "random"},
				}
				eventsChan <- event

				ioMock.EXPECT().Write(matchers.ContainsString(fmt.Sprintf("failed to unmarshal envelope %s:", event.Topic))).
					Do(func(p []byte) (int, error) {
						wg.Done()
						return 0, nil
					})
				return wg, false
			},
		},
		"test_event_wrong_type": {
			mockExec: func(ioMock *mocksIo.MockWriteCloser, spiMock *mocksCtrd.MockcontainerdSpi) (*sync.WaitGroup, bool) {
				eventsChan := make(chan *events.Envelope, 1)
				errChan := make(chan error, 1)

				spiMock.EXPECT().Subscribe(gomock.Any(), eventsFilter).Return(eventsChan, errChan)

				wg := &sync.WaitGroup{}
				wg.Add(1)

				e, _ := typeurl.MarshalAny(&eventstypes.ContainerCreate{})
				event := &events.Envelope{
					Namespace: namespace,
					Topic:     runtime.TaskOOMEventTopic,
					Event:     e,
				}
				eventsChan <- event

				ioMock.EXPECT().Write(matchers.ContainsString(fmt.Sprintf("failed to parse %s envelope:", event.Topic))).
					Do(func(p []byte) (int, error) {
						wg.Done()
						return 0, nil
					})
				return wg, false
			},
		},
		"test_event_missing_container_cache": {
			mockExec: func(ioMock *mocksIo.MockWriteCloser, spiMock *mocksCtrd.MockcontainerdSpi) (*sync.WaitGroup, bool) {
				eventsChan := make(chan *events.Envelope, 1)
				errChan := make(chan error, 1)

				spiMock.EXPECT().Subscribe(gomock.Any(), eventsFilter).Return(eventsChan, errChan)

				wg := &sync.WaitGroup{}
				wg.Add(1)

				e := &eventstypes.TaskOOM{ContainerID: "some-random-id"}
				eBytes, _ := typeurl.MarshalAny(e)
				event := &events.Envelope{
					Namespace: namespace,
					Topic:     runtime.TaskOOMEventTopic,
					Event:     eBytes,
				}
				eventsChan <- event

				ioMock.EXPECT().Write(matchers.ContainsString(fmt.Sprintf("missing container info for container - %s", e.ContainerID))).
					Do(func(p []byte) (int, error) {
						wg.Done()
						return 0, nil
					})
				return wg, false
			},
		},
		"test_event_oom_correct": {
			mockExec: func(ioMock *mocksIo.MockWriteCloser, spiMock *mocksCtrd.MockcontainerdSpi) (*sync.WaitGroup, bool) {
				eventsChan := make(chan *events.Envelope, 1)
				errChan := make(chan error, 1)

				spiMock.EXPECT().Subscribe(gomock.Any(), eventsFilter).Return(eventsChan, errChan)

				wg := &sync.WaitGroup{}
				wg.Add(1)

				e := &eventstypes.TaskOOM{ContainerID: containerID}
				eBytes, _ := typeurl.MarshalAny(e)
				event := &events.Envelope{
					Namespace: namespace,
					Topic:     runtime.TaskOOMEventTopic,
					Event:     eBytes,
				}
				eventsChan <- event

				ioMock.EXPECT().Write(matchers.ContainsString(fmt.Sprintf("updated info cache for container ID = %s with OOM killed = true", e.ContainerID))).
					Do(func(p []byte) (int, error) {
						wg.Done()
						return 0, nil
					})
				return wg, true
			},
		},
		"test_event_error_received": {
			mockExec: func(ioMock *mocksIo.MockWriteCloser, spiMock *mocksCtrd.MockcontainerdSpi) (*sync.WaitGroup, bool) {
				eventsChan := make(chan *events.Envelope, 1)
				errChan := make(chan error, 1)

				spiMock.EXPECT().Subscribe(gomock.Any(), eventsFilter).Return(eventsChan, errChan)

				wg := &sync.WaitGroup{}
				wg.Add(1)
				errChan <- log.NewError("test error")

				ioMock.EXPECT().Write(matchers.ContainsString("failed to receive envelope:")).
					Do(func(p []byte) (int, error) {
						wg.Done()
						return 0, nil
					})
				return wg, false
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)

			ctrl := gomock.NewController(t)
			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			ctrdClient := &containerdClient{
				spi:       spiMock,
				ctrdCache: newContainerInfoCache(),
			}
			ctrdClient.ctrdCache.cache[containerID] = &containerInfo{
				c: &types.Container{
					ID: containerID,
				},
			}
			mockIOWriter := mocksIo.NewMockWriteCloser(ctrl)

			logrus.SetLevel(logrus.DebugLevel)
			logrus.SetOutput(mockIOWriter)

			testWg, expectedOOMKilled := testCaseData.mockExec(mockIOWriter, spiMock)

			defer func() {
				ctrdClient.eventsCancel()
				ctrl.Finish()

				logrus.SetLevel(logrus.InfoLevel)
				logrus.SetOutput(os.Stdout)
			}()
			go ctrdClient.processEvents(namespace)

			testutil.AssertWithTimeout(t, testWg, testCasesTimeout)
			testutil.AssertEqual(t, expectedOOMKilled, ctrdClient.ctrdCache.cache[containerID].isOOmKilled())
		})
	}
}

func TestClientInternalIsImageUsed(t *testing.T) {
	testImgRef := "test.image/ref:latest"
	entryDigest := digest.NewDigest(digest.SHA256, sha256.New())

	testCases := map[string]struct {
		mockExec func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) (bool, error)
	}{
		"test_rootfs_error": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) (bool, error) {
				imageMock.EXPECT().Name().Return(testImgRef)
				err := log.NewErrorf("test error")
				imageMock.EXPECT().RootFS(ctx).Return(nil, err)
				return false, err
			},
		},
		"test_not_used": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) (bool, error) {
				imageMock.EXPECT().Name().Return(testImgRef)
				imageMock.EXPECT().RootFS(ctx).Return([]digest.Digest{entryDigest}, nil)
				spiMock.EXPECT().ListSnapshots(ctx, fmt.Sprintf(snapshotsWalkFilterFormat, entryDigest.String())).Return(nil, nil)
				return false, nil
			},
		},
		"test_used": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) (bool, error) {
				imageMock.EXPECT().Name().Return(testImgRef)
				imageMock.EXPECT().RootFS(ctx).Return([]digest.Digest{entryDigest}, nil)
				spiMock.EXPECT().ListSnapshots(ctx, fmt.Sprintf(snapshotsWalkFilterFormat, entryDigest.String())).Return([]snapshots.Info{{}}, log.NewErrorf("test err"))
				return true, nil
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// init mocks
			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			imageMock := mocksContainerd.NewMockImage(ctrl)

			ctx := context.Background()
			ctrdClient := &containerdClient{
				spi: spiMock,
			}
			// mock exec
			expectedIsUsed, expectedErr := testCaseData.mockExec(ctx, spiMock, imageMock)

			isUsed, err := ctrdClient.isImageUsed(ctx, imageMock)
			testutil.AssertError(t, expectedErr, err)
			testutil.AssertEqual(t, expectedIsUsed, isUsed)
		})
	}
}

func TestClientInternalRemoveUnusedImage(t *testing.T) {
	testImgRef := "test.image/ref:latest"
	entryDigest := digest.NewDigest(digest.SHA256, sha256.New())

	testCases := map[string]struct {
		mockExec func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error
	}{
		"test_is_used_error": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				imageMock.EXPECT().Name().Return(testImgRef).Times(2)
				err := log.NewErrorf("test error")
				imageMock.EXPECT().RootFS(ctx).Return(nil, err)
				return err
			},
		},
		"test_used": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				imageMock.EXPECT().Name().Return(testImgRef).Times(2)
				imageMock.EXPECT().RootFS(ctx).Return([]digest.Digest{entryDigest}, nil)
				spiMock.EXPECT().ListSnapshots(ctx, fmt.Sprintf(snapshotsWalkFilterFormat, entryDigest.String())).Return([]snapshots.Info{{}}, nil)
				return nil
			},
		},
		"test_not_used_delete_error": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				imageMock.EXPECT().Name().Return(testImgRef).Times(2)
				imageMock.EXPECT().RootFS(ctx).Return([]digest.Digest{entryDigest}, nil)
				spiMock.EXPECT().ListSnapshots(ctx, fmt.Sprintf(snapshotsWalkFilterFormat, entryDigest.String())).Return(nil, nil)
				err := log.NewError("test error")
				spiMock.EXPECT().DeleteImage(ctx, testImgRef).Return(err)
				return err
			},
		},
		"test_not_used_delete_no_error": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				imageMock.EXPECT().Name().Return(testImgRef).Times(3)
				imageMock.EXPECT().RootFS(ctx).Return([]digest.Digest{entryDigest}, nil)
				spiMock.EXPECT().ListSnapshots(ctx, fmt.Sprintf(snapshotsWalkFilterFormat, entryDigest.String())).Return(nil, nil)
				spiMock.EXPECT().DeleteImage(ctx, testImgRef).Return(nil)
				return nil
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// init mocks
			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			imageMock := mocksContainerd.NewMockImage(ctrl)

			ctx := context.Background()
			ctrdClient := &containerdClient{
				spi: spiMock,
			}
			// mock exec
			expectedErr := testCaseData.mockExec(ctx, spiMock, imageMock)

			err := ctrdClient.removeUnusedImage(ctx, imageMock)
			testutil.AssertError(t, expectedErr, err)
		})
	}
}

func TestClientInternalHandleImageExpired(t *testing.T) {
	testImgRef := "test.image/ref:latest"
	entryDigest := digest.NewDigest(digest.SHA256, sha256.New())

	testCases := map[string]struct {
		mockExec func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error
	}{
		"test_get_image_error": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				err := log.NewErrorf("test error")
				spiMock.EXPECT().GetImage(ctx, testImgRef).Return(nil, err)
				return err
			},
		},
		"test_get_image_not_found_error": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				spiMock.EXPECT().GetImage(ctx, testImgRef).Return(nil, errdefs.ErrNotFound)
				return nil
			},
		},
		"test_unused_error": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				spiMock.EXPECT().GetImage(ctx, testImgRef).Return(imageMock, nil)
				imageMock.EXPECT().Name().Return(testImgRef).Times(2)
				err := log.NewErrorf("test error")
				imageMock.EXPECT().RootFS(ctx).Return(nil, err)
				return err
			},
		},
		"test_no_error": {
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, imageMock *mocksContainerd.MockImage) error {
				spiMock.EXPECT().GetImage(ctx, testImgRef).Return(imageMock, nil)
				imageMock.EXPECT().Name().Return(testImgRef).Times(3)
				imageMock.EXPECT().RootFS(ctx).Return([]digest.Digest{entryDigest}, nil)
				spiMock.EXPECT().ListSnapshots(ctx, fmt.Sprintf(snapshotsWalkFilterFormat, entryDigest.String())).Return(nil, nil)
				spiMock.EXPECT().DeleteImage(ctx, testImgRef).Return(nil)
				return nil
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// init mocks
			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			imageMock := mocksContainerd.NewMockImage(ctrl)

			ctx := context.Background()
			ctrdClient := &containerdClient{
				spi: spiMock,
			}
			// mock exec
			expectedErr := testCaseData.mockExec(ctx, spiMock, imageMock)

			err := ctrdClient.handleImageExpired(ctx, testImgRef)
			testutil.AssertError(t, expectedErr, err)
		})
	}
}

func TestClientInternalManageExpiry(t *testing.T) {
	testImgRef := "test.image/ref:latest"
	entryDigest := digest.NewDigest(digest.SHA256, sha256.New())

	testCases := map[string]struct {
		imagesExpiry time.Duration
		mockExec     func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, watcherMock *MockresourcesWatcher, imageMock *mocksContainerd.MockImage) error
	}{
		"test_expired_error": {
			imagesExpiry: 1 * time.Second,
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, watcherMock *MockresourcesWatcher, imageMock *mocksContainerd.MockImage) error {
				imageMock.EXPECT().Name().Return(testImgRef).Times(3)
				imageMock.EXPECT().Metadata().Return(images.Image{CreatedAt: time.Now().Add(-24 * time.Hour)})
				err := log.NewError("test error")
				imageMock.EXPECT().RootFS(ctx).Return(nil, err)
				return err
			},
		},
		"test_expired_not_used_no_error": {
			imagesExpiry: 1 * time.Second,
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, watcherMock *MockresourcesWatcher, imageMock *mocksContainerd.MockImage) error {
				imageMock.EXPECT().Name().Return(testImgRef).Times(4)
				imageMock.EXPECT().Metadata().Return(images.Image{CreatedAt: time.Now().Add(-24 * time.Hour)})
				imageMock.EXPECT().RootFS(ctx).Return([]digest.Digest{entryDigest}, nil)
				spiMock.EXPECT().ListSnapshots(ctx, fmt.Sprintf(snapshotsWalkFilterFormat, entryDigest.String())).Return(nil, nil)
				spiMock.EXPECT().DeleteImage(ctx, testImgRef).Return(nil)
				return nil
			},
		},
		"test_expired_used_no_error": {
			imagesExpiry: 1 * time.Second,
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, watcherMock *MockresourcesWatcher, imageMock *mocksContainerd.MockImage) error {
				imageMock.EXPECT().Name().Return(testImgRef).Times(3)
				imageMock.EXPECT().Metadata().Return(images.Image{CreatedAt: time.Now().Add(-24 * time.Hour)})
				imageMock.EXPECT().RootFS(ctx).Return([]digest.Digest{entryDigest}, nil)
				spiMock.EXPECT().ListSnapshots(ctx, fmt.Sprintf(snapshotsWalkFilterFormat, entryDigest.String())).Return([]snapshots.Info{{}}, nil)
				return nil
			},
		},
		"test_not_expired_watch_already_watched_error": {
			imagesExpiry: 5 * time.Hour,
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, watcherMock *MockresourcesWatcher, imageMock *mocksContainerd.MockImage) error {
				imageMock.EXPECT().Name().Return(testImgRef)
				imageMock.EXPECT().Metadata().Return(images.Image{CreatedAt: time.Now().Add(-1 * time.Hour)})
				watcherMock.EXPECT().Watch(testImgRef, gomock.Any(), gomock.Any()).Return(alreadyWatchedError)
				return nil
			},
		},
		"test_not_expired_watch_error": {
			imagesExpiry: 5 * time.Hour,
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, watcherMock *MockresourcesWatcher, imageMock *mocksContainerd.MockImage) error {
				imageMock.EXPECT().Name().Return(testImgRef)
				imageMock.EXPECT().Metadata().Return(images.Image{CreatedAt: time.Now().Add(-1 * time.Hour)})
				err := log.NewError("test error")
				watcherMock.EXPECT().Watch(testImgRef, gomock.Any(), gomock.Any()).Return(err)
				return err
			},
		},
		"test_not_expired_watch_no_error": {
			imagesExpiry: 5 * time.Hour,
			mockExec: func(ctx context.Context, spiMock *mocksCtrd.MockcontainerdSpi, watcherMock *MockresourcesWatcher, imageMock *mocksContainerd.MockImage) error {
				imageMock.EXPECT().Name().Return(testImgRef)
				imageMock.EXPECT().Metadata().Return(images.Image{CreatedAt: time.Now().Add(-1 * time.Hour)})
				watcherMock.EXPECT().Watch(testImgRef, gomock.Any(), gomock.Any()).Return(nil)
				return nil
			},
		},
	}
	for testCaseName, testCaseData := range testCases {
		t.Run(testCaseName, func(t *testing.T) {
			t.Log(testCaseName)

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// init mocks
			spiMock := mocksCtrd.NewMockcontainerdSpi(ctrl)
			watcherMock := NewMockresourcesWatcher(ctrl)
			imageMock := mocksContainerd.NewMockImage(ctrl)

			ctx := context.Background()
			ctrdClient := &containerdClient{
				spi:           spiMock,
				imagesWatcher: watcherMock,
				imageExpiry:   testCaseData.imagesExpiry,
			}
			// mock exec
			expectedErr := testCaseData.mockExec(ctx, spiMock, watcherMock, imageMock)

			err := ctrdClient.manageImageExpiry(ctx, imageMock)
			testutil.AssertError(t, expectedErr, err)
		})
	}
}
