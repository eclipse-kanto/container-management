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
	"fmt"
	"log"
	"path/filepath"
	"runtime"

	"github.com/containerd/containerd"
	ctrdoci "github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/runtime/linux/runctypes"
	runcoptions "github.com/containerd/containerd/runtime/v2/runc/options"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
)

const (
	// RuntimeTypeV2runscV1 is the runtime type name for gVisor containerd shim implement the shim v2 api.
	RuntimeTypeV2runscV1 = "io.containerd.runsc.v1"
	// RuntimeTypeV2kataV2 is the runtime type name for kata-runtime containerd shim implement the shim v2 api.
	RuntimeTypeV2kataV2 = "io.containerd.kata.v2"
	// RuntimeTypeV2runcV1 is the runtime type name for runc containerd shim implement the shim v2 api.
	RuntimeTypeV2runcV1 = "io.containerd.runc.v1"

	rootFSPathDefault = "rootfs"
)

var (
	// RuntimeTypeV1 is the runtime type name for containerd shim interface v1 version.
	RuntimeTypeV1 = fmt.Sprintf("io.containerd.runtime.v1.%s", runtime.GOOS)
	// CtrdRuntimes contains all runtime type names
	CtrdRuntimes = map[string]string{
		RuntimeTypeV1:        "runc",
		RuntimeTypeV2runscV1: "runhcs",
		RuntimeTypeV2kataV2:  "kata",
		RuntimeTypeV2runcV1:  "runc",
	}
)

// WithRuntimeOpts sets the runtime configuration for the container to be created.
func WithRuntimeOpts(container *types.Container, runtimeRootPath string) containerd.NewContainerOpts {
	var (
		options interface{}
	)

	if container.HostConfig.Runtime == "" {
		container.HostConfig.Runtime = RuntimeTypeV1
	}

	runtimePath := CtrdRuntimes[container.HostConfig.Runtime]
	runtimeRootPathFinal := filepath.Join(runtimePath, fmt.Sprintf("runtimes-%s", container.HostConfig.Runtime))

	switch container.HostConfig.Runtime {
	case RuntimeTypeV1, RuntimeTypeV2runscV1, RuntimeTypeV2kataV2:
		options = &runctypes.RuncOptions{
			Runtime:     runtimePath,
			RuntimeRoot: runtimeRootPathFinal,
		}
	case RuntimeTypeV2runcV1:
		options = &runcoptions.Options{
			BinaryName: runtimePath,
			Root:       runtimeRootPathFinal,
		}
	default:
		return nil

	}

	log.Printf("Will create options for runtime with name >> %s ", container.HostConfig.Runtime)
	return containerd.WithRuntime(container.HostConfig.Runtime, options)
}

// WithSnapshotOpts sets the snapshotting configuration for the container to be created.
func WithSnapshotOpts(snapshotID string, snapshotterType string) []containerd.NewContainerOpts {
	return []containerd.NewContainerOpts{
		containerd.WithSnapshotter(snapshotterType), // NB! It's very important to set the snapshotter first in the opts - the snapshot ID processing depends on it
		containerd.WithSnapshot(snapshotID),
	}
}

// WithSpecOpts sets the OCI specification configuration options for the container to be created.
func WithSpecOpts(container *types.Container, image containerd.Image, execRoot string) containerd.NewContainerOpts {
	return containerd.WithNewSpec(
		ctrdoci.WithImageConfig(image),
		WithCommonOptions(container),
		WithProcessOptions(container),
		WithDevices(container),
		WithMounts(container),
		WithNamespaces(container),
		WithHooks(container, execRoot),
		WithResources(container),
		ctrdoci.WithRootFSPath(rootFSPathDefault),
	)
}
