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
	"github.com/containerd/containerd/remotes"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

// GetImage returns a locally existing image
func (spi *ctrdSpi) GetImage(ctx context.Context, imageRef string) (containerd.Image, error) {
	ctx = spi.setContext(ctx, true)
	return spi.client.GetImage(ctx, imageRef)
}

// PullImage pulls and unpacks an image locally
func (spi *ctrdSpi) PullImage(ctx context.Context, imageRef string, resolver remotes.Resolver) (containerd.Image, error) {
	ctx = spi.setContext(ctx, true)
	options := []containerd.RemoteOpt{
		containerd.WithSchema1Conversion,
		containerd.WithPullSnapshotter(spi.snapshotterType),
		containerd.WithPullUnpack,
	}
	if resolver != nil {
		options = append(options, containerd.WithResolver(resolver))
	} else {
		log.Warn("the default resolver by containerd will be used for image %s", imageRef)
	}

	return spi.client.Pull(ctx, imageRef, options...)
}
