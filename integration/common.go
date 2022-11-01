// Copyright (c) 2022 Contributors to the Eclipse Foundation
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

package integration

import (
	"fmt"
	"strings"
	"time"

	"github.com/eclipse/ditto-clients-golang"
	"github.com/eclipse/ditto-clients-golang/model"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type containerManagementSuite struct {
	suite.Suite
	mqttClient          mqtt.Client
	dittoClient         *ditto.Client
	cfg                 *testConfig
	containerID         string
	containerURL        string
	containerFactoryURL string
	topicModify         string
}

type testConfig struct {
	Broker                   string `def:"tcp://localhost:1883"`
	MqttQuiesceMs            int    `def:"500"`
	DittoAddress             string
	DittoUser                string `def:"ditto"`
	DittoPassword            string `def:"ditto"`
	EventTimeoutMs           int    `def:"30000"`
	MqttAcknowledgeTimeoutMs int    `def:"3000"`
}

type edgeConfig struct {
	DeviceID string `json:"deviceId"`
}

type requestBody struct {
	param  string
	params map[string]interface{}
}

type jsonFeature struct {
	Definition []string               `json:"definition,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

type config struct {
	DomainName string `json:"domainName,omitempty"`
}

const (
	containerFactoryID = "ContainerFactory"
)

func (suite *containerManagementSuite) newTestConnection() {
	cfg := &testConfig{}

	suite.T().Log(getConfigHelp(*cfg))

	if err := initConfigFromEnv(cfg); err != nil {
		suite.T().Skip(err)
	}

	suite.T().Logf("test config: %+v", *cfg)

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(uuid.New().String()).
		SetKeepAlive(30 * time.Second).
		SetCleanSession(true).
		SetAutoReconnect(true)

	mqttClient := mqtt.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		require.NoError(suite.T(), token.Error(), "connect to MQTT broker")
	}

	edgeDeviceCfg, err := getEdgeDeviceConfig(mqttClient)
	if err != nil {
		mqttClient.Disconnect(uint(cfg.MqttQuiesceMs))
		require.NoError(suite.T(), err, "get edge device config")
	}

	dittoClient, err := ditto.NewClientMQTT(mqttClient, ditto.NewConfiguration())
	if err == nil {
		err = dittoClient.Connect()
	}

	if err != nil {
		mqttClient.Disconnect(uint(cfg.MqttQuiesceMs))
		require.NoError(suite.T(), err, "initialize ditto client")
	}

	suite.cfg = cfg
	suite.dittoClient = dittoClient
	suite.mqttClient = mqttClient
	suite.containerID = edgeDeviceCfg.DeviceID + ":edge:containers"
	suite.containerURL = fmt.Sprintf("%s/api/2/things/%s", strings.TrimSuffix(cfg.DittoAddress, "/"), suite.containerID)
	suite.containerFactoryURL = fmt.Sprintf("%s/features/%s", suite.containerURL, containerFactoryID)
	ns := model.NewNamespacedIDFrom(suite.containerID)
	suite.topicModify = fmt.Sprintf("%s/%s/things/twin/events/modified", ns.Namespace, ns.Name)
}

func (suite *containerManagementSuite) disconnect() {
	suite.dittoClient.Disconnect()
	suite.mqttClient.Disconnect(uint(suite.cfg.MqttQuiesceMs))
}

func getContainerID(s string) string {
	result := strings.Split(s, "/")
	return result[2]
}
