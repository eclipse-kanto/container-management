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

package things

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	hostPath      = "/path/dev"
	containerPath = "/path/dev"
	cgroupPerm    = "r"
)

func TestFromAPIDevice(t *testing.T) {
	apiDevice := types.DeviceMapping{
		PathOnHost:        hostPath,
		PathInContainer:   hostPath,
		CgroupPermissions: hostPath,
	}

	thingsDevice := fromAPIDevice(apiDevice)

	t.Run("test_from_api_device_host_path", func(t *testing.T) {
		testutil.AssertEqual(t, apiDevice.PathOnHost, thingsDevice.PathOnHost)
	})

	t.Run("test_from_api_device_container_path", func(t *testing.T) {
		testutil.AssertEqual(t, apiDevice.PathInContainer, thingsDevice.PathInContainer)
	})

	t.Run("test_from_api_device_cgroup_permissions", func(t *testing.T) {
		testutil.AssertEqual(t, apiDevice.CgroupPermissions, thingsDevice.CgroupPermissions)
	})

}

func TestToAPIDevice(t *testing.T) {
	thingsDevice := device{
		PathOnHost:        hostPath,
		PathInContainer:   hostPath,
		CgroupPermissions: hostPath,
	}

	apiDevice := toAPIDevice(&thingsDevice)

	t.Run("test_to_api_device_host_path", func(t *testing.T) {
		testutil.AssertEqual(t, thingsDevice.PathOnHost, apiDevice.PathOnHost)
	})
	t.Run("test_to_api_device_container_path", func(t *testing.T) {
		testutil.AssertEqual(t, thingsDevice.PathInContainer, apiDevice.PathInContainer)
	})
	t.Run("test_to_api_device_cgroup_permissions", func(t *testing.T) {
		testutil.AssertEqual(t, thingsDevice.CgroupPermissions, apiDevice.CgroupPermissions)
	})
}
