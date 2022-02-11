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

package things

import (
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
)

type propagationMode string

const (
	rprivate propagationMode = "RPRIVATE"
	private  propagationMode = "PRIVATE"
	rshared  propagationMode = "RSHARED"
	shared   propagationMode = "SHARED"
	rslave   propagationMode = "RSLAVE"
	slave    propagationMode = "SLAVE"
)

type mountPoint struct {
	Source          string          `json:"source"`
	Destination     string          `json:"destination"`
	PropagationMode propagationMode `json:"propagationMode,omitempty"`
}

func toAPIMountPoint(internalMP *mountPoint) types.MountPoint {
	return types.MountPoint{
		Destination:     internalMP.Destination,
		Source:          internalMP.Source,
		PropagationMode: toAPIPRMode(internalMP.PropagationMode),
	}
}
func fromAPIMountPoint(apiMP types.MountPoint) *mountPoint {
	return &mountPoint{
		Destination:     apiMP.Destination,
		Source:          apiMP.Source,
		PropagationMode: fromAPIPRMode(apiMP.PropagationMode),
	}
}

func fromAPIPRMode(prMode string) propagationMode {
	switch prMode {
	case types.RPrivatePropagationMode:
		return rprivate
	case types.PrivatePropagationMode:
		return private
	case types.RSharedPropagationMode:
		return rshared
	case types.SharedPropagationMode:
		return shared
	case types.RSlavePropagationMode:
		return rslave
	case types.SlavePropagationMode:
		return slave
	default:
		return rprivate
	}
}

func toAPIPRMode(prMode propagationMode) string {
	switch prMode {
	case rprivate:
		return types.RPrivatePropagationMode
	case private:
		return types.PrivatePropagationMode
	case rshared:
		return types.RSharedPropagationMode
	case shared:
		return types.SharedPropagationMode
	case rslave:
		return types.RSlavePropagationMode
	case slave:
		return types.SlavePropagationMode
	default:
		return types.RPrivatePropagationMode
	}
}
