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
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/imgcrypt"
	"github.com/containerd/imgcrypt/images/encryption"
	"github.com/containers/ocicrypt/config"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/matchers"
	mocksContainerd "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	mocksCtrd "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctrd"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/golang/mock/gomock"
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
