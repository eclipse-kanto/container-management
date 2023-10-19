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
	"testing"

	ctrtypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"

	"github.com/eclipse-kanto/update-manager/api/types"
)

func TestToContainerMinimalConfig(t *testing.T) {
	containerConfig := createSimpleDesiredComponent(testContainerName, testContainerVersion)
	container, err := toContainer(containerConfig)
	testutil.AssertNil(t, err)
	testutil.AssertNotEqual(t, "", container.ID)
	testutil.AssertEqual(t, testContainerName, container.Name)
	testutil.AssertEqual(t, testContainerName+":"+testContainerVersion, container.Image.Name)
	testutil.AssertEqual(t, testContainerName+"-domain", container.DomainName)
	testutil.AssertEqual(t, testContainerName+"-host", container.HostName)

	testutil.AssertEqual(t, false, container.HostConfig.Privileged)
	testutil.AssertEqual(t, ctrtypes.NetworkModeBridge, container.HostConfig.NetworkMode)
	testutil.AssertEqual(t, ctrtypes.UnlessStopped, container.HostConfig.RestartPolicy.Type)
	testutil.AssertEqual(t, 0, container.HostConfig.RestartPolicy.MaximumRetryCount)
	testutil.AssertNil(t, container.HostConfig.Resources)
	testutil.AssertNil(t, container.HostConfig.Devices)
	testutil.AssertNil(t, container.HostConfig.PortMappings)
	testutil.AssertNil(t, container.HostConfig.ExtraHosts)
	testutil.AssertEqual(t, ctrtypes.LogConfigDriverJSONFile, container.HostConfig.LogConfig.DriverConfig.Type)
	testutil.AssertEqual(t, 2, container.HostConfig.LogConfig.DriverConfig.MaxFiles)
	testutil.AssertEqual(t, "100M", container.HostConfig.LogConfig.DriverConfig.MaxSize)
	testutil.AssertEqual(t, ctrtypes.LogModeBlocking, container.HostConfig.LogConfig.ModeConfig.Mode)
	testutil.AssertEqual(t, "", container.HostConfig.LogConfig.ModeConfig.MaxBufferSize)
	testutil.AssertEqual(t, ctrtypes.IOConfig{}, *container.IOConfig)
	testutil.AssertNil(t, container.Mounts)
}

func TestToContainer(t *testing.T) {
	containerConfig := &types.ComponentWithConfig{
		Component: types.Component{ID: testContainerName, Version: testContainerVersion},
		Config: []*types.KeyValuePair{
			// device mappings
			{Key: "device", Value: "/dev/abc:/dev/ABC"}, // valid setting
			{Key: "device", Value: "/dev/xyz"},          // invalid setting, shall be ignored
			// port mappings
			{Key: "port", Value: "80:8888/tcp"},           // valid setting
			{Key: "port", Value: "123.456.789.000:80:80"}, // invalid setting, shall be ignored
			// mounts
			{Key: "mount", Value: "/tmp:/var/tmp"}, // valid setting
			{Key: "mount", Value: "/TMP"},          // invalid setting, shall be ignored
			// extra hosts
			{Key: "host", Value: "ctr_host"},
			{Key: "host", Value: "testhost"},
			// env & cmd
			{Key: "env", Value: "DEBUG=true"},
			{Key: "cmd", Value: "arg1"},
			{Key: "env", Value: "ENV1="},
			{Key: "cmd", Value: "arg2"},
			// restart policy
			{Key: "restartPolicy", Value: "on-failure"},
			{Key: "restartMaxRetries", Value: "5"},
			{Key: "restartTimeout", Value: "X"}, // not valid, shall fallback to 0
			// io config
			{Key: "terminal", Value: "YES"},
			{Key: "interactive", Value: "1"},
			{Key: "memory", Value: "50M"},
		},
	}
	container, err := toContainer(containerConfig)
	testutil.AssertNil(t, err)

	testutil.AssertEqual(t, 1, len(container.HostConfig.Devices))
	testutil.AssertEqual(t, "/dev/abc", container.HostConfig.Devices[0].PathOnHost)
	testutil.AssertEqual(t, "/dev/ABC", container.HostConfig.Devices[0].PathInContainer)

	testutil.AssertEqual(t, 1, len(container.HostConfig.PortMappings))
	testutil.AssertEqual(t, "0.0.0.0", container.HostConfig.PortMappings[0].HostIP)
	testutil.AssertEqual(t, uint16(80), container.HostConfig.PortMappings[0].HostPort)
	testutil.AssertEqual(t, uint16(80), container.HostConfig.PortMappings[0].HostPortEnd)
	testutil.AssertEqual(t, uint16(8888), container.HostConfig.PortMappings[0].ContainerPort)
	testutil.AssertEqual(t, "tcp", container.HostConfig.PortMappings[0].Proto)

	testutil.AssertEqual(t, 1, len(container.Mounts))
	testutil.AssertEqual(t, "/tmp", container.Mounts[0].Source)
	testutil.AssertEqual(t, "/var/tmp", container.Mounts[0].Destination)

	testutil.AssertEqual(t, []string{"ctr_host", "testhost"}, container.HostConfig.ExtraHosts)

	testutil.AssertEqual(t, []string{"arg1", "arg2"}, container.Config.Cmd)
	testutil.AssertEqual(t, []string{"DEBUG=true", "ENV1="}, container.Config.Env)
	testutil.AssertEqual(t, &ctrtypes.RestartPolicy{Type: ctrtypes.OnFailure, MaximumRetryCount: 5}, container.HostConfig.RestartPolicy)
	testutil.AssertEqual(t, &ctrtypes.IOConfig{Tty: false, OpenStdin: true}, container.IOConfig)
	testutil.AssertEqual(t, &ctrtypes.Resources{Memory: "50M"}, container.HostConfig.Resources)
}
