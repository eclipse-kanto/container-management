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

//go:build integration

package integration

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/eclipse-kanto/kanto/integration/util"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/websocket"
)

const (
	statusCreated               = "CREATED"
	statusRunning               = "RUNNING"
	requestURL                  = "http://127.0.0.1:5000"
	httpResponse                = "<html><body><h1>It works!</h1></body></html>\n"
	ctrFactoryFeatureDefinition = "[\"com.bosch.iot.suite.edge.containers:ContainerFactory:1.2.0\"]"
)

type ctrFactorySuite struct {
	ctrManagementSuite
	ctrFeatureID string
}

var ctrFeatureIDs []string
var isPropertyChanged bool

func (suite *ctrFactorySuite) SetupSuite() {
	suite.SetupCommonSuite()
	suite.assertContainerFactoryFeature()
}

func (suite *ctrFactorySuite) TearDownSuite() {
	for _, element := range ctrFeatureIDs {
		url := fmt.Sprintf("%s/features/%s", suite.ctrThingURL, element)
		body, _ := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, url, nil)
		if string(body) != "" {
			connEvents, err := util.NewDigitalTwinWSConnection(suite.Cfg)
			require.NoError(suite.T(), err, "failed to create websocket connection")
			defer connEvents.Close()
			suite.removeCtrFeature(connEvents, element)
		}
	}
	suite.TearDown()
}

func TestCtrFactorySuite(t *testing.T) {
	suite.Run(t, new(ctrFactorySuite))
}

func (suite *ctrFactorySuite) TestCreateCommand() {
	wsConnection := suite.createConnection()

	defer wsConnection.Close()

	params := make(map[string]interface{})
	params["imageRef"] = "docker.io/library/influxdb:1.8.4"
	params["start"] = true

	suite.create(params)

	err := util.ProcessWSMessages(suite.Cfg, wsConnection, suite.processCtrFeatureCreated)
	ctrFeatureIDs = append(ctrFeatureIDs, suite.ctrFeatureID)

	require.NoError(suite.T(), err, "error while creating container feature")
	require.Equal(suite.T(), statusRunning, suite.getActualCtrStatus(), "container status is not expected")

	suite.removeCtrFeature(wsConnection, suite.ctrFeatureID)
}

func (suite *ctrFactorySuite) TestCreateWithConfigCommand() {
	wsConnection := suite.createConnection()

	defer wsConnection.Close()

	params := make(map[string]interface{})
	params["imageRef"] = "docker.io/library/influxdb:1.8.4"
	params["start"] = true
	params["config"] = make(map[string]interface{})

	suite.createWithConfig(params)

	err := util.ProcessWSMessages(suite.Cfg, wsConnection, suite.processCtrFeatureCreated)
	ctrFeatureIDs = append(ctrFeatureIDs, suite.ctrFeatureID)

	require.NoError(suite.T(), err, "error while creating container feature")
	require.Equal(suite.T(), statusRunning, suite.getActualCtrStatus(), "container status is not expected")

	suite.removeCtrFeature(wsConnection, suite.ctrFeatureID)
}

func (suite *ctrFactorySuite) TestCreateWithConfigPortMapping() {
	wsConnection := suite.createConnection()

	defer wsConnection.Close()

	config := make(map[string]interface{})
	config["extraHosts"] = []string{"ctrhost:host_ip"}
	config["portMappings"] = []map[string]interface{}{
		{
			"hostPort":      5000,
			"containerPort": 80,
		},
	}

	params := make(map[string]interface{})
	params["imageRef"] = "docker.io/library/httpd:latest"
	params["start"] = true
	params["config"] = config

	suite.createWithConfig(params)

	err := util.ProcessWSMessages(suite.Cfg, wsConnection, suite.processCtrFeatureCreated)
	ctrFeatureIDs = append(ctrFeatureIDs, suite.ctrFeatureID)

	body, err := doRequest()

	require.NoError(suite.T(), err, "failed to reach requested URL on host from the running container")
	require.Equal(suite.T(), httpResponse, string(body), "HTTP response from the running container is not expected")

	suite.removeCtrFeature(wsConnection, suite.ctrFeatureID)
}

func (suite *ctrFactorySuite) getActualCtrStatus() string {
	ctrPropertyPath := fmt.Sprintf("%s/features/%s/properties/status/state/status", suite.ctrThingURL, suite.ctrFeatureID)
	body, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, ctrPropertyPath, nil)
	require.NoError(suite.T(), err, "error while getting the container feature property status")

	return strings.Trim(string(body), "\"")
}

func (suite *ctrFactorySuite) processCtrFeatureCreated(event *protocol.Envelope) (bool, error) {
	if event.Topic.String() == suite.topicCreated {
		suite.ctrFeatureID = getCtrFeatureID(event.Path)
		return false, nil
	}
	if event.Topic.String() == suite.topicModify {
		if suite.ctrFeatureID == "" {
			return true, fmt.Errorf("event for creating container feature is not received")
		}
		status, check := event.Value.(map[string]interface{})
		if status["status"].(string) == statusCreated {
			isPropertyChanged = true
			return false, nil
		}
		if isPropertyChanged && status["status"].(string) == statusRunning {
			return check && status["status"].(string) == statusRunning, nil
		}
		return true, fmt.Errorf("event for modify container feature status is not received")
	}
	return false, fmt.Errorf("container feature is not created")
}

func (suite *ctrFactorySuite) removeCtrFeature(connEvents *websocket.Conn, ctrFeatureID string) {
	suite.startListening(connEvents, "START-SEND-EVENTS", fmt.Sprintf("/features/%s", ctrFeatureID))

	suite.remove(ctrFeatureID)

	err := util.ProcessWSMessages(suite.Cfg, connEvents, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicDeleted {
			return true, nil
		}
		return false, fmt.Errorf("event for deleting feature not received")
	})

	require.NoError(suite.T(), err, "error while deleting container feature")
}

func (suite *ctrFactorySuite) createConnection() *websocket.Conn {
	wsConnection, err := util.NewDigitalTwinWSConnection(suite.Cfg)
	require.NoError(suite.T(), err, "failed to create websocket connection")
	suite.startListening(wsConnection, "START-SEND-EVENTS", "/features/Container:*")
	return wsConnection
}

func (suite *ctrFactorySuite) assertContainerFactoryFeature() {
	ctrFactoryFeature := fmt.Sprintf("%s/features/%s", suite.ctrThingURL, ctrFactoryFeatureID)
	body, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, ctrFactoryFeature, nil)
	require.NoError(suite.T(), err, "error while getting the container factory feature")

	ctrFactoryDefinition := fmt.Sprintf("%s/definition", ctrFactoryFeature)
	body, err = util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, ctrFactoryDefinition, nil)
	require.NoError(suite.T(), err, "error while getting the container factory feature definition")

	require.Equal(suite.T(), ctrFactoryFeatureDefinition, string(body), "container factory definition is not expected")
}

func doRequest() ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s request failed: %s", http.MethodGet, requestURL, resp.Status)
	}

	return io.ReadAll(resp.Body)
}
