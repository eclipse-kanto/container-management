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

package client

// DeviceCredentials represents the device authentication ID and device password
type DeviceCredentials interface {
	GetAuthID() string
	GetPassword() string
}
type credentials struct {
	authID   string
	password string
}

func (c *credentials) GetAuthID() string {
	return c.authID
}

func (c *credentials) GetPassword() string {
	return c.password
}
