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
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/things"
)

func (d *daemon) start() error {
	log.Debug("starting daemon instance")

	if err := d.loadContainerManagersStoredInfo(); err != nil {
		log.ErrorErr(err, "could not load and restore persistent data for the Container Manager Services")
		return err
	}

	if d.config.ThingsConfig.ThingsEnable {
		err := d.startThingsManagers()
		if err != nil {
			log.ErrorErr(err, "could not start the Things Container Manager Services")
		}
	}

	return d.startGrpcServers()

}

func (d *daemon) stop() {
	log.Debug("stopping of the GW CM daemon is requested and started")
	log.Debug("stopping gRPC server ")
	d.stopGrpcServers()

	log.Debug("stopping management local services")
	d.stopContainerManagers()

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
		instnace interface{}
		err      error
	)

	log.Debug("there are %d gRPC servers to be started", len(grpcServerInfos))
	for _, servInfo := range grpcServerInfos {
		log.Debug("will try to start gRPC server local service with ID = %s", servInfo.Registration.ID)
		instnace, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get gRPC server instance - local service ID = %s ", servInfo.Registration.ID)
		} else {
			err = instnace.(registry.GrpcServer).Start()
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
		instnace interface{}
		err      error
	)

	for _, servInfo := range grpcServerInfos {
		instnace, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get gRPC server instance for service ID = %s", servInfo.Registration.ID)
		} else {
			err = instnace.(registry.GrpcServer).Stop()
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
		instnace interface{}
		err      error
		ctrMgr   mgr.ContainerManager
	)
	log.Debug("there are %d container management services to load the data for", len(ctrMrgServices))
	for _, servInfo := range ctrMrgServices {
		instnace, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get container management service instance for service ID = %s", servInfo.Registration.ID)
		} else {
			ctrMgr = instnace.(mgr.ContainerManager)
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

func (d *daemon) startThingsManagers() error {
	log.Debug("starting Things Container Manager services ")
	grpcServerInfos := d.serviceInfoSet.GetAll(registry.ThingsContainerManagerService)
	var (
		instnace interface{}
		err      error
	)

	log.Debug("there are %d Things Container Manager services to be started", len(grpcServerInfos))
	for _, servInfo := range grpcServerInfos {
		log.Debug("will try to start Things Container Manager service local service with ID = %s", servInfo.Registration.ID)
		instnace, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get Things Container Manager service instance - local service ID = %s ", servInfo.Registration.ID)
		} else {
			err = instnace.(things.ContainerThingsManager).Connect()
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
		instnace interface{}
		err      error
	)

	for _, servInfo := range grpcServerInfos {
		instnace, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get Things Container Manager service instance for service ID = %s", servInfo.Registration.ID)
		} else {
			instnace.(things.ContainerThingsManager).Disconnect()
			log.Debug("successfully stopped Things Container Manager service with service ID = %s ", servInfo.Registration.ID)
		}
	}
}

func (d *daemon) stopContainerManagers() {
	log.Debug("will stop container management local services")
	ctrMrgServices := d.serviceInfoSet.GetAll(registry.ContainerManagerService)
	var (
		instnace interface{}
		err      error
	)
	log.Debug("there are %d container management services to be stopped", len(ctrMrgServices))
	for _, servInfo := range ctrMrgServices {
		instnace, err = servInfo.Instance()
		if err != nil {
			log.ErrorErr(err, "could not get container management service instance for service ID = %s", servInfo.Registration.ID)
		} else {
			ctx := context.Background()
			err = instnace.(mgr.ContainerManager).Dispose(ctx)
			if err != nil {
				log.ErrorErr(err, "could not stop container management service for service ID = %s", servInfo.Registration.ID)
			}
		}
	}
}
