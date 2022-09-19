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

import "github.com/eclipse-kanto/container-management/containerm/containers/types"

type device struct {
	PathOnHost        string `json:"pathOnHost"`
	PathInContainer   string `json:"pathInContainer"`
	CgroupPermissions string `json:"cgroupPermissions,omitempty"`
}

func fromAPIDevice(apiDev types.DeviceMapping) *device {
	return &device{
		PathOnHost:        apiDev.PathOnHost,
		PathInContainer:   apiDev.PathInContainer,
		CgroupPermissions: apiDev.CgroupPermissions,
	}
}

func toAPIDevice(internalDev *device) types.DeviceMapping {
	return types.DeviceMapping{
		PathOnHost:        internalDev.PathOnHost,
		PathInContainer:   internalDev.PathInContainer,
		CgroupPermissions: internalDev.CgroupPermissions,
	}
}
