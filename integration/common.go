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
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type containerManagementSuite struct {
	suite.Suite
	mqttClient           MQTT.Client
	dittoClient          *ditto.Client
	cfg                  *testConfig
	ctrThingID           string
	ctrThingURL          string
	ctrFactoryFeatureURL string
	topicCreated         string
	topicModify          string
	topicDeleted         string
}

type testConfig struct {
	Broker                   string `def:"tcp://localhost:1883"`
	MqttQuiesceMs            int    `def:"500"`
	DigitalTwinAPIAddress    string
	DigitalTwinAPIUser       string `def:"ditto"`
	DigitalTwinAPIPassword   string `def:"ditto"`
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

type containerFeature struct {
	Definition []string               `json:"definition,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

const (
	ctrFactoryFeatureID = "ContainerFactory"
)

func (suite *containerManagementSuite) setup() {
	cfg := &testConfig{}

	suite.T().Log(getConfigHelp(*cfg))

	if err := initConfigFromEnv(cfg); err != nil {
		suite.T().Skip(err)
	}

	suite.T().Logf("test config: %+v", *cfg)

	opts := MQTT.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(uuid.New().String()).
		SetKeepAlive(30 * time.Second).
		SetCleanSession(true).
		SetAutoReconnect(true)

	mqttClient := MQTT.NewClient(opts)

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
	suite.ctrThingID = edgeDeviceCfg.DeviceID + ":edge:containers"
	suite.ctrThingURL = fmt.Sprintf("%s/api/2/things/%s", strings.TrimSuffix(cfg.DigitalTwinAPIAddress, "/"), suite.ctrThingID)
	suite.ctrFactoryFeatureURL = fmt.Sprintf("%s/features/%s", suite.ctrThingURL, ctrFactoryFeatureID)
	ctrThingID := model.NewNamespacedIDFrom(suite.ctrThingID)
	suite.topicCreated = fmt.Sprintf("%s/%s/things/twin/events/created", ctrThingID.Namespace, ctrThingID.Name)
	suite.topicModify = fmt.Sprintf("%s/%s/things/twin/events/modified", ctrThingID.Namespace, ctrThingID.Name)
	suite.topicDeleted = fmt.Sprintf("%s/%s/things/twin/events/deleted", ctrThingID.Namespace, ctrThingID.Name)
}

func (suite *containerManagementSuite) tearDown() {
	suite.dittoClient.Disconnect()
	suite.mqttClient.Disconnect(uint(suite.cfg.MqttQuiesceMs))
}

func getCtrFeatureID(topic string) string {
	result := strings.Split(topic, "/")
	return result[2]
}
