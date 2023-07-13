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
	"fmt"
	"strconv"
	"testing"
	"time"

	ctrtypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/util"

	"github.com/eclipse-kanto/update-manager/api/types"
)

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
		assertParameter(t, node.Parameters, keyDomainName, testContainers[i].DomainName, false)
		assertParameter(t, node.Parameters, keyHostName, testContainers[i].HostName, false)
		assertParameter(t, node.Parameters, keyRestartCount, strconv.Itoa(testContainers[i].RestartCount), false)
		assertParameter(t, node.Parameters, keyCreated, testContainers[i].Created, false)
		assertParameter(t, node.Parameters, keyManuallyStopped, strconv.FormatBool(testContainers[i].ManuallyStopped), false)
		assertParameter(t, node.Parameters, keyStartedSuccessfullyBefore, strconv.FormatBool(testContainers[i].StartedSuccessfullyBefore), false)
	}
}

func TestHostConfigParametersPrivileged(t *testing.T) {
	for _, verbose := range []bool{true, false} {
		for _, privileged := range []bool{true, false} {
			t.Run(fmt.Sprintf("test_host_config_params_privileged_%v_verbose_%v", privileged, verbose), func(t *testing.T) {
				hostConfig := &ctrtypes.HostConfig{Privileged: privileged}
				params := hostConfigParameters(hostConfig, verbose)
				lenExpected := 0
				if verbose || privileged {
					lenExpected++
					assertParameter(t, params, keyPrivileged, strconv.FormatBool(privileged), false)
				}
				testutil.AssertEqual(t, lenExpected, len(params))
			})
		}
	}
}

func TestHostConfigParametersRestartPolicy(t *testing.T) {
	testPolicies := []*ctrtypes.RestartPolicy{
		{Type: ctrtypes.No},
		{Type: ctrtypes.Always},
		{Type: ctrtypes.OnFailure, MaximumRetryCount: 5, RetryTimeout: 3 * time.Minute},
		{Type: ctrtypes.UnlessStopped},
	}
	for _, verbose := range []bool{true, false} {
		for _, policy := range testPolicies {
			t.Run(fmt.Sprintf("test_host_config_params_restart_policy_%v_verbose_%v", policy.Type, verbose), func(t *testing.T) {
				hostConfig := &ctrtypes.HostConfig{RestartPolicy: policy, Privileged: true}
				params := hostConfigParameters(hostConfig, verbose)
				lenExpected := 1 // +1 for privileged flag

				if verbose || policy.Type != ctrtypes.UnlessStopped { // unless-stopped is the default restart policy type
					lenExpected++
					assertParameter(t, params, keyRestartPolicy, string(policy.Type), false)
				}
				if verbose || policy.Type == ctrtypes.OnFailure {
					lenExpected++
					assertParameter(t, params, keyRestartMaxRetries, strconv.Itoa(hostConfig.RestartPolicy.MaximumRetryCount), false)
					lenExpected++
					assertParameter(t, params, keyRestartTimeout, hostConfig.RestartPolicy.RetryTimeout.String(), false)
				}
				testutil.AssertEqual(t, lenExpected, len(params))
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
		assertParameter(t, params, keyDevice, testDevice, true)
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
		assertParameter(t, params, keyPort, testPort, true)
	}
	testutil.AssertEqual(t, len(testPortMappings), len(params))
}

func TestHostConfigParametersNetwork(t *testing.T) {
	for _, verbose := range []bool{true, false} {
		for _, network := range []ctrtypes.NetworkMode{ctrtypes.NetworkModeBridge, ctrtypes.NetworkModeHost} {
			t.Run(fmt.Sprintf("test_host_config_params_network_%v_verbose_%v", network, verbose), func(t *testing.T) {
				hostConfig := &ctrtypes.HostConfig{NetworkMode: network, Privileged: true}
				params := hostConfigParameters(hostConfig, verbose)
				lenExpected := 1 // +1 for privileged flag
				if verbose || network != ctrtypes.NetworkModeBridge {
					lenExpected++
					assertParameter(t, params, keyNetwork, string(network), false)
				}
				testutil.AssertEqual(t, lenExpected, len(params))
			})
		}
	}
}

func TestHostConfigParametersExtraHosts(t *testing.T) {
	hostConfig := &ctrtypes.HostConfig{
		ExtraHosts: []string{"ctrhost:host_ip", "somehost:192.168.0.1"},
	}
	params := hostConfigParameters(hostConfig, false)
	for _, host := range hostConfig.ExtraHosts {
		assertParameter(t, params, keyHost, host, true)
	}
	testutil.AssertEqual(t, len(hostConfig.ExtraHosts), len(params))
}

func TestHostConfigParametersLogDriverConfig(t *testing.T) {
	testLogDriverConfigs := []*ctrtypes.LogDriverConfiguration{
		{Type: ctrtypes.LogConfigDriverNone},
		{Type: ctrtypes.LogConfigDriverJSONFile, MaxFiles: 6, MaxSize: "100K"},
		{Type: ctrtypes.LogConfigDriverJSONFile, MaxFiles: 2, MaxSize: "100M", RootDir: "/tmp/logs"},
	}
	for _, verbose := range []bool{true, false} {
		for _, logDriverConfig := range testLogDriverConfigs {
			t.Run(fmt.Sprintf("test_host_config_params_log_driver_config_%v_verbose_%v", logDriverConfig.Type, verbose), func(t *testing.T) {
				hostConfig := &ctrtypes.HostConfig{LogConfig: &ctrtypes.LogConfiguration{DriverConfig: logDriverConfig}, Privileged: true}
				params := hostConfigParameters(hostConfig, verbose)
				lenExpected := 1 // +1 for privileged flag

				if verbose || logDriverConfig.Type != ctrtypes.LogConfigDriverJSONFile { // json-file is the default log driver type
					lenExpected++
					assertParameter(t, params, keyLogDriver, string(logDriverConfig.Type), false)
				}
				if logDriverConfig.Type == ctrtypes.LogConfigDriverJSONFile {
					if verbose || logDriverConfig.MaxFiles != 2 { // 2 is the default value for max files
						lenExpected++
						assertParameter(t, params, keyLogMaxFiles, strconv.Itoa(logDriverConfig.MaxFiles), false)
					}
					if verbose || logDriverConfig.MaxSize != "100M" { // 100M is the default value for max size
						lenExpected++
						assertParameter(t, params, keyLogMaxSize, logDriverConfig.MaxSize, false)
					}
					if logDriverConfig.RootDir != "" {
						lenExpected++
						assertParameter(t, params, keyLogPath, logDriverConfig.RootDir, false)
					}
				}
				testutil.AssertEqual(t, lenExpected, len(params))
			})
		}
	}
}

func TestHostConfigParametersLogModeConfig(t *testing.T) {
	testLogModeConfigs := []*ctrtypes.LogModeConfiguration{
		{Mode: ctrtypes.LogModeBlocking},
		{Mode: ctrtypes.LogModeNonBlocking, MaxBufferSize: "100K"},
		{Mode: ctrtypes.LogModeNonBlocking, MaxBufferSize: "1M"},
	}
	for _, verbose := range []bool{true, false} {
		for _, logModeConfig := range testLogModeConfigs {
			t.Run(fmt.Sprintf("test_host_config_params_log_mode_config_%v_verbose_%v", logModeConfig.Mode, verbose), func(t *testing.T) {
				hostConfig := &ctrtypes.HostConfig{LogConfig: &ctrtypes.LogConfiguration{ModeConfig: logModeConfig}, Privileged: true}
				params := hostConfigParameters(hostConfig, verbose)
				lenExpected := 1 // +1 for privileged flag

				if verbose || logModeConfig.Mode != ctrtypes.LogModeBlocking { // blocking mode is the default log driver mode
					lenExpected++
					assertParameter(t, params, keyLogMode, string(logModeConfig.Mode), false)
				}
				if logModeConfig.Mode == ctrtypes.LogModeNonBlocking {
					if verbose || logModeConfig.MaxBufferSize != "1M" { // 1M is the default value for max buffer size
						lenExpected++
						assertParameter(t, params, keyLogMaxBufferSize, logModeConfig.MaxBufferSize, false)
					}
				}
				testutil.AssertEqual(t, lenExpected, len(params))
			})
		}
	}
}

func TestHostConfigParametersResources(t *testing.T) {
	testResources := []*ctrtypes.Resources{
		{Memory: "200m", MemoryReservation: "100m", MemorySwap: "50m"},
		{Memory: "300m", MemoryReservation: "300m"},
		{Memory: "1g", MemorySwap: "500m"},
	}
	for _, verbose := range []bool{true, false} {
		for _, resource := range testResources {
			t.Run(fmt.Sprintf("test_host_config_params_resources_memory_%v_verbose_%v", resource.Memory, verbose), func(t *testing.T) {
				hostConfig := &ctrtypes.HostConfig{Resources: resource, Privileged: true}
				params := hostConfigParameters(hostConfig, verbose)
				lenExpected := 1 // +1 for privileged flag

				if verbose || resource.Memory != "" {
					lenExpected++
					assertParameter(t, params, keyMemory, resource.Memory, false)
				}
				if verbose || resource.MemoryReservation != "" {
					lenExpected++
					assertParameter(t, params, keyMemoryReservation, resource.MemoryReservation, false)
				}
				if verbose || resource.MemorySwap != "" {
					lenExpected++
					assertParameter(t, params, keyMemorySwap, resource.MemorySwap, false)
				}
				testutil.AssertEqual(t, lenExpected, len(params))
			})
		}
	}
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
		assertParameter(t, params, keyMount, testMount, true)
	}
	testutil.AssertEqual(t, len(testMounts), len(params))
}

func TestContainerConfigParameters(t *testing.T) {
	for _, args := range [][]string{nil, {}, {"arg1, arg2, arg3"}} {
		for _, envs := range [][]string{nil, {}, {"ENV_X=VAL_A", "ENV_Y=VAL_B", "ENV_C="}} {
			t.Run(fmt.Sprintf("test_container_config_parameters_cmd_%v_env_%v", args, envs), func(t *testing.T) {
				params := containerConfigParameters(&ctrtypes.ContainerConfiguration{Env: envs, Cmd: args})
				for _, cmd := range args {
					assertParameter(t, params, keyCmd, cmd, true)
				}
				for _, env := range envs {
					assertParameter(t, params, keyEnv, env, true)
				}
				testutil.AssertEqual(t, len(args)+len(envs), len(params))
			})
		}
	}
}

func TestStateParameters(t *testing.T) {
	testCases := map[string]struct {
		testSetup func(*ctrtypes.Container)
	}{
		"test_state_params_container_creating": {
			testSetup: func(c *ctrtypes.Container) { c.State.Status = ctrtypes.Creating },
		},
		"test_state_params_container_created": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusCreated(c) },
		},
		"test_state_params_container_running": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusRunning(c, 1234) },
		},
		"test_state_params_container_stopped_normally": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusStopped(c, 0, "") },
		},
		"test_state_params_container_stopped_error": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusStopped(c, -1, "stopped with error") },
		},
		"test_state_params_container_paused": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusPaused(c) },
		},
		"test_state_params_container_exited": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusExited(c, 1234, "unexpected exit", false) },
		},
		"test_state_params_container_dead": {
			testSetup: func(c *ctrtypes.Container) { util.SetContainerStatusDead(c) },
		},
		"test_state_params_container_unknown": {
			testSetup: func(c *ctrtypes.Container) {
				c.State.Status = ctrtypes.Unknown
				c.State.FinishedAt = "2023-01-02T15:04:05.999999999Z07:00"
				c.State.ExitCode = 999
			},
		},
	}

	for _, verbose := range []bool{true, false} {
		for testName, testCase := range testCases {
			t.Run(testName+"_verbose_"+strconv.FormatBool(verbose), func(t *testing.T) {
				testContainer := createTestContainer("test-container")
				testContainer.State = &ctrtypes.State{}
				testCase.testSetup(testContainer)
				expStatus := testContainer.State.Status
				params := stateParameters(testContainer.State, verbose)
				lenExpected := 1
				assertParameter(t, params, keyStatus, expStatus.String(), false)
				if verbose || (testContainer.State.FinishedAt != "" && expStatus != ctrtypes.Running) {
					lenExpected++
					assertParameter(t, params, keyFinishedAt, testContainer.State.FinishedAt, false)
				}
				if verbose || (testContainer.State.ExitCode != 0 && expStatus != ctrtypes.Running) {
					lenExpected++
					assertParameter(t, params, keyExitCode, strconv.Itoa(int(testContainer.State.ExitCode)), false)
				}
				testutil.AssertEqual(t, lenExpected, len(params))
			})
		}
	}
}

func TestIOConfigParameters(t *testing.T) {
	for _, verbose := range []bool{true, false} {
		for _, tty := range []bool{true, false} {
			for _, openStdin := range []bool{true, false} {
				t.Run(fmt.Sprintf("test_io_config_params_tty=%v_openstdin=%v_verbose=%v", tty, openStdin, verbose), func(t *testing.T) {
					params := ioConfigParameters(&ctrtypes.IOConfig{Tty: tty, OpenStdin: openStdin}, verbose)
					lenExpected := 0
					if tty || verbose {
						lenExpected++
						assertParameter(t, params, keyTerminal, strconv.FormatBool(tty), false)
					}
					if openStdin || verbose {
						lenExpected++
						assertParameter(t, params, keyInteractive, strconv.FormatBool(openStdin), false)
					}
					testutil.AssertEqual(t, lenExpected, len(params))
				})
			}
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

func assertParameter(t *testing.T, params []*types.KeyValuePair, key string, value string, multiple bool) {
	for _, kv := range params {
		if kv.Key == key {
			if value == kv.Value {
				return
			}
			if !multiple {
				t.Errorf("param '%s' has wrong value: expected %s , got %s", key, value, kv.Value)
				t.Fail()
				return
			}
		}
	}
	t.Errorf("expected param '%s' with value %s not present as key-value pair", key, value)
	t.Fail()
}
