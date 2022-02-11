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

package ctr

import (
	"context"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
)

func (spi *ctrdSpi) CreateContainer(ctx context.Context, containerID string, opts ...containerd.NewContainerOpts) (containerd.Container, error) {
	ctx = spi.setContext(ctx, true)
	return spi.client.NewContainer(ctx, containerID, opts...)
}

func (spi *ctrdSpi) LoadContainer(ctx context.Context, containerID string) (containerd.Container, error) {
	ctx = spi.setContext(ctx, true)
	return spi.client.LoadContainer(ctx, containerID)
}

func (spi *ctrdSpi) CreateTask(ctx context.Context, container containerd.Container, cioCreatorFunc cio.Creator, opts ...containerd.NewTaskOpts) (containerd.Task, error) {
	ctx = spi.setContext(ctx, true)
	return container.NewTask(ctx, cioCreatorFunc, opts...)
}

func (spi *ctrdSpi) LoadTask(ctx context.Context, container containerd.Container, cioReattachFunc cio.Attach) (containerd.Task, error) {
	ctx = spi.setContext(ctx, true)
	return container.Task(ctx, cioReattachFunc)
}
