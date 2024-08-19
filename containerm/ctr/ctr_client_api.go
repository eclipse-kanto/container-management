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
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/streams"
)

// ContainerExitHook represents a hook for clearing logic to the containers resources management on exit
type ContainerExitHook func(*types.Container, int64, error, bool, func() error) error

// ContainerAPIClient provides access to containerd container features
type ContainerAPIClient interface {

	// RemoveContainerImage remove image assoicated with container
	RemoveContainerImage(ctx context.Context, image types.Image) error

	// DestroyContainer kill container and delete it
	DestroyContainer(ctx context.Context, container *types.Container, stopOpts *types.StopOpts, clearIOs bool) (int64, time.Time, error)

	// CreateContainer creates all resources needed in the underlying container management so that a container can be successfully started
	CreateContainer(ctx context.Context, container *types.Container, checkpointDir string) error

	// StartContainer starts the underlying container
	StartContainer(ctx context.Context, container *types.Container, checkpointDir string) (int64, error)

	// GetContainerStats retrieves container stats - CPU, memory, IO and PIDs number, and the time they have been measured
	GetContainerStats(ctx context.Context, container *types.Container) (*types.CPUStats, *types.MemoryStats, *types.IOStats, uint64, time.Time, error)

	// AttachContainer attaches to the container's IO
	AttachContainer(ctx context.Context, container *types.Container, attachConfig *streams.AttachConfig) error

	// PauseContainer pauses a container
	PauseContainer(ctx context.Context, container *types.Container) error

	// UnpauseContainer unpauses a container
	UnpauseContainer(ctx context.Context, container *types.Container) error

	// RestoreContainer restores the container information from the underlying container management client along with initialization of all needed resources - streams, etc.
	RestoreContainer(ctx context.Context, container *types.Container) error

	// Dispose manages stop and dispose the containers client
	Dispose(ctx context.Context) error

	// ListContainers lists all created containers
	ListContainers(ctx context.Context) ([]*types.Container, error)

	// GetContainerInfo provides detailed information about a container
	GetContainerInfo(ctx context.Context, id string) (*types.Container, error)

	// ReleaseContainerResources releases all locally allocated resources for the container by the client implementation - e.g. streams, etc.
	ReleaseContainerResources(ctx context.Context, container *types.Container) error

	// SetContainerExitHooks provides access for hooking a clearing logic to the containers resources management on exit
	SetContainerExitHooks(hooks ...ContainerExitHook)

	//UpdateContainer updates container resource limits
	UpdateContainer(ctx context.Context, container *types.Container, resources *types.Resources) error
}
