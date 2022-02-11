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

import (
	"fmt"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	honoCommandTopicSuffixErrorsResponse = "errors-response"
	mqttTopicSubscribeErrorsBase         = mqttTopicSubscribeEventsBase + honoCommandTopicSuffixErrorsResponse
)

func (client *Client) handleTwinErrors(mqttClient MQTT.Client, message MQTT.Message) {
	fmt.Println("An error response has been pushed from the MQTT services endpoint:")
	fmt.Printf("Errors TOPIC: %s\n", message.Topic())
	fmt.Printf("Errors MSG: %s\n", message.Payload())
}
