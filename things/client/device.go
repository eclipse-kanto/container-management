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

package client

import (
	"github.com/eclipse-kanto/container-management/things/api/model"
)

// Device provides an ability to gets device data
type Device interface {
	GetID() model.NamespacedID
	GetViaGateway() model.NamespacedID
	GetTenantID() string
	GetCredentials() DeviceCredentials
	GetConnection() DeviceConnection
}

type device struct {
	id          model.NamespacedID
	viaGateway  model.NamespacedID
	tenantID    string
	credentials *credentials
	connection  *deviceConnection
}

// GetID returns a device ID
func (dev *device) GetID() model.NamespacedID {
	return dev.id
}

// GetViaGateway returns a device ID via gateway
func (dev *device) GetViaGateway() model.NamespacedID {
	return dev.viaGateway
}

// GetTenantID returns a device tenant ID
func (dev *device) GetTenantID() string {
	return dev.tenantID
}

// GetCredentials returns a device credentials
func (dev *device) GetCredentials() DeviceCredentials {
	return dev.credentials
}

// GetConnection returns a device connection
func (dev *device) GetConnection() DeviceConnection {
	return dev.connection
}
