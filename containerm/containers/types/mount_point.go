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

package types

const (
	// RPrivatePropagationMode represents mount propagation rprivate.
	RPrivatePropagationMode = "rprivate"
	// PrivatePropagationMode represents mount propagation private.
	PrivatePropagationMode = "private"
	// RSharedPropagationMode represents mount propagation rshared.
	RSharedPropagationMode = "rshared"
	// SharedPropagationMode represents mount propagation shared.
	SharedPropagationMode = "shared"
	// RSlavePropagationMode represents mount propagation rslave.
	RSlavePropagationMode = "rslave"
	// SlavePropagationMode represents mount propagation slave.
	SlavePropagationMode = "slave"
)

// MountPoint specifies a mount point from the host to the container
type MountPoint struct {
	Source          string `json:"source,omitempty"`
	Destination     string `json:"destination,omitempty"`
	PropagationMode string `json:"propagation_mode,omitempty"`
	Data            string `json:"data,omitempty"`
}
