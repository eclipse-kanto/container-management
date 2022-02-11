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

package features

import "github.com/eclipse-kanto/container-management/rollouts/api/datatypes"

// SoftwareUpdatableStatus provides the status of a Ditto feature implementing the SoftwareUpdatable v2 Vorto model
type SoftwareUpdatableStatus struct {
	SoftwareModuleType    string                                      `json:"softwareModuleType"`
	LastOperation         *datatypes.OperationStatus                  `json:"lastOperation,omitempty"`
	LastFailedOperation   *datatypes.OperationStatus                  `json:"lastFailedOperation,omitempty"`
	InstalledDependencies map[string]*datatypes.DependencyDescription `json:"installedDependencies,omitempty"`
	ContextDependencies   map[string]*datatypes.DependencyDescription `json:"contextDependencies,omitempty"`
}

// SoftwareUpdatable provides an API for implementing the SoftwareUpdatable v2 Vorto model
type SoftwareUpdatable interface {

	// Get the status module type
	SoftwareModuleType() string

	// Get the status last operation
	LastOperation() *datatypes.OperationStatus

	// Get the status failed operation
	LastFailedOperation() *datatypes.OperationStatus

	// Get the status installed dependencies
	InstalledDependencies() map[string]*datatypes.DependencyDescription

	// Get the status context dependencies
	ContextDependencies() map[string]*datatypes.DependencyDescription

	// Downloads and installs a given list of software modules
	Install(dsAction datatypes.UpdateAction) error

	// Downloads (without installing) a given list of software modules
	Download(dsAction datatypes.UpdateAction) error

	// Try to cancel a running installation
	Cancel(dsAction datatypes.UpdateAction) error

	// Remove an installed software.
	Remove(dsAction datatypes.RemoveAction) error

	// Try to cancel a remove operation
	CancelRemove(dsAction datatypes.RemoveAction) error
}
