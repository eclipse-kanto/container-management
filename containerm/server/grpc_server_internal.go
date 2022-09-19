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

package server

import (
	"net"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	// GrpcServerServiceID respresens the Service ID of the GRPC server
	GrpcServerServiceID = "container-management.server.grpc.v1.service-grpc-server"
)

func init() {
	registry.Register(&registry.Registration{
		ID:       GrpcServerServiceID,
		Type:     registry.GRPCServer,
		InitFunc: registryInit,
	})
}

type grpcServer struct {
	network     string //unix
	addressPath string // /run/container-management/container-management.sock
	services    []registry.GrpcService
	server      *grpc.Server
}

func (grpcServer *grpcServer) Start() error {

	lis, err := net.Listen(grpcServer.network, grpcServer.addressPath)
	if err != nil {
		log.ErrorErr(err, "gRPC server with service ID = %s failed to establish connection on the provided address : %s://%s ", GrpcServerServiceID, grpcServer.network, grpcServer.addressPath)
		return err
	}
	// Register reflection service on gRPC server.
	reflection.Register(grpcServer.server)
	go func() {
		if err = grpcServer.server.Serve(lis); err != nil {
			log.ErrorErr(err, "gRPC server with service ID = %s failed ", GrpcServerServiceID)
			return
		}
		log.Debug("exiting gRPC server thread")
	}()

	return nil
}

func (grpcServer *grpcServer) Stop() error {
	log.Debug("stopping gRPC server instance with ID = %s", GrpcServerServiceID)
	grpcServer.server.Stop()
	log.Debug("successfully stopped gRPC server instance with ID = %s", GrpcServerServiceID)
	return nil

}
