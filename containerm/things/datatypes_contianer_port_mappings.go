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

type portMapping struct {
	Proto         string `json:"proto,omitempty"`
	HostPort      uint16 `json:"hostPort"`
	HostPortEnd   uint16 `json:"hostPortEnd,omitempty"`
	ContainerPort uint16 `json:"containerPort"`
	HostIP        string `json:"hostIP,omitempty"`
}

func toAPIPortMapping(internalPM *portMapping) types.PortMapping {
	return types.PortMapping{
		Proto:         internalPM.Proto,
		ContainerPort: internalPM.ContainerPort,
		HostIP:        internalPM.HostIP,
		HostPort:      internalPM.HostPort,
		HostPortEnd:   internalPM.HostPortEnd,
	}
}
func fromAPIPortMapping(apiPM types.PortMapping) *portMapping {
	return &portMapping{
		Proto:         apiPM.Proto,
		ContainerPort: apiPM.ContainerPort,
		HostIP:        apiPM.HostIP,
		HostPort:      apiPM.HostPort,
		HostPortEnd:   apiPM.HostPortEnd,
	}
}
