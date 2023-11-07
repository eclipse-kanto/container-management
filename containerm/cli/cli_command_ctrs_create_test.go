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

package main

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/golang/mock/gomock"
)

const (
	// command flags
	createCmdFlagName                  = "name"
	createCmdFlagTerminal              = "t"
	createCmdFlagInteractive           = "i"
	createCmdFlagPrivileged            = "privileged"
	createCmdFlagContainerFile         = "file"
	createCmdFlagRestartPolicy         = "rp"
	createCmdFlagRestartPolicyMaxCount = "rp-cnt"
	createCmdFlagRestartPolicyTimeout  = "rp-to"
	createCmdFlagNetwork               = "network"
	createCmdFlagExtraHosts            = "hosts"
	createCmdFlagExtraCapabilities     = "cap-add"
	createCmdFlagDevices               = "devices"
	createCmdFlagMountPoints           = "mp"
	createCmdFlagPorts                 = "ports"
	createCmdFlagEnv                   = "e"
	createCmdFlagLogDriver             = "log-driver"
	createCmdFlagLogDriverMaxFiles     = "log-max-files"
	createCmdFlagLogDriverMaxSize      = "log-max-size"
	createCmdFlagLogDriverPath         = "log-path"
	createCmdFlagLogMode               = "log-mode"
	createCmdFlagLogModeMaxBufferSize  = "log-max-buffer-size"
	createCmdFlagMemory                = "memory"
	createCmdFlagMemoryReservation     = "memory-reservation"
	createCmdFlagMemorySwap            = "memory-swap"
	createCmdFlagKeys                  = "dec-keys"
	createCmdFlagDecRecipients         = "dec-recipients"

	// test input constants
	createContainerImageName = "host/group/image:latest"
)

var (
	// test input args
	createCmdArgs = []string{createContainerImageName}
)

// Tests --------------------
func TestCreateCmdInit(t *testing.T) {
	createCliTest := &createCommandTest{}
	createCliTest.init()

	execTestInit(t, createCliTest)
}

func TestCreateCmdSetupFlags(t *testing.T) {
	createCliTest := &createCommandTest{}
	createCliTest.init()

	expectedCfg := createConfig{
		name:          "",
		terminal:      true,
		interactive:   true,
		privileged:    true,
		containerFile: string("config.json"),
		restartPolicy: restartPolicy{
			kind:          string(types.Always),
			timeout:       10,
			maxRetryCount: 3,
		},
		network:           string(types.NetworkModeHost),
		extraHosts:        []string{"ctrhost:host_ip"},
		extraCapabilities: []string{"CAP_NET_ADMIN"},
		devices:           []string{"/dev/ttyACM0:/dev/ttyACM1:rwm"},
		mountPoints:       []string{"/proc:/proc:rprivate"},
		ports:             []string{"192.168.1.100:80-100:80/udp"},
		logDriver:         string(types.LogConfigDriverNone),
		logMaxFiles:       5,
		logMaxSize:        "200M",
		logRootDirPath:    "/",
		logMode:           string(types.LogModeNonBlocking),
		logMaxBufferSize:  "2M",
		resources: resources{
			memory:            "500M",
			memoryReservation: "300M",
			memorySwap:        "800M",
		},
		decKeys:       []string{"key_filepath:password"},
		decRecipients: []string{"pkcs7:cert_filepath"},
	}

	flagsToApply := map[string]string{
		createCmdFlagName:                  expectedCfg.name,
		createCmdFlagTerminal:              strconv.FormatBool(expectedCfg.terminal),
		createCmdFlagInteractive:           strconv.FormatBool(expectedCfg.interactive),
		createCmdFlagPrivileged:            strconv.FormatBool(expectedCfg.privileged),
		createCmdFlagContainerFile:         expectedCfg.containerFile,
		createCmdFlagRestartPolicy:         expectedCfg.restartPolicy.kind,
		createCmdFlagRestartPolicyMaxCount: strconv.Itoa(expectedCfg.restartPolicy.maxRetryCount),
		createCmdFlagRestartPolicyTimeout:  strconv.FormatInt(expectedCfg.restartPolicy.timeout, 10),
		createCmdFlagNetwork:               expectedCfg.network,
		createCmdFlagExtraHosts:            strings.Join(expectedCfg.extraHosts, ","),
		createCmdFlagExtraCapabilities:     strings.Join(expectedCfg.extraCapabilities, ","),
		createCmdFlagDevices:               strings.Join(expectedCfg.devices, ","),
		createCmdFlagMountPoints:           strings.Join(expectedCfg.mountPoints, ","),
		createCmdFlagPorts:                 strings.Join(expectedCfg.ports, ","),
		createCmdFlagLogDriver:             expectedCfg.logDriver,
		createCmdFlagLogDriverMaxFiles:     strconv.Itoa(expectedCfg.logMaxFiles),
		createCmdFlagLogDriverMaxSize:      expectedCfg.logMaxSize,
		createCmdFlagLogDriverPath:         expectedCfg.logRootDirPath,
		createCmdFlagLogMode:               expectedCfg.logMode,
		createCmdFlagLogModeMaxBufferSize:  expectedCfg.logMaxBufferSize,
		createCmdFlagMemory:                expectedCfg.memory,
		createCmdFlagMemoryReservation:     expectedCfg.memoryReservation,
		createCmdFlagMemorySwap:            expectedCfg.memorySwap,
		createCmdFlagKeys:                  strings.Join(expectedCfg.decKeys, ","),
		createCmdFlagDecRecipients:         strings.Join(expectedCfg.decRecipients, ","),
	}

	execTestSetupFlags(t, createCliTest, flagsToApply, expectedCfg)
}

func TestCreateCmdRun(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	createCliTest := &createCommandTest{}
	createCliTest.initWithCtrl(controller)

	execTestsRun(t, createCliTest)
}

// EOF Tests --------------------------

type createCommandTest struct {
	cliCommandTestBase
	cmdCreate *createCmd
}

func (createTc *createCommandTest) commandConfig() interface{} {
	return createTc.cmdCreate.config
}

func (createTc *createCommandTest) commandConfigDefault() interface{} {
	return createConfig{
		name:        "",
		terminal:    false,
		interactive: false,
		privileged:  false,
		restartPolicy: restartPolicy{
			kind:          "",
			timeout:       30,
			maxRetryCount: 1,
		},
		network:           string(types.NetworkModeBridge),
		extraHosts:        nil,
		extraCapabilities: nil,
		devices:           nil,
		mountPoints:       nil,
		ports:             nil,
		logDriver:         string(types.LogConfigDriverJSONFile),
		logMaxFiles:       2,
		logMaxSize:        "100M",
		logRootDirPath:    "",
		logMode:           string(types.LogModeBlocking),
		logMaxBufferSize:  "1M",
		resources: resources{
			memory:            "",
			memoryReservation: "",
			memorySwap:        "",
		},
	}
}

func (createTc *createCommandTest) prepareCommand(flagsCfg map[string]string) error {
	// setup command to test
	cmd := &createCmd{}
	createTc.cmdCreate, createTc.baseCmd = cmd, cmd

	createTc.cmdCreate.init(createTc.mockRootCommand)
	// setup command flags
	return setCmdFlags(flagsCfg, createTc.cmdCreate.cmd)
}

func (createTc *createCommandTest) runCommand(args []string) error {
	return createTc.cmdCreate.run(args)
}
func (createTc *createCommandTest) generateRunExecutionConfigs() map[string]testRunExecutionConfig {
	return map[string]testRunExecutionConfig{
		// Test devices config
		"test_create_devices": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagDevices: "/dev/ttyACM0:/dev/ttyACM1:rwm",
			},
			mockExecution: createTc.mockExecCreateDevices,
		},
		"test_create_devices_with_privileged": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPrivileged: "true",
				createCmdFlagDevices:    "/dev/ttyACM0:/dev/ttyACM1:rwm",
			},
			mockExecution: createTc.mockExecCreateDevicesWithPrivileged,
		},
		"test_create_devices_incorrect_format_missing_target": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagDevices: "/dev/ttyACM0",
			},
			mockExecution: createTc.mockExecCreateDevicesErrConfigFormat,
		},
		"test_create_devices_incorrect_format_too_many_args": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagDevices: "/dev/ttyACM0:/dev/ttyACM1:rwm:test",
			},
			mockExecution: createTc.mockExecCreateDevicesErrConfigFormat,
		},
		"test_create_devices_incorrect_cgroup_format_too_many": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagDevices: "/dev/ttyACM0:/dev/ttyACM1:rwmm",
			},
			mockExecution: createTc.mockExecCreateDevicesErrCgroupFormat,
		},
		/*		"test_create_devices_incorrect_cgroup_format_too_few_2": { // TODO fix in command code
					args: createCmdArgs,
					flags: map[string]string{
						createCmdFlagDevices: "/dev/ttyACM0:/dev/ttyACM1:rw",
					},
				},
				"test_create_devices_incorrect_cgroup_format_too_few_1": { // TODO fix in command code
					args: createCmdArgs,
					flags: map[string]string{
						createCmdFlagDevices: "/dev/ttyACM0:/dev/ttyACM1:r",
					},
					mockExecution: createTc.mockExecCreateDevicesErrCgroupFormat,
				},
				"test_create_devices_incorrect_cgroup_format_duplicates": { // TODO fix in command code
					args: createCmdArgs,
					flags: map[string]string{
						createCmdFlagDevices: "/dev/ttyACM0:/dev/ttyACM1:rrw",
					},
					mockExecution: createTc.mockExecCreateDevicesErrCgroupFormat,
				},*/
		// Test mount points config
		"test_create_mount_points": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagMountPoints: "/proc:/proc:rprivate",
			},
			mockExecution: createTc.mockExecCreateWithMountPoints,
		},
		"test_create_mount_points_args_too_few": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagMountPoints: "/proc",
			},
			mockExecution: createTc.mockExecCreateWithMountPointsErrIncorrectParams,
		},
		"test_create_mount_points_args_too_many": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagMountPoints: "/proc:/proc:rprivate:test",
			},
			mockExecution: createTc.mockExecCreateWithMountPointsErrIncorrectParams,
		},
		/*		"test_create_mount_points_args_wrong_prop_mode": { // TODO fix in command code
				args: createCmdArgs,
				flags: map[string]string{
					createCmdFlagMountPoints: "/proc:/proc:rprivatee",
				},
				mockExecution: createTc.mockExecCreateWithMountPointsErrIncorrectParams,
			},*/
		"test_create_mount_points_default_prop_mode_set": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagMountPoints: "/proc:/proc",
			},
			mockExecution: createTc.mockExecCreateWithMountPoints,
		},
		// Test port mappings
		"test_create_port_mappings": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "192.168.1.100:80-100:80/udp",
			},
			mockExecution: createTc.mockExecCreateWithPortsFull,
		},
		"test_create_port_mappings_default": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "80:80",
			},
			mockExecution: createTc.mockExecCreateWithPortsDefault,
		},
		"test_create_port_mappings_range": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "80-100:80",
			},
			mockExecution: createTc.mockExecCreateWithPortsRange,
		},
		"test_create_port_mappings_range_and_ip": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "192.168.1.101:80-100:80",
			},
			mockExecution: createTc.mockExecCreateWithPortsRangeAndIP,
		},
		"test_create_port_mappings_range_and_proto": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "80-100:80/udp",
			},
			mockExecution: createTc.mockExecCreateWithPortsRangeAndProto,
		},
		"test_create_port_mappings_ip": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "192.168.1.100:80:80",
			},
			mockExecution: createTc.mockExecCreateWithPortsIP,
		},
		"test_create_port_mappings_ip_and_proto": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "192.168.1.100:80:80/udp",
			},
			mockExecution: createTc.mockExecCreateWithPortsProtoAndIP,
		},
		"test_create_port_mappings_proto": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "80:80/udp",
			},
			mockExecution: createTc.mockExecCreateWithPortsProto,
		},
		"test_create_port_mappings_args_too_few": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "80",
			},
			mockExecution: createTc.mockExecCreateWithPortsIncorrectPortsConfig,
		},
		/*		"test_create_port_mappings_args_too_many": { // TODO fix in command code
					args: createCmdArgs,
					flags: map[string]string{
						createCmdFlagPorts: "80:80/udp/abc",
					},
					mockExecution: createTc.mockExecCreateWithPortsIncorrectPortsConfig,
				},
				"test_create_port_mappings_invalid_protocol": { // TODO fix in command code
					args: createCmdArgs,
					flags: map[string]string{
						createCmdFlagPorts: "80:80/udpppp",
					},
					mockExecution: createTc.mockExecCreateWithPortsIncorrectPortsConfig,
				},*/
		"test_create_port_mappings_args_too_many_clns": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "80:80:80:80/udp",
			},
			mockExecution: createTc.mockExecCreateWithPortsIncorrectPortsConfig,
		},
		/*		"test_create_port_mappings_invalid_host_port": { // TODO fix in command code
				args:       createCmdArgs,
				flags: map[string]string{
					createCmdFlagPorts: "a0:80",
				},
				mockExecution: createTc.mockExecCreateWithPortsIncorrectPortsConfigParseErr,
			},*/
		"test_create_port_mappings_invalid_ctr_port": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "80:8a",
			},
			mockExecution: createTc.mockExecCreateWithPortsIncorrectPortsConfigParseErr,
		},
		"test_create_port_mappings_ip_and_invalid_ctr_port": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "192.168.1.100:80:8a",
			},
			mockExecution: createTc.mockExecCreateWithPortsIncorrectPortsConfigParseErr,
		},
		"test_create_port_mappings_ip_range_and_invalid_ctr_port": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "192.168.1.100:80-100:8a",
			},
			mockExecution: createTc.mockExecCreateWithPortsIncorrectPortsConfigParseErr,
		},
		"test_create_port_mappings_invalid_host_port_end": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "80-10a:80",
			},
			mockExecution: createTc.mockExecCreateWithPortsIncorrectHostRange,
		},
		"test_create_port_mappings_invalid_host_ip": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPorts: "abv:80-100:80",
			},
			mockExecution: createTc.mockExecCreateWithPortsIncorrectHostIP,
		},
		// Test extra hosts
		"test_create_extra_hosts": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagExtraHosts: "ctrhost:host_ip",
			},
			mockExecution: createTc.mockExecCreateWithExtraHosts,
		},
		/*		"test_create_extra_hosts_args_too_few": { // TODO fix in command code
					args: createCmdArgs,
					flags: map[string]string{
						createCmdFlagExtraHosts: "ctrhost",
					},
					mockExecution: createTc.mockExecCreateWithExtraHostsIncorrectConfig,
				},
				"test_create_extra_hosts_args_too_many": { // TODO fix in command code
					args: createCmdArgs,
					flags: map[string]string{
						createCmdFlagExtraHosts: "ctrhost:host_ip:123",
					},
					mockExecution: createTc.mockExecCreateWithExtraHostsIncorrectConfig,
				},*/
		// Test extra capabilities
		"test_create_extra_capabilities": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagExtraCapabilities: "CAP_NET_ADMIN",
			},
			mockExecution: createTc.mockExecCreateWithExtraCapabilities,
		},
		"test_create_extra_capabilities_with_privileged": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagExtraCapabilities: "CAP_NET_ADMIN",
				createCmdFlagPrivileged:        "true",
			},
			mockExecution: createTc.mockExecCreateWithExtraCapabilitiesWithPrivileged,
		},
		// Test privileged
		"test_create_privileged": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagPrivileged: "true",
			},
			mockExecution: createTc.mockExecCreateWithPrivileged,
		},
		// Test container file
		"test_create_no_args": {
			mockExecution: createTc.mockExecCreateWithNoArgs,
		},
		"test_create_container_file": {
			flags: map[string]string{
				createCmdFlagContainerFile: "../pkg/testutil/config/container/valid.json",
			},
			mockExecution: createTc.mockExecCreateContainerFile,
		},
		"test_create_container_file_invalid_path": {
			flags: map[string]string{
				createCmdFlagContainerFile: "/test/test",
			},
			mockExecution: createTc.mockExecCreateContainerFileInvalidPath,
		},
		"test_create_container_file_invalid_json": {
			flags: map[string]string{
				createCmdFlagContainerFile: "../pkg/testutil/config/container/invalid.json",
			},
			mockExecution: createTc.mockExecCreateContainerFileInvalidJSON,
		},
		"test_create_container_file_with_args": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagContainerFile: "../pkg/testutil/config/container/valid.json",
			},
			mockExecution: createTc.mockExecCreateContainerFileWithArgs,
		},
		// Test terminal
		"test_create_terminal": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagTerminal: "true",
			},
			mockExecution: createTc.mockExecCreateWithTerminal,
		},
		// Test interactive
		"test_create_interactive": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagInteractive: "true",
			},
			mockExecution: createTc.mockExecCreateWithInteractive,
		},
		// Test restart policy
		"test_create_restart_policy": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagRestartPolicy: string(types.Always),
			},
			mockExecution: createTc.mockExecCreateWithRestartPolicyDefault,
		},
		"test_create_restart_policy_max_retry": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagRestartPolicy:         string(types.OnFailure),
				createCmdFlagRestartPolicyMaxCount: "5",
			},
			mockExecution: createTc.mockExecCreateWithRestartPolicyMaxRetry,
		},
		"test_create_restart_policy_max_retry_and_timeout": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagRestartPolicy:         string(types.OnFailure),
				createCmdFlagRestartPolicyMaxCount: "5",
				createCmdFlagRestartPolicyTimeout:  "5",
			},
			mockExecution: createTc.mockExecCreateWithRestartPolicyMaxRetryAndTimeout,
		},
		"test_create_restart_policy_no_always_flags_ignored": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagRestartPolicy:         string(types.No),
				createCmdFlagRestartPolicyMaxCount: "5",
				createCmdFlagRestartPolicyTimeout:  "5",
			},
			mockExecution: createTc.mockExecCreateWithRestartPolicyWhenNoFlagsAreIgnored,
		},
		// Test env vars
		"test_create_env_vars_err_format_start_digit": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagEnv: "1VAR=1",
			},
			mockExecution: createTc.mockExecEnvVarIncorrectFormatStartDigit,
		},
		"test_create_env_vars_err_format_start_other": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagEnv: "$VAR=1",
			},
			mockExecution: createTc.mockExecEnvVarIncorrectFormatStartOther,
		},
		"test_create_env_vars_err_format_contains_bad_symbol": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagEnv: "V@R=1",
			},
			mockExecution: createTc.mockExecEnvVarIncorrectFormatContainsBasSymbol,
		},
		"test_create_env_vars_default": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagEnv: "VAR=1",
			},
			mockExecution: createTc.mockExecEnvVarDefault,
		},
		"test_create_env_vars_no_val": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagEnv: "VAR=",
			},
			mockExecution: createTc.mockExecEnvVarNoVal,
		},
		"test_create_env_vars_no_eq": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagEnv: "VAR",
			},
			mockExecution: createTc.mockExecEnvVarNoEq,
		},
		"test_create_env_vars_with_underscore": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagEnv: "_VAR=1",
			},
			mockExecution: createTc.mockExecEnvVarWithUnderscore,
		},
		"test_create_env_vars_with_comma": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagEnv: "VAR=1,2",
			},
			mockExecution: createTc.mockExecEnvVarWithComma,
		},
		"test_create_env_vars_with_string": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagEnv: "VAR=\"a b\"",
			},
			mockExecution: createTc.mockExecEnvVarWithString,
		},
		// Test args
		"test_create_args_command": {
			args:          append(createCmdArgs, "echo"),
			mockExecution: createTc.mockExecArgsDefault,
		},
		"test_create_args_command_with_arguments": {
			args:          append(createCmdArgs, "echo", "test", "execution!"),
			mockExecution: createTc.mockExecArgsDefault,
		},
		"test_create_name": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagName: "test-name",
			},
			mockExecution: createTc.mockExecCreateName,
		},
		"test_create_name_invalid_char": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagName: "test-name/",
			},
			mockExecution: createTc.mockExecCreateNameInvalidChar,
		},
		"test_create_name_invalid_at_start": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagName: "-test-name",
			},
			mockExecution: createTc.mockExecCreateNameInvalidCharAtStart,
		},
		// Test log
		"test_create_log_none": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagLogDriver: string(types.LogConfigDriverNone),
			},
			mockExecution: createTc.mockExecCreateLogDriverNone,
		},
		"test_create_log_fully_configured": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagLogDriverMaxSize:     "10M",
				createCmdFlagLogDriverMaxFiles:    "5",
				createCmdFlagLogDriverPath:        "/",
				createCmdFlagLogMode:              string(types.LogModeNonBlocking),
				createCmdFlagLogModeMaxBufferSize: "5M",
			},
			mockExecution: createTc.mockExecCreateLogFullyConfigured,
		},
		"test_create_log_mode_blocking_buff_ignored": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagLogMode:              string(types.LogModeBlocking),
				createCmdFlagLogModeMaxBufferSize: "5M",
			},
			mockExecution: createTc.mockExecCreateLogBlockingBuffFlagIgnored,
		},
		// Test network mode
		"test_create_network_mode_host": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagNetwork: string(types.NetworkModeHost),
			},
			mockExecution: createTc.mockExecCreateNetworkModeHost,
		},
		"test_create_network_mode_with_key_used": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagNetwork:    string(types.NetworkModeHost),
				createCmdFlagExtraHosts: "ctrhost:host_ip",
			},
			mockExecution: createTc.mockExecCreateNetworkModeHostReservedKeyUsed,
		},
		"test_create_network_mode_with_key_and_net_if_used": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagNetwork:    string(types.NetworkModeHost),
				createCmdFlagExtraHosts: "ctrhost:host_ip_eth0",
			},
			mockExecution: createTc.mockExecCreateNetworkModeHostReservedKeyUsed,
		},
		"test_create_network_mode_invalid": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagNetwork: "custom",
			},
			mockExecution: createTc.mockExecCreateNetworkModeInvalid,
		},
		// Tests default create
		"test_create_ID_and_image_default": {
			args:          createCmdArgs,
			mockExecution: createTc.mockExecCreateDefault,
		},
		// Test memory
		"test_create_memory_fully_configured": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagMemory:            "200M",
				createCmdFlagMemoryReservation: "100M",
				createCmdFlagMemorySwap:        "500M",
			},
			mockExecution: createTc.mockExecCreateMemoryFullyConfigured,
		},
		// Test decryption
		"test_create_decryption_configured": {
			args: createCmdArgs,
			flags: map[string]string{
				createCmdFlagKeys:          "key_filepath:password",
				createCmdFlagDecRecipients: "pkcs7:cert_filepath",
			},
			mockExecution: createTc.mockExecCreateImageDecryptionConfigured,
		},
	}
}

// Mocked executions ---------------------------------------------------------------------------------
func (createTc *createCommandTest) mockExecCreateDevices(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			Devices: []types.DeviceMapping{{
				PathOnHost:        "/dev/ttyACM0",
				PathInContainer:   "/dev/ttyACM1",
				CgroupPermissions: "rwm",
			}},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateDevicesWithPrivileged(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewErrorf("cannot create the container as privileged and with specified devices at the same time - choose one of the options")
}

func (createTc *createCommandTest) mockExecCreateDevicesErrConfigFormat(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("incorrect configuration value for device mapping")
}

func (createTc *createCommandTest) mockExecCreateDevicesErrCgroupFormat(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("incorrect cgroup permissions format for device mapping")
}

func (createTc *createCommandTest) mockExecCreateWithMountPoints(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		Mounts: []types.MountPoint{{
			Destination:     "/proc",
			Source:          "/proc",
			PropagationMode: string(types.RPrivatePropagationMode),
		}},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateWithMountPointsErrIncorrectParams(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("Incorrect number of parameters of the mount point")
}

func (createTc *createCommandTest) mockExecCreateWithPortsRange(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			PortMappings: []types.PortMapping{{
				ContainerPort: 80,
				HostPort:      80,
				HostPortEnd:   100,
			}},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateWithPortsDefault(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			PortMappings: []types.PortMapping{{
				ContainerPort: 80,
				HostPort:      80,
			}},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateWithPortsRangeAndIP(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			PortMappings: []types.PortMapping{{
				ContainerPort: 80,
				HostPort:      80,
				HostIP:        "192.168.1.101",
				HostPortEnd:   100,
			}},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateWithPortsRangeAndProto(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: createContainerImageName,
		},
		HostConfig: &types.HostConfig{
			PortMappings: []types.PortMapping{{
				Proto:         "udp",
				ContainerPort: 80,
				HostPort:      80,
				HostPortEnd:   100,
			}},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateWithPortsIP(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			PortMappings: []types.PortMapping{{
				ContainerPort: 80,
				HostPort:      80,
				HostIP:        "192.168.1.100",
			}},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateWithPortsProtoAndIP(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			PortMappings: []types.PortMapping{{
				Proto:         "udp",
				ContainerPort: 80,
				HostPort:      80,
				HostIP:        "192.168.1.100",
			}},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateWithPortsFull(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			PortMappings: []types.PortMapping{{
				Proto:         "udp",
				ContainerPort: 80,
				HostIP:        "192.168.1.100",
				HostPort:      80,
				HostPortEnd:   100,
			}},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateWithPortsProto(args []string) error {
	container := initExpectedCtr(&types.Container{

		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			PortMappings: []types.PortMapping{{
				Proto:         "udp",
				ContainerPort: 80,
				HostPort:      80,
			}},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateWithNoArgs(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("container image argument is expected")
}

func (createTc *createCommandTest) mockExecCreateWithPortsIncorrectPortsConfig(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("Incorrect port mapping configuration")
}

func (createTc *createCommandTest) mockExecCreateWithPortsIncorrectPortsConfigParseErr(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("Incorrect container port mapping configuration")
}

func (createTc *createCommandTest) mockExecCreateWithPortsIncorrectHostRange(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("Incorrect host range port mapping configuration")
}

func (createTc *createCommandTest) mockExecCreateWithPortsIncorrectHostIP(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("Incorrect host ip port mapping configuration")
}

func (createTc *createCommandTest) mockExecCreateWithExtraHosts(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			ExtraHosts: []string{"ctrhost:host_ip"},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateWithExtraHostsIncorrectConfig(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("Incorrect hosts configuration")
}

func (createTc *createCommandTest) mockExecCreateWithExtraCapabilities(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			ExtraCapabilities: []string{"CAP_NET_ADMIN"},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateWithExtraCapabilitiesWithPrivileged(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("cannot create the container as privileged and with extra capabilities at the same time - choose one of the options")
}

func (createTc *createCommandTest) mockExecCreateWithPrivileged(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			Privileged: true,
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateContainerFile(_ []string) error {
	byteValue, _ := os.ReadFile("../pkg/testutil/config/container/valid.json")
	container := initExpectedCtr(&types.Container{})
	json.Unmarshal(byteValue, container)

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateContainerFileInvalidPath(_ []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	_, err := os.ReadFile("/test/test")
	return err
}

func (createTc *createCommandTest) mockExecCreateContainerFileInvalidJSON(_ []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	byteValue, _ := os.ReadFile("../pkg/testutil/config/container/invalid.json")
	err := json.Unmarshal(byteValue, &types.Container{})
	return err
}

func (createTc *createCommandTest) mockExecCreateContainerFileWithArgs(_ []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("no arguments are expected when container is created from a file")
}

func (createTc *createCommandTest) mockExecCreateWithTerminal(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		IOConfig: &types.IOConfig{
			Tty: true,
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateWithInteractive(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		IOConfig: &types.IOConfig{
			OpenStdin: true,
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateWithRestartPolicyDefault(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			RestartPolicy: &types.RestartPolicy{
				Type: types.Always,
			},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateWithRestartPolicyMaxRetry(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			RestartPolicy: &types.RestartPolicy{
				MaximumRetryCount: 5,
				RetryTimeout:      time.Duration(30) * time.Second,
				Type:              types.OnFailure,
			},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateWithRestartPolicyMaxRetryAndTimeout(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			RestartPolicy: &types.RestartPolicy{
				MaximumRetryCount: 5,
				RetryTimeout:      time.Duration(5) * time.Second,
				Type:              types.OnFailure,
			},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateWithRestartPolicyWhenNoFlagsAreIgnored(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			RestartPolicy: &types.RestartPolicy{
				Type: types.No,
			},
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateDefault(args []string) error {
	container := initExpectedCtr(&types.Container{

		Image: types.Image{
			Name: args[0],
		},
	})

	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func initExpectedCtr(ctr *types.Container) *types.Container {
	//merge default and provided
	if ctr.HostConfig == nil {
		ctr.HostConfig = &types.HostConfig{
			Privileged:        false,
			ExtraHosts:        nil,
			ExtraCapabilities: nil,
			NetworkMode:       types.NetworkModeBridge,
		}
	} else if ctr.HostConfig.NetworkMode == "" {
		ctr.HostConfig.NetworkMode = types.NetworkModeBridge
	}
	if ctr.IOConfig == nil {
		ctr.IOConfig = &types.IOConfig{
			Tty:       false,
			OpenStdin: false,
		}
	}
	if ctr.HostConfig.LogConfig == nil {
		ctr.HostConfig.LogConfig = &types.LogConfiguration{
			DriverConfig: &types.LogDriverConfiguration{
				Type:     types.LogConfigDriverJSONFile,
				MaxFiles: 2,
				MaxSize:  "100M",
			},
			ModeConfig: &types.LogModeConfiguration{
				Mode: types.LogModeBlocking,
			},
		}
	}
	return ctr
}

func (createTc *createCommandTest) mockExecEnvVarIncorrectFormatStartDigit(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewErrorf("invalid environmental variable declaration provided : %s", "1VAR=1")
}

func (createTc *createCommandTest) mockExecEnvVarIncorrectFormatStartOther(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewErrorf("invalid environmental variable declaration provided : %s", "$VAR=1")
}

func (createTc *createCommandTest) mockExecEnvVarIncorrectFormatContainsBasSymbol(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewErrorf("invalid environmental variable declaration provided : %s", "V@R=1")
}

func (createTc *createCommandTest) mockExecEnvVarDefault(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		Config: &types.ContainerConfiguration{
			Env: []string{"VAR=1"},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecEnvVarNoVal(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		Config: &types.ContainerConfiguration{
			Env: []string{"VAR="},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecEnvVarNoEq(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		Config: &types.ContainerConfiguration{
			Env: []string{"VAR"},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecEnvVarWithUnderscore(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		Config: &types.ContainerConfiguration{
			Env: []string{"_VAR=1"},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecEnvVarWithComma(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		Config: &types.ContainerConfiguration{
			Env: []string{"VAR=1,2"},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecEnvVarWithString(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		Config: &types.ContainerConfiguration{
			Env: []string{"VAR=\"a b\""},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecArgsDefault(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		Config: &types.ContainerConfiguration{
			Cmd: args[1:],
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateName(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		Name: "test-name",
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateNameInvalidChar(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewErrorf("invalid container name format : %s", "test-name/")
}

func (createTc *createCommandTest) mockExecCreateNameInvalidCharAtStart(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewErrorf("invalid container name format : %s", "-test-name")
}
func (createTc *createCommandTest) mockExecCreateLogDriverNone(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			LogConfig: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{
					Type: types.LogConfigDriverNone,
				},
				ModeConfig: &types.LogModeConfiguration{
					Mode: types.LogModeBlocking,
				},
			},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateLogFullyConfigured(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			LogConfig: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{
					Type:     types.LogConfigDriverJSONFile,
					MaxSize:  "10M",
					MaxFiles: 5,
					RootDir:  "/",
				},
				ModeConfig: &types.LogModeConfiguration{
					Mode:          types.LogModeNonBlocking,
					MaxBufferSize: "5M",
				},
			},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateLogBlockingBuffFlagIgnored(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			LogConfig: &types.LogConfiguration{
				DriverConfig: &types.LogDriverConfiguration{
					Type:     types.LogConfigDriverJSONFile,
					MaxSize:  "100M",
					MaxFiles: 2,
				},
				ModeConfig: &types.LogModeConfiguration{
					Mode: types.LogModeBlocking,
				},
			},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}

func (createTc *createCommandTest) mockExecCreateNetworkModeHost(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			NetworkMode: types.NetworkModeHost,
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateNetworkModeInvalid(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewErrorf("unsupported network mode custom")
}
func (createTc *createCommandTest) mockExecCreateNetworkModeHostReservedKeyUsed(args []string) error {
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Any()).Times(0)
	return log.NewError("cannot use the host_ip reserved key or any of its modifications when in host network mode")
}
func (createTc *createCommandTest) mockExecCreateMemoryFullyConfigured(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
		},
		HostConfig: &types.HostConfig{
			Resources: &types.Resources{
				Memory:            "200M",
				MemoryReservation: "100M",
				MemorySwap:        "500M",
			},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
func (createTc *createCommandTest) mockExecCreateImageDecryptionConfigured(args []string) error {
	container := initExpectedCtr(&types.Container{
		Image: types.Image{
			Name: args[0],
			DecryptConfig: &types.DecryptConfig{
				Keys:       []string{"key_filepath:password"},
				Recipients: []string{"pkcs7:cert_filepath"},
			},
		},
	})
	createTc.mockClient.EXPECT().Create(gomock.AssignableToTypeOf(context.Background()), gomock.Eq(container)).Times(1).Return(container, nil)
	return nil
}
