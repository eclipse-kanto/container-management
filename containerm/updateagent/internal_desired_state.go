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
	"strings"

	ctrtypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/util"

	"github.com/eclipse-kanto/update-manager/api/types"
	"github.com/pkg/errors"
)

const (
	keyImage                     = "image"
	keyTerminal                  = "terminal"
	keyInteractive               = "interactive"
	keyPrivileged                = "privileged"
	keyRestartPolicy             = "restartPolicy"
	keyRestartMaxRetries         = "restartMaxRetries"
	keyRestartTimeout            = "restartTimeout"
	keyDevice                    = "device"
	keyPort                      = "port"
	keyNetwork                   = "network"
	keyHost                      = "host"
	keyMount                     = "mount"
	keyEnv                       = "env"
	keyCmd                       = "cmd"
	keyLogDriver                 = "logDriver"
	keyLogMaxFiles               = "logMaxFiles"
	keyLogMaxSize                = "logMaxSize"
	keyLogPath                   = "logPath"
	keyLogMode                   = "logMode"
	keyLogMaxBufferSize          = "logMaxBufferSize"
	keyMemory                    = "memory"
	keyMemoryReservation         = "memoryReservation"
	keyMemorySwap                = "memorySwap"
	keyDomainName                = "domainName"
	keyHostName                  = "hostName"
	keyStatus                    = "status"
	keyFinishedAt                = "finishedAt"
	keyExitCode                  = "exitCode"
	keyCreated                   = "created"
	keyRestartCount              = "restartCount"
	keyManuallyStopped           = "manuallyStopped"
	keyStartedSuccessfullyBefore = "startedSuccessfullyBefore"

	keySystemContainers = "systemContainers"
)

type internalDesiredState struct {
	desiredState     *types.DesiredState
	systemContainers []string

	containers []*ctrtypes.Container
	baselines  map[string][]*ctrtypes.Container
}

func (ds *internalDesiredState) findComponent(name string) types.Component {
	for _, component := range ds.desiredState.Domains[0].Components {
		if component.ID == name {
			return component.Component
		}
	}
	return types.Component{}
}

// toInternalDesiredState converts incoming desired state into an internal desired state structure
func toInternalDesiredState(desiredState *types.DesiredState, domainName string) (*internalDesiredState, error) {
	if len(desiredState.Domains) != 1 {
		return nil, fmt.Errorf("one domain expected in desired state specification, but got %d", len(desiredState.Domains))
	}
	if desiredState.Domains[0].ID != domainName {
		return nil, fmt.Errorf("domain id mismatch - expecting %s, received %s", domainName, desiredState.Domains[0].ID)
	}

	containers, err := toContainers(desiredState.Domains[0].Components)
	if err != nil {
		return nil, errors.Wrap(err, "cannot convert desired state components to container configurations")
	}
	baselines, err := baselinesWithContainers(domainName+":", desiredState.Baselines, util.AsNamedMap(containers))
	if err != nil {
		return nil, errors.Wrap(err, "cannot process desired state baselines with containers")
	}
	var systemContainers []string
	for _, configPair := range desiredState.Domains[0].Config {
		if configPair.Key == keySystemContainers {
			systemContainers = strings.Split(configPair.Value, ",")
			for index, name := range systemContainers {
				systemContainers[index] = strings.TrimSpace(name)
			}
		}
	}

	return &internalDesiredState{
		desiredState:     desiredState,
		containers:       containers,
		baselines:        baselines,
		systemContainers: systemContainers,
	}, nil
}
