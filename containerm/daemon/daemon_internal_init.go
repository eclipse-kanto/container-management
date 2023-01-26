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

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	_ "github.com/eclipse-kanto/container-management/containerm/server"
	_ "github.com/eclipse-kanto/container-management/containerm/services"
)

func (d *daemon) init() {
	daemonConfig := d.config
	log.Debug("will start GW CM initialization")

	registrationsMap := registry.RegistrationsMap()

	log.Debug("the current registered services ready for initialization are %+v", registrationsMap)

	var ctx = context.TODO()

	//init events manager services
	initService(ctx, d, registrationsMap, registry.EventsManagerService)

	//init system info manager services
	initService(ctx, d, registrationsMap, registry.SystemInfoService)

	//init container client services
	initService(ctx, d, registrationsMap, registry.ContainerClientService)

	//init network manager services
	initService(ctx, d, registrationsMap, registry.NetworkManagerService)

	//init container manager service
	initService(ctx, d, registrationsMap, registry.ContainerManagerService)

	//init Things container manager service
	if daemonConfig.ThingsConfig.ThingsEnable {
		initService(ctx, d, registrationsMap, registry.ThingsContainerManagerService)
	} else {
		log.Info("Things support is disabled - no Things Container Manager Services will be registered. If you would like to enable Things support, please, reconfigure things-enable to true")
	}

	//init deployment manager service
	initService(ctx, d, registrationsMap, registry.DeploymentManagerService)

	//init grpc services
	initService(ctx, d, registrationsMap, registry.GRPCService)

	//init grpc server services
	initService(ctx, d, registrationsMap, registry.GRPCServer)

}

func initService(ctx context.Context, d *daemon, registrationsMap map[registry.Type][]*registry.Registration, regType registry.Type) {
	var config interface{}
	log.Debug("will initialize all %s services", regType)
	serviceRegs, ok := registrationsMap[regType]
	if ok {
		switch regType {
		case registry.ContainerClientService:
			config = extractCtrClientConfigOptions(d.config)
			break
		case registry.NetworkManagerService:
			config = extractNetManagerConfigOptions(d.config)
			break
		case registry.ContainerManagerService:
			config = extractContainerManagerOptions(d.config)
			break
		case registry.ThingsContainerManagerService:
			config = extractThingsOptions(d.config)
			break
		case registry.GRPCServer:
			config = extractGrpcOptions(d.config)
			break
		case registry.DeploymentManagerService:
			config = extractDeploymentMgrOptions(d.config)
			break
		default:
			config = nil
		}
		d.initServices(ctx, serviceRegs, config)
	} else {
		log.Debug("there are no %s services registered", regType)
	}
}

func (d *daemon) initServices(ctx context.Context, registrations []*registry.Registration, config interface{}) {
	var (
		regCtx   *registry.ServiceRegistryContext
		servInfo *registry.ServiceInfo
	)
	for _, reg := range registrations {
		regCtx = registry.NewContext(ctx, config, reg, d.serviceInfoSet)

		log.Debug("will initialize service instance with ID = %s with context %v", reg.ID, regCtx)
		servInfo = reg.Init(regCtx)
		log.Debug("successfully initialized service instance with ID = %s ", reg.ID)
		d.serviceInfoSet.Add(servInfo)
		log.Debug("successfully added service instance with ID = %s to the local service registry", reg.ID)
	}
}
