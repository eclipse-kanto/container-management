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
	"os"
	"path/filepath"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/snapshots"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/opencontainers/image-spec/identity"
	"golang.org/x/sys/unix"
)

func (spi *ctrdSpi) GetSnapshotID(containerID string) string {
	return fmt.Sprintf(snapshotIDTemplate, containerID)
}

func (spi *ctrdSpi) GetSnapshot(ctx context.Context, containerID string) (snapshots.Info, error) {
	ctx = spi.setContext(ctx, true)
	return spi.snapshotService.Stat(ctx, spi.generateSnapshotID(containerID))
}

func (spi *ctrdSpi) PrepareSnapshot(ctx context.Context, containerID string, image containerd.Image) error {
	ctx = spi.setContext(ctx, false)
	originalCtx := ctx
	ctx = leases.WithLease(ctx, spi.lease.ID)

	diffIDs, err := image.RootFS(ctx)
	if err != nil {
		return err
	}
	parent := identity.ChainID(diffIDs).String()
	snapshotID := spi.generateSnapshotID(containerID)

	// NOTE: The image is always unpacked during pulling. But there
	// may be a crash or a termination for some reason leaving the image stored
	// in containerd without unpacking. And the following creating container
	// request will fail on preparing snapshot because there is no such
	// parent snapshotter. Thus, we should skip the not
	// found error and retry unpacking
	_, err = spi.snapshotService.Prepare(ctx, snapshotID, parent)
	if err == nil || !errdefs.IsNotFound(err) {
		return err
	}
	log.Debug("checking unpack status for image %s on %s snapshotter...", image.Name(), spi.snapshotterType)

	// check unpacked
	unpacked, werr := image.IsUnpacked(ctx, spi.snapshotterType)
	if werr != nil {
		log.ErrorErr(werr, "failed to check unpack status for image %s on %s snapshotter", image.Name(), spi.snapshotterType)
		return werr
	}

	// if it is not unpacked - unpack
	if !unpacked {
		log.Warn("the image %s is not unpacked for %s snapshotter - will try to unpack it...", image.Name(), spi.snapshotterType)
		// NOTE!: don't use container-management lease ID here because container-management lease ID
		// will hold the snapshotter forever, which means that the
		// snapshotter will not be removed if we remove the image
		if werr = image.Unpack(originalCtx, spi.snapshotterType); werr != nil {
			log.WarnErr(werr, "failed to unpack image %s on %s snapshotter", image.Name(), spi.snapshotterType)
			return werr
		}

		// retry
		_, err = spi.snapshotService.Prepare(ctx, snapshotID, parent)
		return err
	}
	return nil
}

func (spi *ctrdSpi) MountSnapshot(ctx context.Context, containerID string, rootFS string) error {
	ctx = spi.setContext(ctx, true)
	mounts, err := spi.snapshotService.Mounts(ctx, spi.generateSnapshotID(containerID))
	if err != nil {
		return err
	} else if len(mounts) != 1 {
		return log.NewErrorf("failed to get mounts for snapshot for container %s: not equals 1", containerID)
	}

	mntFS := filepath.Join(spi.metaPath, spi.snapshotterType, containerID, rootFS)

	mkdirErr := os.MkdirAll(mntFS, 0755)
	if mkdirErr != nil && !os.IsExist(mkdirErr) {
		return mkdirErr
	}
	var mountErr error
	defer func() {
		if mountErr != nil {
			if rmErr := os.RemoveAll(spi.getContainerFSDir(containerID)); rmErr != nil {
				log.WarnErr(rmErr, "error cleaning up mount dirs after snapshot mount failure for container %s", containerID)
			}
		}
	}()
	mountErr = mounts[0].Mount(mntFS)
	return mountErr
}

func (spi *ctrdSpi) RemoveSnapshot(ctx context.Context, containerID string) error {
	ctx = spi.setContext(ctx, true)
	if err := spi.snapshotService.Remove(ctx, spi.generateSnapshotID(containerID)); err != nil && !errdefs.IsNotFound(err) {
		return err
	}
	return nil
}
func (spi *ctrdSpi) UnmountSnapshot(ctx context.Context, containerID string, rootFS string) error {
	ctx = spi.setContext(ctx, true)
	mountFS := spi.getContainerRootFSDir(containerID, rootFS)
	if err := mount.Unmount(mountFS, unix.MNT_FORCE); err != nil {
		log.ErrorErr(err, "error unmounting the rootfs for container ID = %s", containerID)
		return err
	}
	if err := os.RemoveAll(spi.getContainerFSDir(containerID)); err != nil {
		log.ErrorErr(err, "error removing the meta directory for container ID = %s", containerID)
		return err
	}
	return nil
}
