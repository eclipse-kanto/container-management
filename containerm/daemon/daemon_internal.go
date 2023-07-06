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

	"github.com/eclipse-kanto/container-management/containerm/deployment"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/things"

	"github.com/eclipse-kanto/update-manager/api"
)

func (d *daemon) start() error {
	log.Debug("starting daemon instance")

	if err := d.loadContainerManagersStoredInfo(); err != nil {
		log.ErrorErr(err, "could not load and restore persistent data for the Container Manager Services")
		return err
	}

	if d.config.DeploymentManagerConfig.DeploymentEnable {
		if err := d.deploy(); err != nil {
			log.ErrorErr(err, "could not perform initial deploy / update for Deployment Manager Services")
			return err
		}
	}

	if d.config.ThingsConfig.ThingsEnable {
		err := d.startThingsManagers()
		if err != nil {
			log.ErrorErr(err, "could not start the Things Container Manager Services")
		}
	}

	if d.config.UpdateAgentConfig.UpdateAgentEnable {
		log.Debug("Containers Update Agent is enabled.")
		err := d.startUpdateAgents()
		if err != nil {
			log.ErrorErr(err, "could not start the Containers Update Agent Services")
		}
	} else {
		log.Debug("Containers Update Agent is not enabled.")
	}

	return d.startGrpcServers()

}

func (d *daemon) stop() {
	log.Debug("stopping of the GW CM daemon is requested and started")
	log.Debug("stopping gRPC server ")
	d.stopGrpcServers()

	if d.config.DeploymentManagerConfig.DeploymentEnable {
		log.Debug("stopping deployment managers local services")
		d.stopDeploymentManagers()
	}

	log.Debug("stopping management local services")
	d.stopContainerManagers()

	if d.config.UpdateAgentConfig.UpdateAgentEnable {
		log.Debug("stopping Containers Update Agents services")
		d.stopUpdateAgents()
	}

	if d.config.ThingsConfig.ThingsEnable {
		log.Debug("stopping Things Container Manager service")
		d.stopThingsManagers()
	}

	log.Debug("stopping of the GW CM daemon finished")
}

func (d *daemon) startGrpcServers() error {
	log.Debug("starting gRPC servers ")
	grpcServerInfos := d.serviceInfoSet.GetAll(registry.GRPCServer)
	var (
		instance interface{}
		err      error
	)

	log.Debug("there are %d gRPC servers to be started", len(grpcServerInfos))
	for _, servInfo := range grpcServerInfos {
		log.Debug("will try to start gRPC server local service with ID = %s", servInfo.Registration.ID)
		instance, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get gRPC server instance - local service ID = %s ", servInfo.Registration.ID)
		} else {
			err = instance.(registry.GrpcServer).Start()
			if err != nil {
				log.ErrorErr(err, "could not start gRPC server with service ID = %s ", servInfo.Registration.ID)
			} else {
				log.Debug("successfully started gRPC server service with service ID = %s ", servInfo.Registration.ID)
			}
		}
	}
	return err
}

func (d *daemon) stopGrpcServers() {
	log.Debug("will stop gRPC servers")
	grpcServerInfos := d.serviceInfoSet.GetAll(registry.GRPCServer)
	var (
		instance interface{}
		err      error
	)

	for _, servInfo := range grpcServerInfos {
		instance, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get gRPC server instance for service ID = %s", servInfo.Registration.ID)
		} else {
			err = instance.(registry.GrpcServer).Stop()
			if err != nil {
				log.ErrorErr(err, "could not stop gRPC server for service ID = %s ", servInfo.Registration.ID)
			}
		}
	}
}
func (d *daemon) loadContainerManagersStoredInfo() error {
	log.Debug("will load and restore stored data for container management local services")
	ctrMrgServices := d.serviceInfoSet.GetAll(registry.ContainerManagerService)
	var (
		instance interface{}
		err      error
		ctrMgr   mgr.ContainerManager
	)
	log.Debug("there are %d container management services to load the data for", len(ctrMrgServices))
	for _, servInfo := range ctrMrgServices {
		instance, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get container management service instance for service ID = %s", servInfo.Registration.ID)
		} else {
			ctrMgr = instance.(mgr.ContainerManager)
			ctx := context.Background()
			if err = ctrMgr.Load(ctx); err != nil {
				log.ErrorErr(err, "could not load stored data for container management service for service ID = %s", servInfo.Registration.ID)
				return err
			}

			if err = ctrMgr.Restore(ctx); err != nil {
				log.ErrorErr(err, "could not restore stored containers for container management service for service ID = %s", servInfo.Registration.ID)
				return err
			}
		}
	}
	return nil
}

func (d *daemon) deploy() error {
	log.Debug("will perform initial deploy / update for deployment managers local services")
	deploymentMgrServices := d.serviceInfoSet.GetAll(registry.DeploymentManagerService)
	var (
		instance interface{}
		err      error
		dMgr     deployment.Manager
	)
	log.Debug("there are %d deployment manager services to load the data for", len(deploymentMgrServices))
	for _, servInfo := range deploymentMgrServices {
		instance, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get deployment manager service instance for service ID = %s", servInfo.Registration.ID)
		} else {
			dMgr = instance.(deployment.Manager)
			ctx := context.Background()
			if err = dMgr.Deploy(ctx); err != nil {
				log.ErrorErr(err, "could not perform initial deploy / update for deployment manager service for service ID = %s", servInfo.Registration.ID)
				return err
			}
		}
	}
	return nil
}

func (d *daemon) startThingsManagers() error {
	log.Debug("starting Things Container Manager services ")
	grpcServerInfos := d.serviceInfoSet.GetAll(registry.ThingsContainerManagerService)
	var (
		instance interface{}
		err      error
	)

	log.Debug("there are %d Things Container Manager services to be started", len(grpcServerInfos))
	for _, servInfo := range grpcServerInfos {
		log.Debug("will try to start Things Container Manager service local service with ID = %s", servInfo.Registration.ID)
		instance, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get Things Container Manager service instance - local service ID = %s ", servInfo.Registration.ID)
		} else {
			err = instance.(things.ContainerThingsManager).Connect()
			if err != nil {
				log.ErrorErr(err, "could not start Things Container Manager service with service ID = %s ", servInfo.Registration.ID)
			} else {
				log.Debug("successfully started Things Container Manager service with service ID = %s ", servInfo.Registration.ID)
			}
		}
	}
	return err
}

func (d *daemon) stopThingsManagers() {
	log.Debug("will stop Things Container Manager services")
	grpcServerInfos := d.serviceInfoSet.GetAll(registry.ThingsContainerManagerService)
	var (
		instance interface{}
		err      error
	)

	for _, servInfo := range grpcServerInfos {
		instance, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get Things Container Manager service instance for service ID = %s", servInfo.Registration.ID)
		} else {
			instance.(things.ContainerThingsManager).Disconnect()
			log.Debug("successfully stopped Things Container Manager service with service ID = %s ", servInfo.Registration.ID)
		}
	}
}

func (d *daemon) startUpdateAgents() error {
	log.Debug("starting Update Agent services ")
	updateAgentInfos := d.serviceInfoSet.GetAll(registry.UpdateAgentService)
	var instance interface{}
	var err error

	log.Debug("there are %d Update Agent services to be started", len(updateAgentInfos))
	for _, updateAgentInfo := range updateAgentInfos {
		log.Debug("will try to start Update Agent service instance with service ID = %s", updateAgentInfo.Registration.ID)
		instance, err = updateAgentInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get Update Agent service instance with service ID = %s", updateAgentInfo.Registration.ID)
		} else {
			err = instance.(api.UpdateAgent).Start(context.Background())
			if err != nil {
				log.ErrorErr(err, "could not start Update Agent service instance with service ID = %s", updateAgentInfo.Registration.ID)
			} else {
				log.Debug("successfully started Update Agent service instance with service ID = %s ", updateAgentInfo.Registration.ID)
			}
		}
	}
	return err
}

func (d *daemon) stopUpdateAgents() {
	log.Debug("will stop Update Agent services")
	updateAgentInfos := d.serviceInfoSet.GetAll(registry.UpdateAgentService)

	for _, updateAgentInfo := range updateAgentInfos {
		instance, err := updateAgentInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get Update Agent service instance with service ID = %s", updateAgentInfo.Registration.ID)
		} else {
			err = instance.(api.UpdateAgent).Stop()
			if err != nil {
				log.ErrorErr(err, "could not stop gracefully Update Agent service instance with service ID = %s", updateAgentInfo.Registration.ID)
			} else {
				log.Debug("successfully stopped Update Agent service with service ID = %s ", updateAgentInfo.Registration.ID)
			}
		}
	}
}

func (d *daemon) stopContainerManagers() {
	log.Debug("will stop container management local services")
	ctrMrgServices := d.serviceInfoSet.GetAll(registry.ContainerManagerService)
	var (
		instance interface{}
		err      error
	)
	log.Debug("there are %d container management services to be stopped", len(ctrMrgServices))
	for _, servInfo := range ctrMrgServices {
		instance, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get container management service instance for service ID = %s", servInfo.Registration.ID)
		} else {
			ctx := context.Background()
			err = instance.(mgr.ContainerManager).Dispose(ctx)
			if err != nil {
				log.ErrorErr(err, "could not stop container management service for service ID = %s", servInfo.Registration.ID)
			}
		}
	}
}

func (d *daemon) stopDeploymentManagers() {
	log.Debug("will stop deployment managers local services")
	deployMgrServices := d.serviceInfoSet.GetAll(registry.DeploymentManagerService)
	var (
		instance interface{}
		err      error
	)
	log.Debug("there are %d deployment manager services to be stopped", len(deployMgrServices))
	for _, servInfo := range deployMgrServices {
		instance, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get deployment manager service instance for service ID = %s", servInfo.Registration.ID)
		} else {
			ctx := context.Background()
			err = instance.(deployment.Manager).Dispose(ctx)
			if err != nil {
				log.ErrorErr(err, "could not stop deployment manager service for service ID = %s", servInfo.Registration.ID)
			}
		}
	}
}
