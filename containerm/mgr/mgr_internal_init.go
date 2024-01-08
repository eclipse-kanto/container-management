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

package mgr

import (
	"fmt"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/ctr"
	"github.com/eclipse-kanto/container-management/containerm/events"
	"github.com/eclipse-kanto/container-management/containerm/network"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

func newContainerMgr(metaPath string, execPath string, defaultCtrsStopTimeout time.Duration, ctrClient ctr.ContainerAPIClient, netMgr network.ContainerNetworkManager, eventsMgr events.ContainerEventsManager) (ContainerManager, error) {
	if err := util.MkDir(execPath); err != nil {
		return nil, err
	}

	if err := util.MkDir(metaPath); err != nil {
		return nil, err
	}

	locksCache := util.NewLocksCache()
	ctrRepository := containerFsRepository{metaPath: metaPath, locksCache: &locksCache}

	manager := &containerMgr{
		metaPath:               metaPath,
		execPath:               execPath,
		defaultCtrsStopTimeout: defaultCtrsStopTimeout,
		ctrClient:              ctrClient,
		netMgr:                 netMgr,
		eventsMgr:              eventsMgr,
		containers:             make(map[string]*types.Container),
		restartCtrsMgrCache:    newRestartMgrCache(),
		containerRepository:    &ctrRepository,
	}
	fmt.Println("mgr:", defaultCtrsStopTimeout.String())
	ctrClient.SetContainerExitHooks(manager.exitedAndRelease)

	return manager, nil
}

func registryInit(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
	mgrInitOpts := registryCtx.Config.([]ContainerManagerOpt)
	var (
		mgrOpts     = &mgrOpts{}
		err         error
		allServices map[string]*registry.ServiceInfo
	)
	applyOptsMgr(mgrOpts, mgrInitOpts...)

	eventsManagerService, errEvtsMgr := registryCtx.Get(registry.EventsManagerService)
	if errEvtsMgr != nil {
		return nil, errEvtsMgr
	}

	allServices, err = registryCtx.GetByType(registry.ContainerClientService)
	if err != nil {
		return nil, err
	}

	// get the desired container client local service instance
	ctrClientServiceInfo, ok := allServices[mgrOpts.containerClientServiceID]
	if ctrClientServiceInfo == nil || !ok {
		return nil, fmt.Errorf("missing required container client service with id = %s", mgrOpts.containerClientServiceID)
	}
	ctrClientService, err := ctrClientServiceInfo.Instance()
	if err != nil {
		return nil, fmt.Errorf("the required container client service with id = %s has initialization errors %v", mgrOpts.containerClientServiceID, err)
	}

	allServices, err = registryCtx.GetByType(registry.NetworkManagerService)
	if err != nil {
		return nil, err
	}
	// get the desired network manager local service instance
	netMgrServiceInfo, ok := allServices[mgrOpts.networkManagerServiceID]
	if netMgrServiceInfo == nil || !ok {
		return nil, fmt.Errorf("missing required network manager service with id = %s", mgrOpts.networkManagerServiceID)
	}
	netMgrService, err := netMgrServiceInfo.Instance()
	if err != nil {
		return nil, fmt.Errorf("the required network manager service with id = %s has initialization errors %v", mgrOpts.networkManagerServiceID, err)
	}

	//initialize the manager local service
	return newContainerMgr(mgrOpts.metaPath, mgrOpts.rootExec, mgrOpts.defaultCtrsStopTimeout, ctrClientService.(ctr.ContainerAPIClient), netMgrService.(network.ContainerNetworkManager), eventsManagerService.(events.ContainerEventsManager))

}
