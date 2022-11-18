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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/websocket"
)

func (suite *containerManagementSuite) doRequest(method string, url string, reqBody *requestBody) ([]byte, error) {
	var body io.Reader

	if reqBody != nil {
		if reqBody.param != "" {
			body = bytes.NewBuffer([]byte(reqBody.param))
		}
		if len(reqBody.params) > 0 {
			jsonValue, err := json.Marshal(reqBody.params)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(jsonValue)
		}

	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if reqBody != nil {
		correlationID := uuid.New().String()
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("correlation-id", correlationID)
		req.Header.Add("response-required", "true")
	}

	req.SetBasicAuth(suite.cfg.DigitalTwinAPIUser, suite.cfg.DigitalTwinAPIPassword)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s request failed: %s", http.MethodPost, url, resp.Status)
	}

	return io.ReadAll(resp.Body)
}

func getEdgeDeviceConfig(mqttClient MQTT.Client) (*edgeConfig, error) {
	type result struct {
		cfg *edgeConfig
		err error
	}

	ch := make(chan result)

	if token := mqttClient.Subscribe("edge/thing/response", 1, func(client MQTT.Client, message MQTT.Message) {
		var cfg edgeConfig
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

func (suite *containerManagementSuite) newWSConnection() (*websocket.Conn, error) {
	wsAddress, err := asWSAddress(suite.cfg.DigitalTwinAPIAddress)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/ws/2", wsAddress)
	cfg, err := websocket.NewConfig(url, suite.cfg.DigitalTwinAPIAddress)
	if err != nil {
		return nil, err
	}

	auth := fmt.Sprintf("%s:%s", suite.cfg.DigitalTwinAPIUser, suite.cfg.DigitalTwinAPIPassword)
	enc := base64.StdEncoding.EncodeToString([]byte(auth))
	cfg.Header = http.Header{
		"Authorization": {"Basic " + enc},
	}

	return websocket.DialConfig(cfg)
}

func asWSAddress(address string) (string, error) {
	url, err := url.Parse(address)
	if err != nil {
		return "", err
	}

	if url.Scheme == "https" {
		return fmt.Sprintf("wss://%s:%s", url.Hostname(), url.Port()), nil
	}

	return fmt.Sprintf("ws://%s:%s", url.Hostname(), url.Port()), nil
}

func (suite *containerManagementSuite) beginWSWait(ws *websocket.Conn, check func(payload []byte) bool) chan bool {
	timeout := time.Duration(suite.cfg.EventTimeoutMs * int(time.Millisecond))

	ch := make(chan bool)

	go func() {
		resultCh := make(chan bool)

		go func() {
			var payload []byte
			threshold := time.Now().Add(timeout)
			for time.Now().Before(threshold) {
				err := websocket.Message.Receive(ws, &payload)
				if err == nil {
					if check(payload) {
						resultCh <- true
						return
					}
				} else {
					suite.T().Logf("error while waiting for WS message: %v", err)
				}
			}
			suite.T().Logf("WS response not received in %v", timeout)
			resultCh <- false
		}()
		result := suite.awaitChan(resultCh)
		ws.Close()
		ch <- result
	}()

	return ch
}

func (suite *containerManagementSuite) execCreateCommand(command string, params map[string]interface{}) {
	url := fmt.Sprintf("%s/inbox/messages/%s", suite.ctrFactoryFeatureURL, command)
	if _, err := suite.doRequest(http.MethodPost, url, &requestBody{params: params}); err != nil {
		suite.T().Errorf("error while creating container feature: %v", err)
	}

}

func (suite *containerManagementSuite) execRemoveCommand(ctrFeatureID string) {
	ctrFeatureURL := fmt.Sprintf("%s/features/%s", suite.ctrThingURL, ctrFeatureID)
	url := fmt.Sprintf("%s/inbox/messages/remove", ctrFeatureURL)
	if _, err := suite.doRequest(http.MethodPost, url, &requestBody{param: "true"}); err != nil {
		suite.T().Errorf("error while removing container feature: %v", err)
	}
}

func (suite *containerManagementSuite) startEventListener(eventType, filter string, matcher func(map[string]interface{}) bool) chan bool {
	ws, err := suite.newWSConnection()
	require.NoError(suite.T(), err)

	subAck := fmt.Sprintf("%s:ACK", eventType)
	var ackReceived bool
	ackChan := make(chan bool)
	wsListener := func(payload []byte) bool {
		ack := strings.TrimSpace(string(payload))
		if ack == subAck {
			ackReceived = true
			ackChan <- true
			return false
		}
		if !ackReceived {
			suite.T().Logf("skipping event, acknowledgement not received")
			return false
		}
		props := make(map[string]interface{})
		err := json.Unmarshal(payload, &props)
		if err == nil {
			return matcher(props)
		}

		suite.T().Logf("error while waiting for event: %v", err)
		return false
	}
	websocket.Message.Send(ws, fmt.Sprintf("%s?filter=like(resource:path,'%s')", eventType, filter))
	result := suite.beginWSWait(ws, wsListener)
	require.True(suite.T(), suite.awaitChan(ackChan), "event acknowledgement not received")
	return result
}

func (suite *containerManagementSuite) awaitChan(ch chan bool) bool {
	timeout := time.Duration(suite.cfg.EventTimeoutMs * int(time.Millisecond))
	select {
	case result := <-ch:
		return result
	case <-time.After(timeout):
		return false
	}
}

func (suite *containerManagementSuite) getCtrFeature(ctrFeatureID string) model.Feature {
	ctrThingURL := fmt.Sprintf("%s/features/%s", suite.ctrThingURL, ctrFeatureID)
	body, err := suite.doRequest(http.MethodGet, ctrThingURL, nil)

	if err != nil {
		suite.T().Errorf("error while getting the container feature: %v", err)
	}

	var containerFeature = &containerFeature{}
	json.Unmarshal(body, &containerFeature)

	return client.NewFeature(ctrFeatureID,
		client.WithFeatureDefinition(client.NewDefinitionIDFromString(containerFeature.Definition[0])),
		client.WithFeatureProperties(containerFeature.Properties))
}
