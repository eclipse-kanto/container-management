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
	"context"
	"io"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	sysinfotypes "github.com/eclipse-kanto/container-management/containerm/sysinfo/types"
)

// Filter returns if the container matches the defined filter.
type Filter func(container *types.Container) bool

// Client is the client API for gRPC API of the engine.
type Client interface {
	// Create a new container.
	Create(ctx context.Context, config *types.Container) (*types.Container, error)

	// Get the detailed information of container.
	Get(ctx context.Context, id string) (*types.Container, error)

	// List returns the list of containers matching the optional filters provided.
	List(ctx context.Context, filters ...Filter) ([]*types.Container, error)

	// Start a container.
	Start(ctx context.Context, id string) error

	// Stop a container.
	Stop(ctx context.Context, id string, stopOpts *types.StopOpts) error

	// Update a container.
	Update(ctx context.Context, id string, updateOpts *types.UpdateOpts) error

	// Attach to a container
	Attach(ctx context.Context, id string, stdin bool) (io.Writer, io.ReadCloser, error)

	// Restart restart a running container.
	Restart(ctx context.Context, id string, timeout int64) error

	// Pause a container.
	Pause(ctx context.Context, id string) error

	// Resumes a container.
	Resume(ctx context.Context, id string) error

	// Rename renames a container.
	Rename(ctx context.Context, id string, name string) error

	// Remove removes a container, it may be running or stopped and so on.
	Remove(ctx context.Context, id string, force bool, stopOpts *types.StopOpts) error

	ProjectInfo(ctx context.Context) (sysinfotypes.ProjectInfo, error)

	// Logs prints the logs for a container
	Logs(ctx context.Context, id string, tail int32) error

	// Dispose the client instance
	Dispose() error
}
