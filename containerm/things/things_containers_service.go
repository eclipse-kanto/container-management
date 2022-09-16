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

package things

import (
	"context"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/events"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/things/client"
)

const (
	// ContainerThingsManagerServiceLocalID is the ID of the local container manager service of things
	ContainerThingsManagerServiceLocalID = "container-management.service.local.v1.service-things-container-manager"
	// thingsBackupFileNameTemplate is the filename template for backup of things
	thingsBackupFileNameTemplate = "sup_back_%s.json"
)

func init() {
	registry.Register(&registry.Registration{
		ID:       ContainerThingsManagerServiceLocalID,
		Type:     registry.ThingsContainerManagerService,
		InitFunc: registryInit,
	})
}

type managedFeature interface {
	register(ctx context.Context) error
	dispose()
}

// ContainerThingsManager interface declares connect and disconnect functions
type ContainerThingsManager interface {
	Connect() error
	Disconnect()
}

type containerThingsMgr struct {
	enabledFeatureIds []string
	storageRoot       string
	mgr               mgr.ContainerManager
	eventsMgr         events.ContainerEventsManager
	thingsClient      *client.Client

	containerThingID string
	managedFeatures  map[string]managedFeature
	initMutex        sync.Mutex
}

func (tMgr *containerThingsMgr) Connect() error {

	if err := tMgr.thingsClient.Connect(); err != nil {
		return err
	}
	return nil
}

func (tMgr *containerThingsMgr) Disconnect() {
	tMgr.initMutex.Lock()
	defer tMgr.initMutex.Unlock()

	tMgr.disposeFeatures()
	tMgr.thingsClient.Disconnect()
}
