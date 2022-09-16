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
	"fmt"

	"github.com/eclipse-kanto/container-management/things/client/protocol"
)

const (
	topicPublishEvent     = "e/%s/%s"
	topicPublishTelemetry = "t/%s/%s"
)

// DeviceConnection enebles the connection to the device
type DeviceConnection interface {
	SendTelemetry(envelope protocol.Envelope) error
	SendEvent(envelope protocol.Envelope) error
}

type deviceConnection struct {
	//configs connection
	client *Client
}

func (conn *deviceConnection) SendTelemetry(envelope protocol.Envelope) error {
	return conn.client.publish(fmt.Sprintf(topicPublishTelemetry, conn.client.device.tenantID, conn.client.device.id), envelope, 1, false)
}

func (conn deviceConnection) SendEvent(envelope protocol.Envelope) error {
	return conn.client.publish(fmt.Sprintf(topicPublishEvent, conn.client.device.tenantID, conn.client.device.id), envelope, 1, false)
}
