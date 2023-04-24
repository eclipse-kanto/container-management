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

package util

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

const (
	imageRef   = "docker.io/library/hello-world:latest"
	exitCode0  = int64(0)
	exitCode1  = int64(1)
	errorMsg   = "Some error message"
	pid        = int64(10000)
	pidDefault = int64(-1)
)

func TestSetContainerStatus(t *testing.T) {
	t.Run("test_set_container_status_created", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		SetContainerStatusCreated(ctr)
		if ctr.Created == "" {
			t.Errorf("created not set correctly")
		}
		if ctr.State.Pid != pidDefault {
			t.Errorf("pid not set correctly")
		}
		if ctr.State.StartedAt != "" {
			t.Errorf("startedAt not set correctly")
		}
		if ctr.State.Error != "" {
			t.Errorf("error not set correctly")
		}
		if ctr.State.ExitCode != exitCode0 {
			t.Errorf("exit code not set correctly")
		}
		if ctr.State.FinishedAt != "" {
			t.Errorf("finishedAt not set correctly")
		}
		if ctr.State.Exited == true {
			t.Errorf("exited flag not set correctly")
		}
		if ctr.State.Dead == true {
			t.Errorf("dead flag not set correctly")
		}
		if ctr.State.Restarting == true {
			t.Errorf("restarting flag not set correctly")
		}
		if ctr.State.Paused == true {
			t.Errorf("paused flag not set correctly")
		}
		if ctr.State.Running == true {
			t.Errorf("running flag not set correctly")
		}
		if ctr.State.Status != types.Created {
			t.Errorf("expected status created, but was %s", ctr.State.Status)
		}
		if ctr.State.OOMKilled == true {
			t.Errorf("OOM killed flag not set correctly")
		}
	})

	t.Run("test_set_container_status_running", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		pid := int64(18000)
		SetContainerStatusRunning(ctr, pid)
		if ctr.Created != "" {
			t.Errorf("created not set correctly")
		}
		if ctr.State.Pid != pid {
			t.Errorf("pid not set correctly")
		}
		if ctr.State.StartedAt == "" {
			t.Errorf("startedAt not set correctly")
		}
		if ctr.State.Error != "" {
			t.Errorf("error not set correctly")
		}
		if ctr.State.ExitCode != exitCode0 {
			t.Errorf("exit code not set correctly")
		}
		if ctr.State.FinishedAt != "" {
			t.Errorf("finishedAt not set correctly")
		}
		if ctr.State.Exited == true {
			t.Errorf("exited flag not set correctly")
		}
		if ctr.State.Dead == true {
			t.Errorf("dead flag not set correctly")
		}
		if ctr.State.Restarting == true {
			t.Errorf("restarting flag not set correctly")
		}
		if ctr.State.Paused == true {
			t.Errorf("paused flag not set correctly")
		}
		if ctr.State.Running == false {
			t.Errorf("running flag not set correctly")
		}
		if ctr.State.Status != types.Running {
			t.Errorf("expected status created, but was %s", ctr.State.Status)
		}
		if ctr.State.OOMKilled == true {
			t.Errorf("OOM killed flag not set correctly")
		}
	})
	t.Run("test_set_container_status_stopped", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		SetContainerStatusStopped(ctr, exitCode1, errorMsg)
		if ctr.Created != "" {
			t.Errorf("created not set correctly")
		}
		if ctr.State.Pid != pidDefault {
			t.Errorf("pid not set correctly")
		}
		if ctr.State.StartedAt != "" {
			t.Errorf("startedAt not set correctly")
		}
		if ctr.State.Error == "" {
			t.Errorf("error not set correctly")
		}
		if ctr.State.ExitCode == exitCode0 {
			t.Errorf("exit code not set correctly")
		}
		if ctr.State.FinishedAt == "" {
			t.Errorf("finishedAt not set correctly")
		}
		if ctr.State.Exited == true {
			t.Errorf("exited flag not set correctly")
		}
		if ctr.State.Dead == true {
			t.Errorf("dead flag not set correctly")
		}
		if ctr.State.Restarting == true {
			t.Errorf("restarting flag not set correctly")
		}
		if ctr.State.Paused == true {
			t.Errorf("paused flag not set correctly")
		}
		if ctr.State.Running == true {
			t.Errorf("running flag not set correctly")
		}
		if ctr.State.Status != types.Stopped {
			t.Errorf("expected status created, but was %s", ctr.State.Status)
		}
		if ctr.State.OOMKilled == true {
			t.Errorf("OOM killed flag not set correctly")
		}
	})

	t.Run("test_set_container_status_exited", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		SetContainerStatusExited(ctr, exitCode1, errorMsg, false)
		if ctr.Created != "" {
			t.Errorf("created not set correctly")
		}
		if ctr.State.Pid != pidDefault {
			t.Errorf("pid not set correctly")
		}
		if ctr.State.StartedAt != "" {
			t.Errorf("startedAt not set correctly")
		}
		if ctr.State.Error == "" {
			t.Errorf("error not set correctly")
		}
		if ctr.State.ExitCode == exitCode0 {
			t.Errorf("exit code not set correctly")
		}
		if ctr.State.FinishedAt == "" {
			t.Errorf("finishedAt not set correctly")
		}
		if ctr.State.Exited == false {
			t.Errorf("exited flag not set correctly")
		}
		if ctr.State.Dead == true {
			t.Errorf("dead flag not set correctly")
		}
		if ctr.State.Restarting == true {
			t.Errorf("restarting flag not set correctly")
		}
		if ctr.State.Paused == true {
			t.Errorf("paused flag not set correctly")
		}
		if ctr.State.Running == true {
			t.Errorf("running flag not set correctly")
		}
		if ctr.State.Status != types.Exited {
			t.Errorf("expected status created, but was %s", ctr.State.Status)
		}
		if ctr.State.OOMKilled == true {
			t.Errorf("OOM killed flag not set correctly")
		}
	})

	t.Run("test_set_container_status_exited_oom_killed", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		SetContainerStatusExited(ctr, exitCode1, errorMsg, true)
		if ctr.Created != "" {
			t.Errorf("created not set correctly")
		}
		if ctr.State.Pid != pidDefault {
			t.Errorf("pid not set correctly")
		}
		if ctr.State.StartedAt != "" {
			t.Errorf("startedAt not set correctly")
		}
		if ctr.State.Error == "" {
			t.Errorf("error not set correctly")
		}
		if ctr.State.ExitCode == exitCode0 {
			t.Errorf("exit code not set correctly")
		}
		if ctr.State.FinishedAt == "" {
			t.Errorf("finishedAt not set correctly")
		}
		if ctr.State.Exited == false {
			t.Errorf("exited flag not set correctly")
		}
		if ctr.State.Dead == true {
			t.Errorf("dead flag not set correctly")
		}
		if ctr.State.Restarting == true {
			t.Errorf("restarting flag not set correctly")
		}
		if ctr.State.Paused == true {
			t.Errorf("paused flag not set correctly")
		}
		if ctr.State.Running == true {
			t.Errorf("running flag not set correctly")
		}
		if ctr.State.Status != types.Exited {
			t.Errorf("expected status created, but was %s", ctr.State.Status)
		}
		if ctr.State.OOMKilled == false {
			t.Errorf("OOM killed flag not set correctly")
		}
	})

	t.Run("test_set_container_status_paused", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		SetContainerStatusPaused(ctr)
		if ctr.Created != "" {
			t.Errorf("created not set correctly")
		}
		if ctr.State.Pid != 0 {
			t.Errorf("pid not set correctly")
		}
		if ctr.State.StartedAt != "" {
			t.Errorf("startedAt not set correctly")
		}
		if ctr.State.Error != "" {
			t.Errorf("error not set correctly")
		}
		if ctr.State.ExitCode != exitCode0 {
			t.Errorf("exit code not set correctly")
		}
		if ctr.State.FinishedAt != "" {
			t.Errorf("finishedAt not set correctly")
		}
		if ctr.State.Exited == true {
			t.Errorf("exited flag not set correctly")
		}
		if ctr.State.Dead == true {
			t.Errorf("dead flag not set correctly")
		}
		if ctr.State.Restarting == true {
			t.Errorf("restarting flag not set correctly")
		}
		if ctr.State.Paused == false {
			t.Errorf("paused flag not set correctly")
		}
		if ctr.State.Running == true {
			t.Errorf("running flag not set correctly")
		}
		if ctr.State.Status != types.Paused {
			t.Errorf("expected status created, but was %s", ctr.State.Status)
		}
		if ctr.State.OOMKilled == true {
			t.Errorf("OOM killed flag not set correctly")
		}
	})
	t.Run("test_set_container_status_unpaused", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		SetContainerStatusUnpaused(ctr)
		if ctr.Created != "" {
			t.Errorf("created not set correctly")
		}
		if ctr.State.Pid != 0 {
			t.Errorf("pid not set correctly")
		}
		if ctr.State.StartedAt != "" {
			t.Errorf("startedAt not set correctly")
		}
		if ctr.State.Error != "" {
			t.Errorf("error not set correctly")
		}
		if ctr.State.ExitCode != exitCode0 {
			t.Errorf("exit code not set correctly")
		}
		if ctr.State.FinishedAt != "" {
			t.Errorf("finishedAt not set correctly")
		}
		if ctr.State.Exited == true {
			t.Errorf("exited flag not set correctly")
		}
		if ctr.State.Dead == true {
			t.Errorf("dead flag not set correctly")
		}
		if ctr.State.Restarting == true {
			t.Errorf("restarting flag not set correctly")
		}
		if ctr.State.Paused == true {
			t.Errorf("paused flag not set correctly")
		}
		if ctr.State.Running == false {
			t.Errorf("running flag not set correctly")
		}
		if ctr.State.Status != types.Running {
			t.Errorf("expected status created, but was %s", ctr.State.Status)
		}
		if ctr.State.OOMKilled == true {
			t.Errorf("OOM killed flag not set correctly")
		}
	})
	t.Run("test_set_container_status_dead", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		SetContainerStatusDead(ctr)
		if ctr.Created != "" {
			t.Errorf("created not set correctly")
		}
		if ctr.State.Pid != 0 {
			t.Errorf("pid not set correctly")
		}
		if ctr.State.StartedAt != "" {
			t.Errorf("startedAt not set correctly")
		}
		if ctr.State.Error != "" {
			t.Errorf("error not set correctly")
		}
		if ctr.State.ExitCode != exitCode0 {
			t.Errorf("exit code not set correctly")
		}
		if ctr.State.FinishedAt != "" {
			t.Errorf("finishedAt not set correctly")
		}
		if ctr.State.Exited == true {
			t.Errorf("exited flag not set correctly")
		}
		if ctr.State.Dead == false {
			t.Errorf("dead flag not set correctly")
		}
		if ctr.State.Restarting == true {
			t.Errorf("restarting flag not set correctly")
		}
		if ctr.State.Paused == true {
			t.Errorf("paused flag not set correctly")
		}
		if ctr.State.Running == true {
			t.Errorf("running flag not set correctly")
		}
		if ctr.State.Status != types.Dead {
			t.Errorf("expected status created, but was %s", ctr.State.Status)
		}
		if ctr.State.OOMKilled == true {
			t.Errorf("OOM killed flag not set correctly")
		}
	})
}

// type isRestartPolicyX func(*types.RestartPolicy) bool
func TestRestartPolicyChecks(t *testing.T) {
	tests := map[string]struct {
		policy   *types.RestartPolicy
		isFunc   func(*types.RestartPolicy) bool
		expected bool
	}{
		"test_is_restart_policy_always_nil": {
			isFunc:   IsRestartPolicyAlways,
			expected: false,
		},
		"test_is_restart_policy_always": {
			policy: &types.RestartPolicy{
				Type: types.Always,
			},
			isFunc:   IsRestartPolicyAlways,
			expected: true,
		},
		"test_is_restart_policy_none_nil": {
			policy:   nil,
			isFunc:   IsRestartPolicyNone,
			expected: true,
		},
		"test_is_restart_policy_none": {
			policy: &types.RestartPolicy{
				Type: types.No,
			},
			isFunc:   IsRestartPolicyNone,
			expected: true,
		},
		"test_is_restart_policy_unless_stopped": {
			policy:   nil,
			isFunc:   IsRestartPolicyUnlessStopped,
			expected: false,
		},
		"test_is_restart_policy_on_failure": {
			policy: &types.RestartPolicy{
				Type: types.OnFailure,
			},
			isFunc:   IsRestartPolicyOnFailure,
			expected: true,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			actual := testCase.isFunc(testCase.policy)
			if actual != testCase.expected {
				t.Error("incorrect policy assertion")
			}
		})
	}
}

func TestFillDefaultsOnCleanContainer(t *testing.T) {
	ctr := &types.Container{}
	FillDefaults(ctr)

	t.Run("test_fill_defaults_id", func(t *testing.T) {
		if ctr.ID == "" {
			t.Error("container id not generated")
		}
	})

	t.Run("test_fill_defaults_name", func(t *testing.T) {
		if ctr.Name != ctr.ID {
			t.Error("container name not set")
		}
	})

	t.Run("test_fill_defaults_domain", func(t *testing.T) {
		if ctr.DomainName != fmt.Sprintf("%s-domain", ctr.Name) {
			t.Error("container domain not set")
		}
	})

	t.Run("test_fill_defaults_host_name", func(t *testing.T) {
		if ctr.HostName != fmt.Sprintf("%s-host", ctr.Name) {
			t.Error("container domain not set")
		}
	})

	t.Run("test_fill_defaults_host_config", func(t *testing.T) {
		if ctr.HostConfig == nil {
			t.Error("container host config not set")
		}
		if ctr.HostConfig.NetworkMode != types.NetworkModeBridge {
			t.Error("container host config default network not set to bridge")
		}
		if ctr.HostConfig.Privileged {
			t.Error("container host config must not be privileged")
		}
	})

	t.Run("test_fill_defaults_host_config_log_config", func(t *testing.T) {
		if ctr.HostConfig.LogConfig == nil {
			t.Error("container host config log config not set")
		}
		if ctr.HostConfig.LogConfig.DriverConfig == nil {
			t.Error("container host config log config driver config not set")
		}
		if ctr.HostConfig.LogConfig.DriverConfig.Type != types.LogConfigDriverJSONFile {
			t.Errorf("container host config log config unexpected type: %s", ctr.HostConfig.LogConfig.DriverConfig.Type)
		}
		if ctr.HostConfig.LogConfig.DriverConfig.MaxFiles != jsonFileLogConfigDefaultMaxFile {
			t.Errorf("container host config log config unexpected max files: %d", ctr.HostConfig.LogConfig.DriverConfig.MaxFiles)
		}
		if ctr.HostConfig.LogConfig.DriverConfig.MaxSize != jsonFileLogConfigDefaultMaxSize {
			t.Errorf("container host config log config unexpected max size: %s", ctr.HostConfig.LogConfig.DriverConfig.MaxSize)
		}
		if ctr.HostConfig.LogConfig.ModeConfig == nil {
			t.Error("container host config log config mode config not set")
		}
		if ctr.HostConfig.LogConfig.ModeConfig.Mode != types.LogModeBlocking {
			t.Errorf("container host config log config mode config unexpected type: %v", ctr.HostConfig.LogConfig.ModeConfig.Mode)
		}
		if ctr.HostConfig.LogConfig.ModeConfig.MaxBufferSize != "" {
			t.Errorf("container host config log config mode config unexpected max buff size: %v", ctr.HostConfig.LogConfig.ModeConfig.MaxBufferSize)
		}
	})

	t.Run("test_fill_defaults_host_config_port_mappings", func(t *testing.T) {
		ctrPorts := &types.Container{
			HostConfig: &types.HostConfig{
				PortMappings: []types.PortMapping{
					{
						HostIP:        "",
						Proto:         "",
						ContainerPort: hostConfigContainerPort,
						HostPort:      hostConfigHostPort,
						HostPortEnd:   0,
					},
				},
			},
		}
		FillDefaults(ctrPorts)
		if ctrPorts.HostConfig.PortMappings[0].HostIP != "0.0.0.0" {
			t.Error("container host config port mapping default host ip not set")
		}
		if ctrPorts.HostConfig.PortMappings[0].HostPortEnd != ctrPorts.HostConfig.PortMappings[0].HostPort {
			t.Error("container host config port mapping default host port end not set")
		}
		if ctrPorts.HostConfig.PortMappings[0].Proto != "tcp" {
			t.Error("container host config port mapping default proto not set")
		}
	})

	t.Run("test_fill_defaults_host_config_restart_policy", func(t *testing.T) {
		if ctr.HostConfig.RestartPolicy == nil {
			t.Error("container host config restart policy not set")
		}
		if ctr.HostConfig.RestartPolicy.Type != types.UnlessStopped {
			t.Errorf("container host config unexpected restart policy: %s", ctr.HostConfig.RestartPolicy.Type)
		}

	})

	t.Run("test_fill_defaults_host_config_devices", func(t *testing.T) {
		ctrDevices := &types.Container{
			HostConfig: &types.HostConfig{
				Devices: []types.DeviceMapping{{
					PathOnHost:        hostConfigDeviceHost,
					PathInContainer:   "",
					CgroupPermissions: "",
				}},
			},
		}
		FillDefaults(ctrDevices)
		if ctrDevices.HostConfig.Devices[0].PathInContainer != hostConfigDeviceHost {
			t.Error("container host config device config host path in container not set")
		}
		if ctrDevices.HostConfig.Devices[0].CgroupPermissions != "rwm" {
			t.Errorf("container host config device config unexpected cgroup permissions: %s", ctrDevices.HostConfig.Devices[0].CgroupPermissions)
		}
	})

	t.Run("test_fill_defaults_host_config_devices", func(t *testing.T) {
		ctrMounts := &types.Container{
			Mounts: []types.MountPoint{{
				Destination:     mountDest,
				Source:          mountSrc,
				PropagationMode: "",
			}},
		}
		FillDefaults(ctrMounts)
		if ctrMounts.Mounts[0].PropagationMode != types.RPrivatePropagationMode {
			t.Errorf("container mount config unexpected propagation mode: %s", ctrMounts.Mounts[0].PropagationMode)
		}
	})

	t.Run("test_fill_defaults_io_config", func(t *testing.T) {
		if ctr.IOConfig == nil {
			t.Error("container io config not set")
		}
		if ctr.IOConfig.StdinOnce {
			t.Error("container io config unexpected StdinOnce")
		}
		if ctr.IOConfig.OpenStdin {
			t.Error("container io config unexpected OpenStdin")
		}
		if ctr.IOConfig.AttachStdout {
			t.Error("container io config unexpected AttachStdout")
		}
		if ctr.IOConfig.AttachStdin {
			t.Error("container io config unexpected AttachStdin")
		}
		if ctr.IOConfig.AttachStderr {
			t.Error("container io config unexpected AttachStderr")
		}
		if ctr.IOConfig.Tty {
			t.Error("container io config unexpected Tty")
		}
	})
}

func TestFillResources(t *testing.T) {
	t.Run("test_fill_resources_with_swap", func(t *testing.T) {
		ctrResources := &types.Container{
			HostConfig: &types.HostConfig{
				Resources: &types.Resources{
					Memory:     "200M",
					MemorySwap: "300M",
				},
			},
		}
		FillMemorySwap(ctrResources)
		if ctrResources.HostConfig.Resources.MemorySwap != "300M" {
			t.Errorf("container resources unexpected swap memory limit: %s", ctrResources.HostConfig.Resources.MemorySwap)
		}
	})
	t.Run("test_fill_resources_missing_swap", func(t *testing.T) {
		ctrResources := &types.Container{
			HostConfig: &types.HostConfig{
				Resources: &types.Resources{
					Memory: "200M",
				},
			},
		}
		FillMemorySwap(ctrResources)
		if ctrResources.HostConfig.Resources.MemorySwap != "400M" {
			t.Errorf("container resources unexpected swap memory limit: %s", ctrResources.HostConfig.Resources.MemorySwap)
		}
	})
	t.Run("test_fill_resources_unlimited_swap", func(t *testing.T) {
		ctrResources := &types.Container{
			HostConfig: &types.HostConfig{
				Resources: &types.Resources{
					Memory:     "200M",
					MemorySwap: types.MemoryUnlimited,
				},
			},
		}
		FillMemorySwap(ctrResources)
		if ctrResources.HostConfig.Resources.MemorySwap != "" {
			t.Errorf("container resources unexpected swap memory limit: %s", ctrResources.HostConfig.Resources.MemorySwap)
		}
	})
}

func TestCalculateUptime(t *testing.T) {
	t.Run("test_calculate_uptime_no_state", func(t *testing.T) {
		ctr := &types.Container{}
		uptime := CalculateUptime(ctr)
		if uptime != 0 {
			t.Errorf("expected uptime 0: got: %s", uptime)
		}
	})

	t.Run("test_calculate_uptime_state_created", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		SetContainerStatusCreated(ctr)

		//SetContainerStatusRunning(ctr, 1)
		SetContainerStatusStopped(ctr, 0, "")

		uptime := CalculateUptime(ctr)
		if uptime != 0 {
			t.Errorf("expected uptime 0, got: %s", uptime)
		}
	})

	t.Run("test_calculate_uptime_state_running_and_stopped", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		SetContainerStatusRunning(ctr, 1)
		SetContainerStatusStopped(ctr, 0, "")

		uptime := CalculateUptime(ctr)
		if uptime == 0 {
			t.Errorf("unexpected uptime 0, expected: %s", uptime)
		}
	})

	t.Run("test_calculate_uptime_state_running_and_exited", func(t *testing.T) {
		ctr := &types.Container{
			State: &types.State{},
		}
		SetContainerStatusRunning(ctr, 1)
		SetContainerStatusExited(ctr, 0, "", false)

		uptime := CalculateUptime(ctr)
		if uptime == 0 {
			t.Errorf("unexpected uptime 0, expected: %s", uptime)
		}
	})
}

func TestFillDefaultsAlreadyFilled(t *testing.T) {
	// TODO create new ctr with custom configs, must not be changed after calling FillDefaults()
	//ctr := &types.Container{}
	//FillDefaults(ctr)
}

func TestGetImageHost(t *testing.T) {
	testCases := map[string]struct {
		testImgRef   string
		expectedHost string
	}{
		"test_simple": {
			testImgRef:   "testhost/img:latest",
			expectedHost: "testhost",
		},
		"test_with_port": {
			testImgRef:   "testhost:456/img:latest",
			expectedHost: "testhost:456",
		},
		"test_simple_with_sub": {
			testImgRef:   "testhost/sub/img:latest",
			expectedHost: "testhost",
		},
		"test_with_port_with_sub": {
			testImgRef:   "testhost:456/sub/img:latest",
			expectedHost: "testhost:456",
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			testutil.AssertEqual(t, testData.expectedHost, GetImageHost(testData.testImgRef))
		})
	}
}

func TestReadContainer(t *testing.T) {
	const prefix = "container-management-test-"

	notExistPath := func() string {
		for i := 0; i < 100; i++ {
			path := filepath.Join(os.TempDir(), prefix+strconv.Itoa(rand.Int()))
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return path
			}
		}
		return ""
	}()
	testutil.AssertNotEqual(t, "", notExistPath)

	tmpFile, err := os.CreateTemp("", prefix)
	testutil.AssertNil(t, err)
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	ctr := &types.Container{Image: types.Image{Name: "test-image-name"}}
	data, err := json.Marshal(ctr)
	testutil.AssertNil(t, err)
	err = ioutil.WriteFile(tmpFile.Name(), data, 0644)
	testutil.AssertNil(t, err)

	testCases := map[string]struct {
		path        string
		expectedCtr *types.Container
		expectedErr error
	}{
		"test_open_error": {
			path:        notExistPath,
			expectedCtr: nil,
			expectedErr: log.NewErrorf("open %s: no such file or directory", notExistPath),
		},
		"test_read_error": {
			path:        "/",
			expectedCtr: nil,
			expectedErr: log.NewError("read /: is a directory"),
		},
		"test_no_error": {
			path:        tmpFile.Name(),
			expectedCtr: ctr,
			expectedErr: nil,
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			actualCtr, actualErr := ReadContainer(testData.path)
			testutil.AssertEqual(t, testData.expectedCtr, actualCtr)
			testutil.AssertError(t, testData.expectedErr, actualErr)
		})
	}
}

func TestAsNamedMap(t *testing.T) {
	t.Run("TestAsNamedMap", func(t *testing.T) {
		container := []*types.Container{
			{
				DomainName: "testDomainName1",
				Name:       "testName1",
			},
			{
				DomainName: "testDomainName2",
				Name:       "testName2",
			},
		}
		res := AsNamedMap(container)

		testutil.AssertEqual(t, "testDomainName1", res["testName1"].DomainName)
		testutil.AssertEqual(t, "testName1", res["testName1"].Name)
		testutil.AssertEqual(t, "testDomainName2", res["testName2"].DomainName)
		testutil.AssertEqual(t, "testName2", res["testName2"].Name)
		testutil.AssertEqual(t, "", res["testName2"].HostName)
	})
}
