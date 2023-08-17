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

package util

import (
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

type errorTest struct {
	inputString string
	errMessage  string
}

func TestParseDeviceMappings(t *testing.T) {
	testCases := map[string]struct {
		inputString    string
		expectedDevice *types.DeviceMapping
	}{
		"test_parse_device_mapping_valid_input_without_cgrops": {
			inputString: "/dev/ttyACM0:/dev/ttyUSB0",
			expectedDevice: &types.DeviceMapping{
				PathOnHost:        "/dev/ttyACM0",
				PathInContainer:   "/dev/ttyUSB0",
				CgroupPermissions: "rwm",
			},
		},
		"test_parse_device_mapping_valid_input_with_readonly": {
			inputString: "/dev/ttyACM1:/dev/ttyUSB1:r",
			expectedDevice: &types.DeviceMapping{
				PathOnHost:        "/dev/ttyACM1",
				PathInContainer:   "/dev/ttyUSB1",
				CgroupPermissions: "r",
			},
		},
		"test_parse_device_mapping_valid_input_with_two_cgroup_permissions": {
			inputString: "/dev/ttyACM2:/dev/ttyUSB2:mw",
			expectedDevice: &types.DeviceMapping{
				PathOnHost:        "/dev/ttyACM2",
				PathInContainer:   "/dev/ttyUSB2",
				CgroupPermissions: "mw",
			},
		},
	}

	index := 0
	inputStrings := make([]string, len(testCases))
	expectedDevices := make([]types.DeviceMapping, len(testCases))

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Log(testCase.inputString)

			res, err := ParseDeviceMapping(testCase.inputString)
			testutil.AssertNil(t, err)
			testutil.AssertEqual(t, testCase.expectedDevice, res)

			res, err = ParseDeviceMapping(DeviceMappingToString(res))
			testutil.AssertNil(t, err)
			testutil.AssertEqual(t, testCase.expectedDevice, res)

			inputStrings[index] = testCase.inputString
			expectedDevices[index] = *res
			index++
		})
	}

	t.Run("test_parse_device_mapping_multiple", func(t *testing.T) {
		res, err := ParseDeviceMappings(inputStrings)
		testutil.AssertNil(t, err)
		testutil.AssertEqual(t, expectedDevices, res)
	})
}

func TestParseDeviceMappingsError(t *testing.T) {
	testCases := map[string]errorTest{
		"test_parse_device_mapping_input_empty": {
			inputString: "",
			errMessage:  "incorrect configuration value for device mapping",
		},
		"test_parse_device_mapping_input_too_long": {
			inputString: "/dev/ttyACM1:/dev/ttyUSB1:r:w:m",
			errMessage:  "incorrect configuration value for device mapping",
		},
		"test_parse_device_mapping_input_too_short": {
			inputString: "/dev/ttyACM1",
			errMessage:  "incorrect configuration value for device mapping",
		},
		"test_parse_device_mapping_no_cgroup_permissions": {
			inputString: "/dev/ttyACM1:/dev/ttyUSB1:",
			errMessage:  "incorrect cgroup permissions format for device mapping",
		},
		"test_parse_device_mapping_cgroup_permissions_too_long": {
			inputString: "/dev/ttyACM1:/dev/ttyUSB1:rwmrwm",
			errMessage:  "incorrect cgroup permissions format for device mapping",
		},
		"test_parse_device_mapping_invalid_cgroup_permission": {
			inputString: "/dev/ttyACM1:/dev/ttyUSB1:R",
			errMessage:  "incorrect cgroup permissions format for device mapping",
		},
	}

	inputStrings := make([]string, 2)
	inputStrings[0] = "/dev/ttyACM0:/dev/ttyUSB0"

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			inputStrings[1] = testCase.inputString

			res, err := ParseDeviceMappings(inputStrings)
			testutil.AssertError(t, log.NewErrorf(testCase.errMessage+" %s", testCase.inputString), err)
			testutil.AssertNil(t, res)
		})
	}
}

func TestParseMountPoints(t *testing.T) {
	testCases := map[string]struct {
		inputString   string
		expectedMount *types.MountPoint
	}{
		"test_parse_mount_point_valid_input_propagation_mode_missing": {
			inputString: "/home/someuser:/home/root",
			expectedMount: &types.MountPoint{
				Source:          "/home/someuser",
				Destination:     "/home/root",
				PropagationMode: types.RPrivatePropagationMode,
			},
		},
		"test_parse_mount_point_valid_input_propagation_mode_private": {
			inputString: "/home/someuser:/home/root:private",
			expectedMount: &types.MountPoint{
				Source:          "/home/someuser",
				Destination:     "/home/root",
				PropagationMode: types.PrivatePropagationMode,
			},
		},
		"test_parse_mount_point_valid_input_propagation_mode_rprivate": {
			inputString: "/var:/var:rprivate",
			expectedMount: &types.MountPoint{
				Source:          "/var",
				Destination:     "/var",
				PropagationMode: types.RPrivatePropagationMode,
			},
		},
		"test_parse_mount_point_valid_input_propagation_mode_shared": {
			inputString: "/etc:/etc:shared",
			expectedMount: &types.MountPoint{
				Source:          "/etc",
				Destination:     "/etc",
				PropagationMode: types.SharedPropagationMode,
			},
		},
		"test_parse_mount_point_valid_input_propagation_mode_rshared": {
			inputString: "/usr/bin:/usr/bin:rshared",
			expectedMount: &types.MountPoint{
				Source:          "/usr/bin",
				Destination:     "/usr/bin",
				PropagationMode: types.RSharedPropagationMode,
			},
		},
		"test_parse_mount_point_valid_input_propagation_mode_slave": {
			inputString: "/data:/data:slave",
			expectedMount: &types.MountPoint{
				Source:          "/data",
				Destination:     "/data",
				PropagationMode: types.SlavePropagationMode,
			},
		},
		"test_parse_mount_point_valid_input_propagation_mode_rslave": {
			inputString: "/tmp:/tmp:rslave",
			expectedMount: &types.MountPoint{
				Source:          "/tmp",
				Destination:     "/tmp",
				PropagationMode: types.RSlavePropagationMode,
			},
		},
	}

	index := 0
	inputStrings := make([]string, len(testCases))
	expectedMounts := make([]types.MountPoint, len(testCases))

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {

			t.Log(testCase.inputString)

			res, err := ParseMountPoint(testCase.inputString)
			testutil.AssertNil(t, err)
			testutil.AssertEqual(t, testCase.expectedMount, res)

			res, err = ParseMountPoint(MountPointToString(res))
			testutil.AssertNil(t, err)
			testutil.AssertEqual(t, testCase.expectedMount, res)

			inputStrings[index] = testCase.inputString
			expectedMounts[index] = *res
			index++
		})
	}

	t.Run("test_parse_mount_points_multiple", func(t *testing.T) {
		res, err := ParseMountPoints(inputStrings)
		testutil.AssertNil(t, err)
		testutil.AssertEqual(t, expectedMounts, res)
	})
}

func TestParseMountPointsError(t *testing.T) {
	testCases := map[string]errorTest{
		"test_parse_mount_point_input_empty": {
			inputString: "",
			errMessage:  "Incorrect number of parameters of the mount point",
		},
		"test_parse_mount_point_input_too_long": {
			inputString: "/data:/data:private:shared",
			errMessage:  "Incorrect number of parameters of the mount point",
		},
		"test_parse_mount_point_input_too_short": {
			inputString: "/home",
			errMessage:  "Incorrect number of parameters of the mount point",
		},
	}

	inputStrings := make([]string, 2)
	inputStrings[0] = "/etc:/etc"

	for testName, testCase := range testCases {
		t.Log(testName)

		inputStrings[1] = testCase.inputString

		res, err := ParseMountPoints(inputStrings)
		testutil.AssertError(t, log.NewErrorf(testCase.errMessage+" %s", testCase.inputString), err)
		testutil.AssertNil(t, res)
	}
}

func TestParsePortMappings(t *testing.T) {
	testCases := map[string]struct {
		inputString  string
		expectedPort *types.PortMapping
	}{
		"test_parse_port_mapping_input_host_and_container_port_only": {
			inputString: "80:80",
			expectedPort: &types.PortMapping{
				ContainerPort: 80,
				HostPort:      80,
			},
		},
		"test_parse_port_mapping_input_host_and_container_port_plus_protocol": {
			inputString: "88:8888/udp",
			expectedPort: &types.PortMapping{
				Proto:         "udp",
				ContainerPort: 8888,
				HostPort:      88,
			},
		},
		"test_parse_port_mapping_input_host_range": {
			inputString: "5000-6000:8080/udp",
			expectedPort: &types.PortMapping{
				Proto:         "udp",
				ContainerPort: 8080,
				HostPort:      5000,
				HostPortEnd:   6000,
			},
		},
		"test_parse_port_mapping_input_host_ip_included": {
			inputString: "192.168.0.1:7000-8000:8081/tcp",
			expectedPort: &types.PortMapping{
				Proto:         "tcp",
				ContainerPort: 8081,
				HostPort:      7000,
				HostPortEnd:   8000,
				HostIP:        "192.168.0.1",
			},
		},
	}

	index := 0
	inputStrings := make([]string, len(testCases))
	expectedPorts := make([]types.PortMapping, len(testCases))

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Log(testCase.inputString)

			res, err := ParsePortMapping(testCase.inputString)
			testutil.AssertNil(t, err)
			testutil.AssertEqual(t, testCase.expectedPort, res)

			res, err = ParsePortMapping(PortMappingToString(res))
			testutil.AssertNil(t, err)
			testutil.AssertEqual(t, testCase.expectedPort, res)

			inputStrings[index] = testCase.inputString
			expectedPorts[index] = *res
			index++
		})
	}

	t.Run("test_parse_port_mappings_multiple", func(t *testing.T) {
		res, err := ParsePortMappings(inputStrings)
		testutil.AssertNil(t, err)
		testutil.AssertEqual(t, expectedPorts, res)
	})
}

func TestParsePortMappingsError(t *testing.T) {
	testCases := map[string]errorTest{
		"test_parse_port_mapping_input_empty": {
			inputString: "",
			errMessage:  "Incorrect port mapping configuration",
		},
		"test_parse_port_mapping_input_too_long": {
			inputString: "192.168.1.100:5000-6000:127.0.0.1:80/tcp",
			errMessage:  "Incorrect port mapping configuration",
		},
		"test_parse_port_mapping_input_too_short": {
			inputString: "8080",
			errMessage:  "Incorrect port mapping configuration",
		},
		"test_parse_port_mapping_invalid_host_ip": {
			inputString: "192.168.1.300:8080:8080",
			errMessage:  "Incorrect host ip port mapping configuration",
		},
		"test_parse_port_mapping_invalid_host_port": {
			inputString: "FF00:8080",
			errMessage:  "Incorrect host port mapping configuration",
		},
		"test_parse_port_mapping_invalid_host_port_range": {
			inputString: "100-FF:8080",
			errMessage:  "Incorrect host range port mapping configuration",
		},
		"test_parse_port_mapping_invalid_container_port": {
			inputString: "5000-6000:BABE",
			errMessage:  "Incorrect container port mapping configuration",
		},
	}

	inputStrings := make([]string, 2)
	inputStrings[0] = "80:80"

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			inputStrings[1] = testCase.inputString

			res, err := ParsePortMappings(inputStrings)
			testutil.AssertError(t, log.NewErrorf(testCase.errMessage+" %s", testCase.inputString), err)
			testutil.AssertNil(t, res)
		})
	}
}
