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
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/sysinfo"
)

func init() {
	registry.Register(&registry.Registration{
		ID:   SystemInfoServiceID,
		Type: registry.GRPCService,
		InitFunc: func(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
			sysInfomgrService, err := registryCtx.Get(registry.SystemInfoService)
			if err != nil {
				return nil, err
			}
			return &systemInfo{sysInfoMgr: sysInfomgrService.(sysinfo.SystemInfoManager)}, nil
		},
	})
}
