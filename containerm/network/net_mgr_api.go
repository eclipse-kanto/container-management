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

package network

import (
	"context"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
)

// ContainerNetworkManager  abstracts container's network operations
type ContainerNetworkManager interface {
	// Manage performs any container network initialization operations that are needed so that a container is connectable afterwards
	Manage(ctx context.Context, container *types.Container) error

	// Connect is used to connect a container to a network.
	Connect(ctx context.Context, containers *types.Container) error

	// Disconnect disconnects the given container from given network
	Disconnect(ctx context.Context, container *types.Container, force bool) error

	// ReleaseNetworkResources releases all locally allocated resources for the container by the network manager implementation - e.g. network endpoints, etc.
	ReleaseNetworkResources(ctx context.Context, container *types.Container) error

	// Dispose manages stop and dispose of the network manager
	Dispose(ctx context.Context) error

	// Restore restores all networking resources for all running containers
	Restore(ctx context.Context, container []*types.Container) error

	// Initialize initializes all base networks for the manager depending on the modes supported - currently on bridge is supported
	Initialize(ctx context.Context) error
}
