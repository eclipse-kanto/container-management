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
	"fmt"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/images"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/opencontainers/image-spec/specs-go/v1"
	"path/filepath"

	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/namespaces"
)

const (
	snapshotIDTemplate = "%s-snapshot"
)

func (spi *ctrdSpi) setContext(ctx context.Context, withLease bool) context.Context {
	ctx = namespaces.WithNamespace(ctx, spi.namespace)
	if withLease {
		ctx = leases.WithLease(ctx, spi.lease.ID)
	}
	return ctx
}

func (spi *ctrdSpi) getContainerRootFSDir(containerID string, rootFS string) string {
	return filepath.Join(spi.metaPath, spi.snapshotterType, containerID, rootFS)
}

func (spi *ctrdSpi) getContainerFSDir(containerID string) string {
	return filepath.Join(spi.metaPath, spi.snapshotterType, containerID)
}

func (spi *ctrdSpi) generateSnapshotID(containerID string) string {
	return fmt.Sprintf(snapshotIDTemplate, containerID)
}

// best effort
func (spi *ctrdSpi) disperseImageResources(initialLeases []leases.Lease) {
	var (
		images []containerd.Image
		err    error
	)

	if images, err = spi.ListImages(context.TODO()); err != nil {
		log.ErrorErr(err, "could not list images")
		return
	}
	var imageToContent = make(map[containerd.Image][]v1.Descriptor)
	for _, i := range images {
		var foundLease bool
		for _, l := range initialLeases {
			if i.Name() == l.ID {
				foundLease = true
				break
			}
		}
		if !foundLease {
			var descriptors []v1.Descriptor
			if descriptors, err = getImageDescriptors(context.TODO(), i.ContentStore(), i.Target()); err != nil {
				log.ErrorErr(err, "could not get content of image = %s", i.Name())
			} else {
				imageToContent[i] = descriptors
			}
		}
	}

	if len(imageToContent) == 0 {
		return
	}

	var contentToRemove, contentToKeep []v1.Descriptor
	for i, descriptors := range imageToContent {
		var imageLease leases.Lease
		if imageLease, err = spi.leaseService.Create(context.TODO(), leases.WithID(i.Name())); err != nil {
			log.ErrorErr(err, "could not create lease with ID = %s", i.Name())
			contentToKeep = append(contentToKeep, descriptors...)
			continue
		}

		for _, desc := range descriptors {
			resource := leases.Resource{
				ID:   desc.Digest.String(),
				Type: "content",
			}
			if err = spi.leaseService.AddResource(context.TODO(), imageLease, resource); err != nil {
				log.ErrorErr(err, "could not add resource with id = %s to lease with id = %s", resource.ID, imageLease.ID)
			} else {
				contentToRemove = append(contentToRemove, desc)
			}
		}
	}

	for _, r := range contentToRemove {
		var keep bool
		for _, k := range contentToKeep {
			if r.Digest.String() == k.Digest.String() {
				keep = true
			}
		}
		if !keep {
			r := leases.Resource{
				ID:   r.Digest.String(),
				Type: "content",
			}
			if err = spi.leaseService.DeleteResource(context.TODO(), *spi.lease, r); err != nil {
				log.ErrorErr(err, "could not remove resource with id = %s from lease with id = %s", r.ID, spi.lease.ID)
			}
		}
	}
}

func getImageDescriptors(ctx context.Context, cs content.Store, desc v1.Descriptor) ([]v1.Descriptor, error) {
	var descriptors []v1.Descriptor

	switch desc.MediaType {
	case images.MediaTypeDockerSchema2ManifestList, v1.MediaTypeImageIndex:
		descriptors = append(descriptors, desc)

		children, err := images.Children(ctx, cs, desc)
		if err != nil {
			if errdefs.IsNotFound(err) {
				return []v1.Descriptor{}, nil
			}
			return []v1.Descriptor{}, err
		}
		for _, child := range children {
			descs, e := getImageDescriptors(ctx, cs, child)
			if e != nil {
				return []v1.Descriptor{}, err
			}
			descriptors = append(descriptors, descs...)
		}
	case images.MediaTypeDockerSchema2Manifest, v1.MediaTypeImageManifest:
		children, err := images.Children(ctx, cs, desc)
		if err != nil {
			if errdefs.IsNotFound(err) {
				return []v1.Descriptor{}, nil
			}
			return []v1.Descriptor{}, err
		}
		descriptors = append(descriptors, children...)
		descriptors = append(descriptors, desc)
		return descriptors, nil
	default:
		return nil, log.NewErrorf("unhandled media type: %s", desc.MediaType)
	}
	return descriptors, nil
}
