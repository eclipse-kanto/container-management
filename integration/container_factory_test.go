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
	"encoding/json"
	"fmt"
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
	containerManagementSuite
	ctrFeatureID string
}

func (suite *ctrFactorySuite) SetupSuite() {
	suite.init()
	suite.assertContainerFactoryFeature()
}

func (suite *ctrFactorySuite) TearDownSuite() {
	url := fmt.Sprintf("%s/features", suite.ctrThingURL)
	body, err := util.SendDigitalTwinRequest(suite.suiteInit.Cfg, http.MethodGet, url, nil)
	require.NoError(suite.T(), err, "failed to get container thing features")

	features := make(map[string]interface{})
	err = json.Unmarshal(body, &features)
	require.NoError(suite.T(), err, "failed to unmarshal container thing features")

	for key := range features {
		if strings.Contains(key, "Container:") {
			connEvents, err := util.NewDigitalTwinWSConnection(suite.suiteInit.Cfg)
			require.NoError(suite.T(), err, "failed to create websocket connection")
			defer connEvents.Close()
			suite.processCtrFeatureRemoved(connEvents, key)
		}
	}

	suite.suiteInit.TearDown()
}

func TestContainerFactorySuite(t *testing.T) {
	suite.Run(t, new(ctrFactorySuite))
}

func (suite *ctrFactorySuite) TestCreateOperation() {
	connEvents, err := util.NewDigitalTwinWSConnection(suite.suiteInit.Cfg)
	require.NoError(suite.T(), err, "failed to create websocket connection")
	defer connEvents.Close()

	suite.startListening(connEvents, "START-SEND-EVENTS", "/features/Container:*")

	params := make(map[string]interface{})
	params["imageRef"] = "docker.io/library/influxdb:1.8.4"
	params["start"] = true

	suite.execCreateCommand("create", params)

	err = util.ProcessWSMessages(suite.suiteInit.Cfg, connEvents, suite.processCtrFeatureCreated)

	require.NoError(suite.T(), err, "error while creating container feature")
	require.Equal(suite.T(), statusRunning, suite.getActualCtrStatus(), "container status is not expected")

	suite.processCtrFeatureRemoved(connEvents, suite.ctrFeatureID)
}

func (suite *ctrFactorySuite) TestCreateWithConfigOperation() {
	connEvents, err := util.NewDigitalTwinWSConnection(suite.suiteInit.Cfg)
	require.NoError(suite.T(), err, "failed to create websocket connection")
	defer connEvents.Close()

	suite.startListening(connEvents, "START-SEND-EVENTS", "/features/Container:*")

	params := make(map[string]interface{})
	params["imageRef"] = "docker.io/library/influxdb:1.8.4"
	params["start"] = true
	params["config"] = make(map[string]interface{})

	suite.execCreateCommand("createWithConfig", params)

	err = util.ProcessWSMessages(suite.suiteInit.Cfg, connEvents, suite.processCtrFeatureCreated)

	require.NoError(suite.T(), err, "error while creating container feature")
	require.Equal(suite.T(), statusRunning, suite.getActualCtrStatus(), "container status is not expected")

	suite.processCtrFeatureRemoved(connEvents, suite.ctrFeatureID)
}

func (suite *ctrFactorySuite) TestCreateWithConfigPortMapping() {
	connEvents, err := util.NewDigitalTwinWSConnection(suite.suiteInit.Cfg)
	require.NoError(suite.T(), err, "failed to create websocket connection")
	defer connEvents.Close()

	suite.startListening(connEvents, "START-SEND-EVENTS", "/features/Container:*")

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

	suite.execCreateCommand("createWithConfig", params)

	err = util.ProcessWSMessages(suite.suiteInit.Cfg, connEvents, suite.processCtrFeatureCreated)

	body, err := util.SendDigitalTwinRequest(suite.suiteInit.Cfg, http.MethodGet, requestURL, nil)

	require.NoError(suite.T(), err, "error while getting the requested URL")
	require.Equal(suite.T(), httpResponse, string(body), "HTTP response is not expected")

	suite.processCtrFeatureRemoved(connEvents, suite.ctrFeatureID)
}

func (suite *containerManagementSuite) assertContainerFactoryFeature() {
	ctrFactoryFeature := fmt.Sprintf("%s/features/%s", suite.ctrThingURL, ctrFactoryFeatureID)
	body, err := util.SendDigitalTwinRequest(suite.suiteInit.Cfg, http.MethodGet, ctrFactoryFeature, nil)
	require.NoError(suite.T(), err, "error while getting the container factory feature")

	ctrFactoryDefinition := fmt.Sprintf("%s/definition", ctrFactoryFeature)
	body, err = util.SendDigitalTwinRequest(suite.suiteInit.Cfg, http.MethodGet, ctrFactoryDefinition, nil)
	require.NoError(suite.T(), err, "error while getting the container factory feature definition")

	require.Equal(suite.T(), ctrFactoryFeatureDefinition, string(body), "container factory definition is not expected")
}

func (suite *ctrFactorySuite) getActualCtrStatus() string {
	ctrPropertyPath := fmt.Sprintf("%s/features/%s/properties/status/state/status", suite.ctrThingURL, suite.ctrFeatureID)
	body, err := util.SendDigitalTwinRequest(suite.suiteInit.Cfg, http.MethodGet, ctrPropertyPath, nil)
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
			return false, fmt.Errorf("container feature is not created")
		}
		status, check := event.Value.(map[string]interface{})
		if status["status"].(string) == statusCreated {
			return false, nil
		}
		return check && status["status"].(string) == statusRunning, nil
	}
	return false, fmt.Errorf("event for creating feature not received")
}

func (suite *ctrFactorySuite) processCtrFeatureRemoved(connEvents *websocket.Conn, ctrFeatureID string) {
	suite.startListening(connEvents, "START-SEND-EVENTS", fmt.Sprintf("/features/%s", ctrFeatureID))

	suite.execRemoveCommand(ctrFeatureID)

	err := util.ProcessWSMessages(suite.suiteInit.Cfg, connEvents, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicDeleted {
			return true, nil
		}
		return false, fmt.Errorf("event for deleting feature not received")
	})

	require.NoError(suite.T(), err, "error while deleting container feature")
}
