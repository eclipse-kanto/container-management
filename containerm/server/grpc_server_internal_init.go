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
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"google.golang.org/grpc"
)

func newGrpcServer(network string, addressPath string, grpcServices []registry.GrpcService) (registry.GrpcServer, error) {
	server := grpc.NewServer()

	for _, grpcService := range grpcServices {
		grpcService.Register(server)
	}

	return &grpcServer{network: network, addressPath: addressPath, services: grpcServices, server: server}, nil
}

func registryInit(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
	grpcServerInitOpts := registryCtx.Config.([]GrpcServerOpt)
	var (
		grpcServerOpts = &grpcServerOpts{}
	)
	applyOptsGrpcServer(grpcServerOpts, grpcServerInitOpts...)

	allGrpcServices, err := registryCtx.GetByType(registry.GRPCService)
	if err != nil {
		return nil, err
	}

	grpcServices := []registry.GrpcService{}
	for _, grpcServiceInfo := range allGrpcServices {
		if grpcServiceInfo == nil {
			// No current use case for this "if", will be tested when needed
			nilServiceInfoErr := log.NewErrorf("Service info is nil!")
			log.ErrorErr(nilServiceInfoErr, "Service will not be registered!")
			continue
		}
		grpcService, err := grpcServiceInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "Service will not be registered due to initialization error!")
			continue
		}
		grpcServices = append(grpcServices, grpcService.(registry.GrpcService))
	}

	return newGrpcServer(grpcServerOpts.network, grpcServerOpts.addressPath, grpcServices)

}
