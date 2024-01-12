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

package mgr

import (
	"context"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/streams"
)

// ContainerManager represents the container manager abstraction
type ContainerManager interface {

	// Load loads the containers from the persistent memory cache
	Load(ctx context.Context) error

	// Restore recovers containers that are active in the underlying container management system
	Restore(ctx context.Context) error

	// Create creates a new container
	Create(ctx context.Context, config *types.Container) (*types.Container, error)

	// Get the detailed information about a container
	Get(ctx context.Context, id string) (*types.Container, error)

	// List returns the list of available containers
	List(ctx context.Context) ([]*types.Container, error)

	// Start starts a container that has been stopped or created
	Start(ctx context.Context, id string) error

	// Attach attaches the container's IO
	Attach(ctx context.Context, id string, attachConfig *streams.AttachConfig) error

	// Stop stops a running container
	Stop(ctx context.Context, id string, stopOpts *types.StopOpts) error

	// Update updates a running container
	Update(ctx context.Context, id string, updateOpts *types.UpdateOpts) error

	// Restart restarts a running container
	Restart(ctx context.Context, id string, timeout int64) error

	// Pause pauses a running container
	Pause(ctx context.Context, id string) error

	// Unpause resumes a paused container
	Unpause(ctx context.Context, id string) error

	// Rename renames a container
	Rename(ctx context.Context, id string, name string) error

	// Remove removes a container, it may be running or stopped and so on
	Remove(ctx context.Context, id string, force bool, stopOpts *types.StopOpts) error

	// Metrics retrieves metrics data about a container
	Metrics(ctx context.Context, id string) (*types.Metrics, error)

	// Dispose stops and disposes the network manager
	Dispose(ctx context.Context) error
}
