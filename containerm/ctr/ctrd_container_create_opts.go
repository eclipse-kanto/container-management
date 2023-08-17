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

package ctr

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/containers"
	ctrdoci "github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/runtime/linux/runctypes"
	runcoptions "github.com/containerd/containerd/runtime/v2/runc/options"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

const rootFSPathDefault = "rootfs"

var (
	// CtrdRuntimes contains all runtime type names
	CtrdRuntimes = map[types.Runtime]string{
		types.RuntimeTypeV1:        "runc",
		types.RuntimeTypeV2runscV1: "runhcs",
		types.RuntimeTypeV2kataV2:  "kata",
		types.RuntimeTypeV2runcV1:  "runc",
		types.RuntimeTypeV2runcV2:  "runc",
	}
)

// WithRuntimeOpts sets the runtime configuration for the container to be created.
func WithRuntimeOpts(container *types.Container, runtimeRootPath string) containerd.NewContainerOpts {
	var (
		options interface{}
	)

	runtimePath := CtrdRuntimes[container.HostConfig.Runtime]
	runtimeRootPathFinal := filepath.Join(runtimePath, fmt.Sprintf("runtimes-%s", container.HostConfig.Runtime))

	useSystemd := util.IsRunningSystemd()
	switch container.HostConfig.Runtime {
	case types.RuntimeTypeV1, types.RuntimeTypeV2runscV1, types.RuntimeTypeV2kataV2:
		options = &runctypes.RuncOptions{
			Runtime:       runtimePath,
			RuntimeRoot:   runtimeRootPathFinal,
			SystemdCgroup: useSystemd,
		}
	case types.RuntimeTypeV2runcV1, types.RuntimeTypeV2runcV2:
		options = &runcoptions.Options{
			BinaryName:    runtimePath,
			Root:          runtimeRootPathFinal,
			SystemdCgroup: useSystemd,
		}
	default:
		return func(_ context.Context, client *containerd.Client, _ *containers.Container) error {
			// do nothing
			return nil
		}

	}

	log.Info("will create options for runtime with name = %s, for container ID = ", container.HostConfig.Runtime, container.ID)
	return containerd.WithRuntime(string(container.HostConfig.Runtime), options)
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
	var args, env []string
	if container.Config != nil {
		args = container.Config.Cmd
		env = container.Config.Env
	}

	specOpts := []ctrdoci.SpecOpts{
		ctrdoci.WithImageConfigArgs(image, args),
		WithCommonOptions(container),
		ctrdoci.WithEnv(env),
		WithDevices(container),
		WithMounts(container),
		WithNamespaces(container),
		WithHooks(container, execRoot),
		WithResources(container),
		WithCgroupsPath(container),
		ctrdoci.WithRootFSPath(rootFSPathDefault),
	}

	if container.HostConfig.Privileged {
		specOpts = append(specOpts, ctrdoci.WithPrivileged)
	}
	if len(container.HostConfig.ExtraCapabilities) > 0 {
		specOpts = append(specOpts, ctrdoci.WithAddedCapabilities(container.HostConfig.ExtraCapabilities))
	}

	return containerd.WithNewSpec(specOpts...)
}
