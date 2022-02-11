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

package services

import (
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/registry"
)

func init() {
	registry.Register(&registry.Registration{
		ID:   ContainersServiceID,
		Type: registry.GRPCService,
		InitFunc: func(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
			mgrService, err := registryCtx.Get(registry.ContainerManagerService)
			if err != nil {
				return nil, err
			}
			return &containers{mgr: mgrService.(mgr.ContainerManager)}, nil
		},
	})
}
