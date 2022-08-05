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
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/leases"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

// GetImage returns a locally existing image
func (spi *ctrdSpi) GetImage(ctx context.Context, imageRef string) (containerd.Image, error) {
	ctx = spi.setContext(ctx, false)
	return spi.client.GetImage(ctx, imageRef)
}

// PullImage downloads the provided content and returns an image object
func (spi *ctrdSpi) PullImage(ctx context.Context, imageRef string, opts ...containerd.RemoteOpt) (containerd.Image, error) {
	ctx = spi.setContext(ctx, false)
	lease, err := spi.leaseService.Create(ctx, leases.WithID(imageRef))
	if err != nil {
		if !errdefs.IsAlreadyExists(err) {
			log.ErrorErr(err, "error creating lease with ID = %s", imageRef)
			return nil, err
		}
	}
	ctx = leases.WithLease(ctx, lease.ID)
	return spi.client.Pull(ctx, imageRef, opts...)
}

// UnpackImage unpacks the contents of the provided image locally
func (spi *ctrdSpi) UnpackImage(ctx context.Context, image containerd.Image, opts ...containerd.UnpackOpt) error {
	// NB! Do not use leases when unpacking to prevent memory leaks due to unreachable but leased unpacked content
	ctx = spi.setContext(ctx, false)
	return image.Unpack(ctx, spi.snapshotterType, opts...)
}

// ListImages returns all locally existing images
func (spi *ctrdSpi) ListImages(ctx context.Context) ([]containerd.Image, error) {
	ctx = spi.setContext(ctx, false)
	return spi.client.ListImages(ctx)
}

// DeleteImage removes the contents of the provided image from the disk
func (spi *ctrdSpi) DeleteImage(ctx context.Context, imageRef string) error {
	ctx = spi.setContext(ctx, false)

	lease := leases.Lease{ID: imageRef}
	if err := spi.leaseService.Delete(ctx, lease); err != nil {
		return err
	}
	return spi.client.ImageService().Delete(ctx, imageRef)
}
