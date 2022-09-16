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
	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/snapshots"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

// containerClientWrapper is an interface that abstracts the functional scope of the *containerd.Client instance
// that is used by the SPI implementation
// The interface definition is a direct extraction of the *containerd.Client struct function signatures of only such that are used by the SPI.
// The API is based on the currently supported version of containerd client API - 1.5.13 (see go.mod)
type containerClientWrapper interface {
	// NewContainer creates a new container instance
	NewContainer(ctx context.Context, id string, opts ...containerd.NewContainerOpts) (containerd.Container, error)
	// LoadContainer loads a new container instance
	LoadContainer(ctx context.Context, id string) (containerd.Container, error)
	// GetImage retrieves an image from the local cache
	GetImage(ctx context.Context, ref string) (containerd.Image, error)
	// ListImages returns all locally existing images
	ListImages(ctx context.Context, filters ...string) ([]containerd.Image, error)
	// SnapshotService returns the current snapshots manager service
	SnapshotService(snapshotterName string) snapshots.Snapshotter
	// LeasesService returns the current leases manager instance
	LeasesService() leases.Manager
	// ImageService returns the current image store instance
	ImageService() images.Store
	// Pull downloads the provided content and returns an image object
	Pull(ctx context.Context, ref string, opts ...containerd.RemoteOpt) (_ containerd.Image, retErr error)
	// Close closes the internal communication channel
	Close() error
	// Subscribe subscribes for containerd events
	Subscribe(ctx context.Context, filters ...string) (ch <-chan *events.Envelope, errs <-chan error)
}

// containerdSpi is a wrapper interface for providing a context-ready and scoped images, containers and snapshots related functionalities handling
type containerdSpi interface {
	// Wrapper section for managing the OCI images
	// GetImage returns a locally existing image
	GetImage(ctx context.Context, imageRef string) (containerd.Image, error)
	// PullImage downloads the provided content and returns an image object
	PullImage(ctx context.Context, imageRef string, opts ...containerd.RemoteOpt) (containerd.Image, error)
	// UnpackImage unpacks the contents of the provided image locally
	UnpackImage(ctx context.Context, image containerd.Image, opts ...containerd.UnpackOpt) error
	// DeleteImage removes the contents of the provided image from the disk
	DeleteImage(ctx context.Context, imageRef string) error
	// ListImages returns all locally existing images
	ListImages(ctx context.Context) ([]containerd.Image, error)

	// Wrapper section for managing the file system of the container and its snapshots
	// GetSnapshotID generates a new ID for the snapshot to be used for this container
	GetSnapshotID(containerID string) string
	// GetSnapshot returns a snapshot for this container ID
	GetSnapshot(ctx context.Context, containerID string) (snapshots.Info, error)
	// ListSnapshots collects all snapshots matching the provided filters or all if no filters are provided
	ListSnapshots(ctx context.Context, filters ...string) ([]snapshots.Info, error)
	// PrepareSnapshot initializes a new snapshot for the provided container image for the provided container ID
	PrepareSnapshot(ctx context.Context, containerID string, image containerd.Image, opts ...containerd.UnpackOpt) error
	// MountSnapshot mounts the provided rootFS to an already existing snapshot for the provided container ID
	MountSnapshot(ctx context.Context, containerID string, rootFS string) error
	// RemoveSnapshot removes the snapshot and allocated resources for the provided container ID
	RemoveSnapshot(ctx context.Context, containerID string) error
	// UnmountSnapshot unmounts the snapshot and allocated resources for the provided container ID and rootFS
	UnmountSnapshot(ctx context.Context, containerID string, rootFS string) error

	// Wrapper section for managing the container instances and relevant processes allocated
	// LoadContainer loads an existing container instance
	LoadContainer(ctx context.Context, containerID string) (containerd.Container, error)
	// CreateContainer creates a new container instance
	CreateContainer(ctx context.Context, containerID string, opts ...containerd.NewContainerOpts) (containerd.Container, error)
	// CreateTask creates a process based on the container's metadata and starts it
	CreateTask(ctx context.Context, container containerd.Container, cioCreatorFunc cio.Creator, opts ...containerd.NewTaskOpts) (containerd.Task, error)
	// LoadTask returns a running task with reattaching the existing streams
	LoadTask(ctx context.Context, container containerd.Container, reattachFunc cio.Attach) (containerd.Task, error)

	// Dispose releases all resources for the instance
	Dispose(ctx context.Context) error

	// Subscribe subscribes for containerd events
	Subscribe(ctx context.Context, filters ...string) (ch <-chan *events.Envelope, errs <-chan error)
}

type ctrdSpi struct {
	client          containerClientWrapper
	lease           *leases.Lease
	namespace       string
	snapshotterType string
	metaPath        string
	snapshotService snapshots.Snapshotter
	imageService    images.Store
}

const containerdGCExpireLabel = "containerd.io/gc.expire"

func newContainerdSpi(rpcAddress, namespace, snapshotterType, metaPath, leaseID string) (containerdSpi, error) {
	ctrdClient, err := containerd.New(rpcAddress, containerd.WithDefaultNamespace(namespace))
	if err != nil {
		return nil, err
	}

	var lease leases.Lease

	leaseSrv := ctrdClient.LeasesService()
	leaseList, err := leaseSrv.List(context.TODO())
	if err != nil {
		return nil, err
	}
	log.Debug("got all leases")

	for _, l := range leaseList {
		log.Debug("checking lease with ID = %s", l.ID)
		if l.ID != leaseID {
			continue
		}
		log.Debug("found lease with ID = %s", leaseID)
		foundExpireLabel := false
		for k := range l.Labels {
			if k == containerdGCExpireLabel {
				foundExpireLabel = true
				break
			}
		}
		log.Debug("is expired lease %s - %v", leaseID, foundExpireLabel)
		// found a lease that matched the condition, just return
		if !foundExpireLabel {
			// remove images content from the lease
			var resources []leases.Resource
			if resources, err = leaseSrv.ListResources(context.TODO(), l); err == nil {
				for _, r := range resources {
					if r.Type == "content" {
						// delete only dereferences the resources from the lease
						if err = leaseSrv.DeleteResource(context.TODO(), l, r); err != nil {
							log.ErrorErr(err, "could not remove resource with ID = %s of lease with ID = %s", r.ID, l.ID)
						}
					}
				}
			} else {
				log.ErrorErr(err, "could not list resources of lease with ID = %s", l.ID)
			}
			log.Debug("will set lease to %v with ID - %s", &l, (&l).ID)
			return &ctrdSpi{
				client:          ctrdClient,
				lease:           &l,
				namespace:       namespace,
				snapshotterType: snapshotterType,
				metaPath:        metaPath,
				snapshotService: ctrdClient.SnapshotService(snapshotterType),
			}, nil
		}
		log.Debug("deleting expired lease %s", leaseID)
		// found a lease with id is container-management.lease and has expire time,
		// then just delete it and wait to recreate a new lease.
		if err := leaseSrv.Delete(context.TODO(), l); err != nil {
			return nil, err
		}

	}
	log.Debug("creating new lease with id = %s ", leaseID)
	// not found a matched lease so it must be created
	if lease, err = leaseSrv.Create(context.TODO(), leases.WithID(leaseID)); err != nil {
		return nil, err
	}
	log.Debug("will set lease to %v with ID - %s", &lease, (&lease).ID)
	return &ctrdSpi{
		client:          ctrdClient,
		lease:           &lease,
		namespace:       namespace,
		snapshotterType: snapshotterType,
		metaPath:        metaPath,
		snapshotService: ctrdClient.SnapshotService(snapshotterType),
	}, nil
}

func (spi *ctrdSpi) Dispose(ctx context.Context) error {
	return spi.client.Close()
}
