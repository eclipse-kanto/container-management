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
	"github.com/eclipse-kanto/container-management/containerm/log"
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

func (spi *ctrdSpi) unleaseImageResources() {
	if rs, err := spi.leaseService.ListResources(context.TODO(), *spi.lease); err == nil {
		for _, r := range rs {
			if r.Type == "content" {
				// delete only dereferences the resources from the lease
				if err = spi.leaseService.DeleteResource(context.TODO(), *spi.lease, r); err != nil {
					log.ErrorErr(err, "could not remove resource with ID = %s of lease with ID = %s", r.ID, spi.lease.ID)
				}
			}
		}
	} else {
		log.ErrorErr(err, "could not list resources of lease with ID = %s", spi.lease.ID)
	}
	return
}
