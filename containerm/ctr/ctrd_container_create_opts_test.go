// Copyright (c) 2023 Contributors to the Eclipse Foundation
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
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/images"
)

func TestWithRuntimeOpts(t *testing.T) {
	tests := map[string]struct {
		container *types.Container
	}{
		"test_runtime_type_v1": {
			&types.Container{
				ID:   testCtrID1,
				Name: testContainerName,
				HostConfig: &types.HostConfig{
					Runtime: types.RuntimeTypeV1,
				},
			},
		},
		"test_runtime_type_v2": {
			&types.Container{
				ID:   testCtrID1,
				Name: testContainerName,
				HostConfig: &types.HostConfig{
					Runtime: types.RuntimeTypeV2runcV1,
				},
			},
		},
		"testing_default": {
			&types.Container{
				ID:   testCtrID1,
				Name: testContainerName,
				HostConfig: &types.HostConfig{
					Runtime: "",
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testutil.AssertNotNil(t, WithRuntimeOpts(test.container, ""))
		})
	}
}

func TestWithSpecOpts(t *testing.T) {
	tests := map[string]struct {
		container *types.Container
	}{
		"test_config": {
			&types.Container{
				Config: &types.ContainerConfiguration{
					Cmd: []string{"test"},
					Env: []string{"test"},
				},
				HostConfig: &types.HostConfig{},
			},
		},
		"test_privileged": {
			&types.Container{
				HostConfig: &types.HostConfig{
					Privileged: true,
				},
			},
		},
		"test_extra_capabilities": {
			&types.Container{
				HostConfig: &types.HostConfig{
					ExtraCapabilities: []string{"CAP_NET_ADMIN"},
				},
			},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			testutil.AssertNotNil(t, WithSpecOpts(test.container, containerd.NewImage(&containerd.Client{}, images.Image{}), "/tmp/test"))
		})
	}
}
