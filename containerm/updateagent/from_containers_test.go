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

package updateagent

import (
	"strconv"
	"testing"
	"time"

	ctrtypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/util"

	"github.com/eclipse-kanto/update-manager/api/types"
)

type testExpectedParams struct {
	nonVerboseParams []*types.KeyValuePair
	verboseParams    []*types.KeyValuePair
}

var verboseNonPrivilegedKV = &types.KeyValuePair{
	Key:   keyPrivileged,
	Value: "false",
}

var verboseNonPrivilegedKVs = []*types.KeyValuePair{
	verboseNonPrivilegedKV,
}

func TestFromContainers(t *testing.T) {
	container := createTestContainer("test-container-0")
	util.SetContainerStatusCreated(container)
	container.StartedSuccessfullyBefore = true
	util.SetContainerStatusRunning(container, 3421)

	testContainers := []*ctrtypes.Container{container}
	swNodes := fromContainers(testContainers, true)
	testutil.AssertEqual(t, len(testContainers), len(swNodes))
	for i, node := range swNodes {
		testutil.AssertEqual(t, testContainers[i].Name, node.ID)
		testutil.AssertEqual(t, "v1.2.3", node.Version)
		testutil.AssertEqual(t, types.SoftwareTypeContainer, node.Type)
		// TODO make assert parameters better
		assertParameter(t, node.Parameters, keyDomainName, testContainers[i].DomainName)
		assertParameter(t, node.Parameters, keyHostName, testContainers[i].HostName)
		assertParameter(t, node.Parameters, keyRestartCount, strconv.Itoa(testContainers[i].RestartCount))
		assertParameter(t, node.Parameters, keyCreated, testContainers[i].Created)
		assertParameter(t, node.Parameters, keyManuallyStopped, strconv.FormatBool(testContainers[i].ManuallyStopped))
		assertParameter(t, node.Parameters, keyStartedSuccessfullyBefore, strconv.FormatBool(testContainers[i].StartedSuccessfullyBefore))
	}
}

func TestHostConfigParameters(t *testing.T) {
	testCases := map[string]struct {
		hostConfig     ctrtypes.HostConfig
		expectedParams testExpectedParams
	}{
		"test_host_config_params_privileged": {
			hostConfig: ctrtypes.HostConfig{Privileged: true},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyPrivileged, Value: "true"},
				},
			},
		},
		"test_host_config_params_non_privileged": {
			hostConfig: ctrtypes.HostConfig{Privileged: false},
			expectedParams: testExpectedParams{
				verboseParams: verboseNonPrivilegedKVs,
			},
		},

		"test_host_config_params_restart_policy_no": {
			hostConfig: ctrtypes.HostConfig{RestartPolicy: &ctrtypes.RestartPolicy{Type: ctrtypes.No}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyRestartPolicy, Value: "no"},
				},
				verboseParams: []*types.KeyValuePair{
					verboseNonPrivilegedKV,
					{Key: keyRestartMaxRetries, Value: "0"},
					{Key: keyRestartTimeout, Value: "0s"},
				},
			},
		},
		"test_host_config_params_restart_policy_always": {
			hostConfig: ctrtypes.HostConfig{RestartPolicy: &ctrtypes.RestartPolicy{Type: ctrtypes.Always}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyRestartPolicy, Value: "always"},
				},
				verboseParams: []*types.KeyValuePair{
					verboseNonPrivilegedKV,
					{Key: keyRestartMaxRetries, Value: "0"},
					{Key: keyRestartTimeout, Value: "0s"},
				},
			},
		},
		"test_host_config_params_restart_policy_on_failure": {
			hostConfig: ctrtypes.HostConfig{RestartPolicy: &ctrtypes.RestartPolicy{Type: ctrtypes.OnFailure, MaximumRetryCount: 5, RetryTimeout: 3 * time.Minute}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyRestartPolicy, Value: "on-failure"},
					{Key: keyRestartMaxRetries, Value: "5"},
					{Key: keyRestartTimeout, Value: "3m0s"},
				},
				verboseParams: verboseNonPrivilegedKVs,
			},
		},
		"test_host_config_params_restart_policy_unless_stopped": {
			hostConfig: ctrtypes.HostConfig{RestartPolicy: &ctrtypes.RestartPolicy{Type: ctrtypes.UnlessStopped}},
			expectedParams: testExpectedParams{
				verboseParams: []*types.KeyValuePair{
					verboseNonPrivilegedKV,
					{Key: keyRestartPolicy, Value: "unless-stopped"},
					{Key: keyRestartMaxRetries, Value: "0"},
					{Key: keyRestartTimeout, Value: "0s"},
				},
			},
		},

		"test_host_config_params_network_mode_bridge": {
			hostConfig: ctrtypes.HostConfig{NetworkMode: ctrtypes.NetworkModeBridge},
			expectedParams: testExpectedParams{
				verboseParams: []*types.KeyValuePair{
					verboseNonPrivilegedKV,
					{Key: keyNetwork, Value: "bridge"},
				},
			},
		},
		"test_host_config_params_network_mode_host": {
			hostConfig: ctrtypes.HostConfig{NetworkMode: ctrtypes.NetworkModeHost},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyNetwork, Value: "host"},
				},
				verboseParams: verboseNonPrivilegedKVs,
			},
		},

		"test_host_config_params_log_driver_none": {
			hostConfig: ctrtypes.HostConfig{LogConfig: &ctrtypes.LogConfiguration{
				DriverConfig: &ctrtypes.LogDriverConfiguration{Type: ctrtypes.LogConfigDriverNone},
			}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyLogDriver, Value: "none"},
				},
				verboseParams: verboseNonPrivilegedKVs,
			},
		},
		"test_host_config_params_log_driver_json_with_custom_max_files_and_size": {
			hostConfig: ctrtypes.HostConfig{LogConfig: &ctrtypes.LogConfiguration{
				DriverConfig: &ctrtypes.LogDriverConfiguration{Type: ctrtypes.LogConfigDriverJSONFile, MaxFiles: 6, MaxSize: "100K"},
			}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyLogMaxFiles, Value: "6"},
					{Key: keyLogMaxSize, Value: "100K"},
				},
				verboseParams: []*types.KeyValuePair{
					verboseNonPrivilegedKV,
					{Key: keyLogDriver, Value: "json-file"},
				},
			},
		},
		"test_host_config_params_log_driver_json_with_custom_log_path": {
			hostConfig: ctrtypes.HostConfig{LogConfig: &ctrtypes.LogConfiguration{
				DriverConfig: &ctrtypes.LogDriverConfiguration{Type: ctrtypes.LogConfigDriverJSONFile, MaxFiles: 2, MaxSize: "100M", RootDir: "/tmp/logs"},
			}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyLogPath, Value: "/tmp/logs"},
				},
				verboseParams: []*types.KeyValuePair{
					verboseNonPrivilegedKV,
					{Key: keyLogDriver, Value: "json-file"},
					{Key: keyLogMaxFiles, Value: "2"},
					{Key: keyLogMaxSize, Value: "100M"},
				},
			},
		},

		"test_host_config_params_log_mode_blocking": {
			hostConfig: ctrtypes.HostConfig{LogConfig: &ctrtypes.LogConfiguration{
				ModeConfig: &ctrtypes.LogModeConfiguration{Mode: ctrtypes.LogModeBlocking},
			}},
			expectedParams: testExpectedParams{
				verboseParams: []*types.KeyValuePair{
					verboseNonPrivilegedKV,
					{Key: keyLogMode, Value: "blocking"},
				},
			},
		},
		"test_host_config_params_log_mode_non_blocking_with_custom_max_buffer_size": {
			hostConfig: ctrtypes.HostConfig{LogConfig: &ctrtypes.LogConfiguration{
				ModeConfig: &ctrtypes.LogModeConfiguration{Mode: ctrtypes.LogModeNonBlocking, MaxBufferSize: "100K"},
			}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyLogMode, Value: "non-blocking"},
					{Key: keyLogMaxBufferSize, Value: "100K"},
				},
				verboseParams: verboseNonPrivilegedKVs,
			},
		},
		"test_host_config_params_log_mode_non_blocking_with_default_max_buffer_size": {
			hostConfig: ctrtypes.HostConfig{LogConfig: &ctrtypes.LogConfiguration{
				ModeConfig: &ctrtypes.LogModeConfiguration{Mode: ctrtypes.LogModeNonBlocking, MaxBufferSize: "1M"},
			}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyLogMode, Value: "non-blocking"},
				},
				verboseParams: []*types.KeyValuePair{
					verboseNonPrivilegedKV,
					{Key: keyLogMaxBufferSize, Value: "1M"},
				},
			},
		},

		"test_host_config_params_resources_with_memory_reservation_swap": {
			hostConfig: ctrtypes.HostConfig{Resources: &ctrtypes.Resources{Memory: "200m", MemoryReservation: "100m", MemorySwap: "50m"}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyMemory, Value: "200m"},
					{Key: keyMemoryReservation, Value: "100m"},
					{Key: keyMemorySwap, Value: "50m"},
				},
				verboseParams: verboseNonPrivilegedKVs,
			},
		},
		"test_host_config_params_resources_with_memory_and_reservation_only": {
			hostConfig: ctrtypes.HostConfig{Resources: &ctrtypes.Resources{Memory: "300m", MemoryReservation: "300m"}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyMemory, Value: "300m"},
					{Key: keyMemoryReservation, Value: "300m"},
				},
				verboseParams: verboseNonPrivilegedKVs,
			},
		},
		"test_host_config_params_resources_with_memory_and_swap_only": {
			hostConfig: ctrtypes.HostConfig{Resources: &ctrtypes.Resources{Memory: "1g", MemorySwap: "150m"}},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyMemory, Value: "1g"},
					{Key: keyMemorySwap, Value: "150m"},
				},
				verboseParams: verboseNonPrivilegedKVs,
			},
		},
	}
	for _, verbose := range []bool{true, false} {
		for testName, testCase := range testCases {
			t.Run(fullTestName(testName, verbose), func(t *testing.T) {
				assertParameters(t, testCase.expectedParams, hostConfigParameters(&testCase.hostConfig, verbose), verbose)
			})
		}
	}
}

func TestHostConfigParametersDevices(t *testing.T) {
	hostConfig := &ctrtypes.HostConfig{}
	testDevices := []string{
		"/dev/ttyACM0:/dev/ttyUSB0:rwm",
		"/dev/ttyACM1:/dev/ttyUSB1:r",
		"/dev/ttyACM2:/dev/ttyUSB2:mw",
	}
	testDeviceMappings, err := util.ParseDeviceMappings(testDevices)
	testutil.AssertNil(t, err)
	hostConfig.Devices = testDeviceMappings
	params := hostConfigParameters(hostConfig, false)
	for _, testDevice := range testDevices {
		assertMultipleParameter(t, params, keyDevice, testDevice)
	}
	testutil.AssertEqual(t, len(testDeviceMappings), len(params))
}

func TestHostConfigParametersPortMappings(t *testing.T) {
	hostConfig := &ctrtypes.HostConfig{}
	testPorts := []string{
		"80:80",
		"88:8888/udp",
		"5000-6000:8080/udp",
		"192.168.0.1:7000-8000:8081/tcp",
	}
	testPortMappings, err := util.ParsePortMappings(testPorts)
	testutil.AssertNil(t, err)
	hostConfig.PortMappings = testPortMappings
	params := hostConfigParameters(hostConfig, false)
	for _, testPort := range testPorts {
		assertMultipleParameter(t, params, keyPort, testPort)
	}
	testutil.AssertEqual(t, len(testPortMappings), len(params))
}

func TestHostConfigParametersExtraHosts(t *testing.T) {
	hostConfig := &ctrtypes.HostConfig{
		ExtraHosts: []string{"ctrhost:host_ip", "somehost:192.168.0.1"},
	}
	params := hostConfigParameters(hostConfig, false)
	for _, host := range hostConfig.ExtraHosts {
		assertMultipleParameter(t, params, keyHost, host)
	}
	testutil.AssertEqual(t, len(hostConfig.ExtraHosts), len(params))
}

func TestMountPointParameters(t *testing.T) {
	testMounts := []string{
		"/home/someuser:/home/root:private", "/var:/var:rprivate",
		"/etc:/etc:shared", "/usr/bin:/usr/bin:rshared",
		"/data:/data:slave", "/tmp:/tmp:rslave",
	}
	testMountPoints, err := util.ParseMountPoints(testMounts)
	testutil.AssertNil(t, err)
	params := mountPointParameters(testMountPoints)
	for _, testMount := range testMounts {
		assertMultipleParameter(t, params, keyMount, testMount)
	}
	testutil.AssertEqual(t, len(testMounts), len(params))
}

func TestContainerConfigParameters(t *testing.T) {
	testCases := map[string]struct {
		args           []string
		envs           []string
		expectedParams []*types.KeyValuePair
	}{
		"test_container_config_parameters_nil_args_and_nil_envs": {},
		"test_container_config_parameters_nil_args_and_empty_envs": {
			envs: []string{},
		},
		"test_container_config_parameters_nil_args_and_some_envs": {
			args: []string{},
			envs: []string{"ENV_X=VAL_A", "ENV_Y=VAL_B", "ENV_C="},
			expectedParams: []*types.KeyValuePair{
				{Key: keyEnv, Value: "ENV_X=VAL_A"},
				{Key: keyEnv, Value: "ENV_Y=VAL_B"},
				{Key: keyEnv, Value: "ENV_C="},
			},
		},
		"test_container_config_parameters_empty_args_and_nil_envs": {
			args: []string{},
		},
		"test_container_config_parameters_empty_args_and_empty_envs": {
			args: []string{},
			envs: []string{},
		},
		"test_container_config_parameters_empty_args_and_some_envs": {
			args: []string{},
			envs: []string{"ENV_X=VAL_A", "ENV_Y=VAL_B", "ENV_C="},
			expectedParams: []*types.KeyValuePair{
				{Key: keyEnv, Value: "ENV_X=VAL_A"},
				{Key: keyEnv, Value: "ENV_Y=VAL_B"},
				{Key: keyEnv, Value: "ENV_C="},
			},
		},
		"test_container_config_parameters_some_args_and_nil_envs": {
			args: []string{"arg1", "arg2", "arg3"},
			expectedParams: []*types.KeyValuePair{
				{Key: keyCmd, Value: "arg1"},
				{Key: keyCmd, Value: "arg2"},
				{Key: keyCmd, Value: "arg3"},
			},
		},
		"test_container_config_parameters_some_args_and_empty_envs": {
			args: []string{"arg1", "arg2", "arg3"},
			envs: []string{},
			expectedParams: []*types.KeyValuePair{
				{Key: keyCmd, Value: "arg1"},
				{Key: keyCmd, Value: "arg2"},
				{Key: keyCmd, Value: "arg3"},
			},
		},
		"test_container_config_parameters_some_args_and_some_envs": {
			args: []string{"arg1", "arg2", "arg3"},
			envs: []string{"ENV_X=VAL_A", "ENV_Y=VAL_B", "ENV_C="},
			expectedParams: []*types.KeyValuePair{
				{Key: keyCmd, Value: "arg1"},
				{Key: keyCmd, Value: "arg2"},
				{Key: keyCmd, Value: "arg3"},
				{Key: keyEnv, Value: "ENV_X=VAL_A"},
				{Key: keyEnv, Value: "ENV_Y=VAL_B"},
				{Key: keyEnv, Value: "ENV_C="},
			},
		},
	}
	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			params := containerConfigParameters(&ctrtypes.ContainerConfiguration{Env: testCase.envs, Cmd: testCase.args})
			for _, kv := range testCase.expectedParams {
				assertMultipleParameter(t, params, kv.Key, kv.Value)
			}
			testutil.AssertEqual(t, len(testCase.expectedParams), len(params))
		})
	}
}

func TestStateParameters(t *testing.T) {
	commonVerboseExpectedParams := []*types.KeyValuePair{
		{Key: keyFinishedAt, Value: ""},
		{Key: keyExitCode, Value: "0"},
	}
	testCases := map[string]struct {
		testSetup      func(*ctrtypes.Container)
		expectedParams testExpectedParams
	}{
		"test_state_params_container_creating": {
			testSetup: func(c *ctrtypes.Container) { c.State.Status = ctrtypes.Creating },
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyStatus, Value: "Creating"},
				},
				verboseParams: commonVerboseExpectedParams,
			},
		},
		"test_state_params_container_created": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusCreated(c) },
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyStatus, Value: "Created"},
				},
				verboseParams: commonVerboseExpectedParams,
			},
		},
		"test_state_params_container_running": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusRunning(c, 1234) },
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyStatus, Value: "Running"},
				},
				verboseParams: commonVerboseExpectedParams,
			},
		},
		"test_state_params_container_stopped_normally": {
			testSetup: func(c *ctrtypes.Container) {
				util.SetContainerStatusStopped(c, 0, "")
				c.State.FinishedAt = "2023-01-01T15:04:05.999999999Z07:00"
			},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyStatus, Value: "Stopped"},
					{Key: keyFinishedAt, Value: "2023-01-01T15:04:05.999999999Z07:00"},
				},
				verboseParams: []*types.KeyValuePair{
					{Key: keyExitCode, Value: "0"},
				},
			},
		},
		"test_state_params_container_stopped_error": {
			testSetup: func(c *ctrtypes.Container) {
				util.SetContainerStatusStopped(c, -1, "stopped with error")
				c.State.FinishedAt = "2023-01-11T15:04:05.999999999Z07:00"
			},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyStatus, Value: "Stopped"},
					{Key: keyFinishedAt, Value: "2023-01-11T15:04:05.999999999Z07:00"},
					{Key: keyExitCode, Value: "-1"},
				},
			},
		},
		"test_state_params_container_paused": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusPaused(c) },
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyStatus, Value: "Paused"},
				},
				verboseParams: commonVerboseExpectedParams,
			},
		},
		"test_state_params_container_exited": {
			testSetup: func(c *ctrtypes.Container) {
				util.SetContainerStatusExited(c, 1234, "unexpected exit", false)
				c.State.FinishedAt = "2023-01-13T15:04:05.999999999Z07:00"
			},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyStatus, Value: "Exited"},
					{Key: keyFinishedAt, Value: "2023-01-13T15:04:05.999999999Z07:00"},
					{Key: keyExitCode, Value: "1234"},
				},
			},
		},
		"test_state_params_container_dead": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusDead(c) },
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyStatus, Value: "Dead"},
				},
				verboseParams: commonVerboseExpectedParams,
			},
		},
		"test_state_params_container_unknown": {
			testSetup: func(c *ctrtypes.Container) {
				c.State.Status = ctrtypes.Unknown
				c.State.FinishedAt = "2023-01-02T15:04:05.999999999Z07:00"
				c.State.ExitCode = 999
			},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyStatus, Value: "Unknown"},
					{Key: keyFinishedAt, Value: "2023-01-02T15:04:05.999999999Z07:00"},
					{Key: keyExitCode, Value: "999"},
				},
			},
		},
	}
	for _, verbose := range []bool{true, false} {
		for testName, testCase := range testCases {
			t.Run(fullTestName(testName, verbose), func(t *testing.T) {
				testContainer := createTestContainer("test-container")
				testContainer.State = &ctrtypes.State{}
				testCase.testSetup(testContainer)
				assertParameters(t, testCase.expectedParams, stateParameters(testContainer.State, verbose), verbose)
			})
		}
	}
}

func TestIOConfigParameters(t *testing.T) {
	testCases := map[string]struct {
		ioConfig       *ctrtypes.IOConfig
		expectedParams testExpectedParams
	}{
		"test_io_config_params_no_tty_no_openstdin": {
			ioConfig: &ctrtypes.IOConfig{},
			expectedParams: testExpectedParams{
				verboseParams: []*types.KeyValuePair{
					{Key: keyTerminal, Value: "false"},
					{Key: keyInteractive, Value: "false"},
				},
			},
		},
		"test_io_config_params_no_tty_with_openstdin": {
			ioConfig: &ctrtypes.IOConfig{OpenStdin: true},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyInteractive, Value: "true"},
				},
				verboseParams: []*types.KeyValuePair{
					{Key: keyTerminal, Value: "false"},
				},
			},
		},
		"test_io_config_params_with_tty_no_openstdin": {
			ioConfig: &ctrtypes.IOConfig{Tty: true},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyTerminal, Value: "true"},
				},
				verboseParams: []*types.KeyValuePair{
					{Key: keyInteractive, Value: "false"},
				},
			},
		},
		"test_io_config_params_with_tty_with_openstdin": {
			ioConfig: &ctrtypes.IOConfig{Tty: true, OpenStdin: true},
			expectedParams: testExpectedParams{
				nonVerboseParams: []*types.KeyValuePair{
					{Key: keyTerminal, Value: "true"},
					{Key: keyInteractive, Value: "true"},
				},
			},
		},
	}
	for _, verbose := range []bool{true, false} {
		for testName, testCase := range testCases {
			t.Run(fullTestName(testName, verbose), func(t *testing.T) {
				assertParameters(t, testCase.expectedParams, ioConfigParameters(testCase.ioConfig, verbose), verbose)
			})
		}
	}
}

func createTestContainer(name string) *ctrtypes.Container {
	container := &ctrtypes.Container{
		Name:       name,
		Image:      ctrtypes.Image{Name: "my-container-registry.com/" + name + ":v1.2.3"},
		HostConfig: &ctrtypes.HostConfig{Privileged: true},
		Mounts:     []ctrtypes.MountPoint{{Source: "/etc", Destination: "/etc", PropagationMode: ctrtypes.RPrivatePropagationMode}},
		Config:     &ctrtypes.ContainerConfiguration{},
		State:      &ctrtypes.State{},
	}
	util.FillDefaults(container)
	return container
}

func assertParameters(t *testing.T, expectedParams testExpectedParams, actualParams []*types.KeyValuePair, verbose bool) {
	expected := expectedParams.nonVerboseParams
	if verbose {
		expected = append(expected, expectedParams.verboseParams...)
	}
	testutil.AssertEqual(t, len(expected), len(actualParams))
	for _, param := range expected {
		assertParameter(t, actualParams, param.Key, param.Value)
	}
}

func assertParameter(t *testing.T, params []*types.KeyValuePair, key string, value string) {
	for _, kv := range params {
		if kv.Key == key {
			if value != kv.Value {
				t.Errorf("param '%s' has wrong value: expected %s , got %s", key, value, kv.Value)
				t.Fail()
			}
			return
		}
	}
	t.Errorf("expected param '%s' with value %s not present as key-value pair", key, value)
	t.Fail()
}

func assertMultipleParameter(t *testing.T, params []*types.KeyValuePair, key string, value string) {
	for _, kv := range params {
		if kv.Key == key && kv.Value == value {
			return
		}
	}
	t.Errorf("expected param '%s' with value %s not present as key-value pair", key, value)
	t.Fail()
}

func fullTestName(testName string, verbose bool) string {
	if !verbose {
		return testName
	}
	return testName + "_verbose"
}
