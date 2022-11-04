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

package client

import (
	"net"
	"net/url"
	"time"

	pbcontainers "github.com/eclipse-kanto/container-management/containerm/api/services/containers"
	pbsysinfo "github.com/eclipse-kanto/container-management/containerm/api/services/sysinfo"
	"google.golang.org/grpc"
)

// New creates a new containers client.
func New(connectionAddress string) (Client, error) {
	// Set up a connection to the server.
	gopts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithDialer(getDialer),
	}

	conn, err := grpc.Dial(connectionAddress, gopts...)
	if err != nil {
		return nil, err
	}
	return newContainersClient(conn)
}

func newContainersClient(conn *grpc.ClientConn) (Client, error) {
	pbClient := pbcontainers.NewContainersClient(conn)
	pbVersion := pbsysinfo.NewSystemInfoClient(conn)
	return &client{
		connection:           conn,
		grpcContainersClient: pbClient,
		grpcSystemInfoClient: pbVersion,
	}, nil
}

func getDialer(addr string, duration time.Duration) (net.Conn, error) {
	url, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	switch url.Scheme {
	case "tcp", "tcp4", "tcp6":
		return net.DialTimeout(url.Scheme, url.Host, duration)
	default:
		return unixConnect(addr)
	}
}

func unixConnect(addr string) (net.Conn, error) {
	unixAddr, err := net.ResolveUnixAddr("unix", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialUnix("unix", nil, unixAddr)
	return conn, err
}
