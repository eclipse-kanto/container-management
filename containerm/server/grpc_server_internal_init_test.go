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

package server

import (
	"context"
	"testing"

	"github.com/containerd/containerd/errdefs"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	mockGrpc "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/registry"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
)

func TestRegistryInit(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	grpcServiceMock := mockGrpc.NewMockGrpcService(ctrl)
	testRegValid := &registry.Registration{
		Type: registry.GRPCService,
		ID:   "test.grpc.service.id",
		InitFunc: func(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
			return grpcServiceMock, nil
		},
	}

	testRegErr := &registry.Registration{
		Type: registry.GRPCService,
		ID:   "test.grpc.service.id",
		InitFunc: func(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
			return nil, errors.New("Instance error")
		},
	}

	sInfo := testRegValid.Init(&registry.ServiceRegistryContext{})
	serviceSet := registry.NewServiceInfoSet()

	sInfoErr := testRegErr.Init(&registry.ServiceRegistryContext{})
	errServiceSet := registry.NewServiceInfoSet()
	errServiceSet.Add(sInfoErr)
	serviceSet.Add(sInfo)
	emptyServiceSet := registry.NewServiceInfoSet()
	grpcServiceMock.EXPECT().Register(gomock.Any()).Times(1).Return(nil)

	testCases := map[string]struct {
		arg            *registry.ServiceRegistryContext
		wantNewtork    string
		wantAdressPath string
		expectedErr    error
	}{
		"register_without_err": {
			arg: registry.NewContext(
				context.Background(),
				[]GrpcServerOpt{WithGrpcServerNetwork("bridge"), WithGrpcServerAddressPath("/run/test.sock")},
				testRegValid,
				serviceSet,
			),
			wantNewtork:    "bridge",
			wantAdressPath: "/run/test.sock",
			expectedErr:    nil,
		},
		"register_with_err_empty_service_set": {
			arg: registry.NewContext(
				context.Background(),
				[]GrpcServerOpt{WithGrpcServerNetwork("bridge"), WithGrpcServerAddressPath("/run/test.sock")},
				testRegValid,
				emptyServiceSet,
			),
			expectedErr: errors.Wrapf(errdefs.ErrNotFound, "no services registered for %s", registry.GRPCService),
		},
		"register_with_err_on_instance": {
			arg: registry.NewContext(
				context.Background(),
				[]GrpcServerOpt{WithGrpcServerNetwork("bridge"), WithGrpcServerAddressPath("/run/test.sock")},
				testRegErr,
				errServiceSet,
			),
			wantNewtork:    "bridge",
			wantAdressPath: "/run/test.sock",
			expectedErr:    nil,
		},
	}
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			result, err := registryInit(testCase.arg)

			testutil.AssertError(t, testCase.expectedErr, err)

			if testCase.expectedErr == nil {
				grpcSrv, ok := result.(*grpcServer)
				testutil.AssertTrue(t, ok)
				testutil.AssertNotNil(t, grpcSrv)
				testutil.AssertEqual(t, testCase.wantNewtork, grpcSrv.network)
				testutil.AssertEqual(t, testCase.wantAdressPath, grpcSrv.addressPath)
			}

		})
	}

}
