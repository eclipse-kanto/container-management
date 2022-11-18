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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/eclipse/ditto-clients-golang"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Broker        string `def:"tcp://localhost:1883"`
	MqttQuiesceMs int    `def:"500"`
	DittoAddress  string
	DittoUser     string `def:"ditto"`
	DittoPassword string `def:"ditto"`
}

type testConnection struct {
	cfg                 *testConfig
	mqttClient          mqtt.Client
	dittoClient         *ditto.Client
	ContainerURL        string
	ContainerFactoryURL string
}

type edgeDeviceConfig struct {
	DeviceID string `json:"deviceId"`
}

const (
	envVariablesPrefix = "SCT"
	featureID          = "ContainerFactory"
)

func newTestConnection(t *testing.T) *testConnection {
	cfg := &testConfig{}
	result := &testConnection{}

	t.Log(getConfigHelp(*cfg, envVariablesPrefix))

	if err := initConfigFromEnv(cfg, envVariablesPrefix); err != nil {
		t.Skip(err)
	}

	t.Logf("test config: %+v", *cfg)

	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(uuid.New().String()).
		SetKeepAlive(30 * time.Second).
		SetCleanSession(true).
		SetAutoReconnect(true)

	mqttClient := mqtt.NewClient(opts)

	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		require.NoError(t, token.Error(), "connect to MQTT broker")
	}

	edgeDeviceCfg, err := getEdgeDeviceConfig(mqttClient)
	if err != nil {
		mqttClient.Disconnect(uint(cfg.MqttQuiesceMs))
		require.NoError(t, err, "get edge device config")
	}

	dittoClient, err := ditto.NewClientMQTT(mqttClient, ditto.NewConfiguration())
	if err == nil {
		err = dittoClient.Connect()
	}

	if err != nil {
		mqttClient.Disconnect(uint(cfg.MqttQuiesceMs))
		require.NoError(t, err, "initialize ditto client")
	}

	result.cfg = cfg
	result.dittoClient = dittoClient
	result.mqttClient = mqttClient
	result.ContainerURL = fmt.Sprintf("%s/api/2/things/%s", strings.TrimSuffix(cfg.DittoAddress, "/"), edgeDeviceCfg.DeviceID+":edge:containers")
	result.ContainerFactoryURL = fmt.Sprintf("%s/features/%s", result.ContainerURL, featureID)

	return result
}

func (test *testConnection) disconnect() {
	test.dittoClient.Disconnect()
	test.mqttClient.Disconnect(uint(test.cfg.MqttQuiesceMs))
}

func getEdgeDeviceConfig(mqttClient mqtt.Client) (*edgeDeviceConfig, error) {
	type result struct {
		cfg *edgeDeviceConfig
		err error
	}

	ch := make(chan result)

	if token := mqttClient.Subscribe("edge/thing/response", 1, func(client mqtt.Client, message mqtt.Message) {
		var cfg edgeDeviceConfig
		if err := json.Unmarshal(message.Payload(), &cfg); err != nil {
			ch <- result{nil, err}
		}
		ch <- result{&cfg, nil}
	}); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	if token := mqttClient.Publish("edge/thing/request", 1, false, ""); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}

	timeout := 5 * time.Second
	select {
	case result := <-ch:
		return result.cfg, result.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("thing config not received in %v", timeout)
	}
}

func (test *testConnection) doRequest(method string, url string) ([]byte, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(test.cfg.DittoUser, test.cfg.DittoPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s request failed: %s", method, url, resp.Status)
	}

	return io.ReadAll(resp.Body)
}
