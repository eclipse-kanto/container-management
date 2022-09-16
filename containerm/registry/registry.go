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

package registry

import (
	"fmt"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"google.golang.org/grpc"
)

// Type is the type of the service in the registry
type Type string

func (t Type) String() string { return string(t) }

const (
	// EventsManagerService implements THE events manager service
	EventsManagerService Type = "container-management.service.events.manager.v1"
	// NetworkManagerService implements THE network manager service
	NetworkManagerService Type = "container-management.service.net.manager.v1"
	// ContainerClientService implements THE containers API client service
	ContainerClientService Type = "container-management.service.ctrs.client.v1"
	// ContainerManagerService implements THE container manager service
	ContainerManagerService Type = "container-management.service.ctrs.manager.v1"
	// SystemInfoService implements THE system information service
	SystemInfoService Type = "container-management.service.system.info.v1"
	// ThingsContainerManagerService implements THE container management via the IoT Rollouts and IoT Things services
	ThingsContainerManagerService Type = "container-management.service.things.ctrs.manager.v1"
	// GRPCService implements a grpc service
	GRPCService Type = "container-management.service.grpc.v1"
	// GRPCServer implements a grpc server
	GRPCServer Type = "container-management.server.grpc.v1"
)

// Registration holds service's information that will be added to the registry
type Registration struct {
	ID       string
	Type     Type
	InitFunc func(registryCtx *ServiceRegistryContext) (interface{}, error)
	//TODO maybe add Requires for required services
}

// Init the registered plugin
func (r *Registration) Init(rc *ServiceRegistryContext) *ServiceInfo {
	log.Debug("initialization of service ID = %s has been requested", r.ID)
	p, err := r.InitFunc(rc)
	if err != nil {
		log.ErrorErr(err, "initialization of service ID = %s has failed - no instance will be added to the registry!", r.ID)
	}
	return &ServiceInfo{
		Registration: r,
		instance:     p,
		err:          err,
	}
}

// GrpcService allows GRPC services to be registered with the underlying server
type GrpcService interface {
	Register(*grpc.Server) error
}

// GrpcServer manages start and stop opeerations of the gRPC server
type GrpcServer interface {
	Start() error
	Stop() error
}

//internal struct to hold all registrations with a synced access
var register = struct {
	sync.RWMutex
	r []*Registration
}{}

// Register allows services to register
func Register(r *Registration) {
	register.Lock()
	defer register.Unlock()
	if r.Type == "" {
		panic(fmt.Errorf("the service has no netType"))
	}
	if r.ID == "" {
		panic(fmt.Errorf("the service has no ID"))
	}
	log.Debug("added service registration for service with ID = %s ", r.ID)
	register.r = append(register.r, r)
}

// RegistrationsMap returns map with all service registrations
func RegistrationsMap() map[Type][]*Registration {
	register.RLock()
	defer register.RUnlock()
	if register.r == nil {
		return nil
	}
	currentRegistrationsMap := make(map[Type][]*Registration)
	for _, reg := range register.r {
		currentRegistrationsMap[reg.Type] = append(currentRegistrationsMap[reg.Type], reg)
	}
	return currentRegistrationsMap
}
