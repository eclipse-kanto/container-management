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
	"context"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/events"
	"github.com/eclipse-kanto/container-management/containerm/mgr"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/eclipse-kanto/container-management/rollouts/api/features"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
)

type softwareUpdatable struct {
	rootThing             model.Thing
	status                *features.SoftwareUpdatableStatus
	mgr                   mgr.ContainerManager
	eventsMgr             events.ContainerEventsManager
	cancelEventsHandler   context.CancelFunc
	depsLock              sync.Mutex
	processOperationsLock sync.Mutex
}

func (su *softwareUpdatable) SoftwareModuleType() string {
	return su.status.SoftwareModuleType
}

func (su *softwareUpdatable) LastOperation() *datatypes.OperationStatus {
	return su.status.LastOperation
}

func (su *softwareUpdatable) LastFailedOperation() *datatypes.OperationStatus {
	return su.status.LastFailedOperation
}

func (su *softwareUpdatable) InstalledDependencies() map[string]*datatypes.DependencyDescription {
	return su.status.InstalledDependencies
}

func (su *softwareUpdatable) ContextDependencies() map[string]*datatypes.DependencyDescription {
	return su.status.ContextDependencies
}

// Downloads and installs a given list of software modules
func (su *softwareUpdatable) Install(updateAction datatypes.UpdateAction) error {
	log.Debug("will perform installation...")

	if err := validateSoftwareUpdateAction(updateAction); err != nil {
		return client.NewMessagesParameterInvalidError(err.Error())
	}
	go su.processUpdateAction(updateAction)
	return nil
}

func (su *softwareUpdatable) Remove(dsAction datatypes.RemoveAction) error {
	if len(dsAction.Software) == 0 {
		return client.NewMessagesParameterInvalidError("there are no DependencyDescriptions to be removed")
	}
	go su.processRemoveAction(dsAction)
	return nil
}

func (su *softwareUpdatable) Download(dsAction datatypes.UpdateAction) error {
	// not supported
	return nil
}

func (su *softwareUpdatable) Cancel(dsAction datatypes.UpdateAction) error {
	// not supported
	return nil
}
func (su *softwareUpdatable) CancelRemove(dsAction datatypes.RemoveAction) error {
	// not supported
	return nil
}
