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

// GrpcServerOpt represents the available configuration options type for the daemon's gRPC server initialization
type GrpcServerOpt func(netOpts *grpcServerOpts) error

type grpcServerOpts struct {
	network     string
	addressPath string
}

func applyOptsGrpcServer(grpcServerOpts *grpcServerOpts, opts ...GrpcServerOpt) error {
	for _, o := range opts {
		if err := o(grpcServerOpts); err != nil {
			return err
		}
	}
	return nil
}

// WithGrpcServerNetwork configures the communication protocol to be used for accessing the server
func WithGrpcServerNetwork(network string) GrpcServerOpt {
	return func(grpcServerOpts *grpcServerOpts) error {
		grpcServerOpts.network = network
		return nil
	}
}

// WithGrpcServerAddressPath configures the address path for communicating with the gRPC server
func WithGrpcServerAddressPath(addressPath string) GrpcServerOpt {
	return func(grpcServerOpts *grpcServerOpts) error {
		grpcServerOpts.addressPath = addressPath
		return nil
	}
}
