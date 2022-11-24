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
	"net/http"
	"strings"

	"github.com/eclipse-kanto/kanto/integration/util"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/websocket"
)

type ctrManagementSuite struct {
	suite.Suite
	util.SuiteInitializer
	ctrThingID           string
	ctrThingURL          string
	ctrFactoryFeatureURL string
	topicCreated         string
	topicModified        string
	topicDeleted         string
}

const (
	statusCreated               = "CREATED"
	statusRunning               = "RUNNING"
	influxdbImageRef            = "docker.io/library/influxdb:1.8.4"
	httpdImageRef               = "docker.io/library/httpd:latest"
	ctrFactoryFeatureID         = "ContainerFactory"
	ctrFactoryFeatureDefinition = "[\"com.bosch.iot.suite.edge.containers:ContainerFactory:1.2.0\"]"
	typeEvents                  = "START-SEND-EVENTS"
	imageRefParam               = "imageRef"
	startParam                  = "start"
	configParam                 = "config"
	createOperation             = "create"
	createWithConfigOperation   = "createWithConfig"
	ctrStatusProperty           = "status"
)

func (suite *ctrManagementSuite) SetupCtrManagementSuite() {
	suite.Setup(suite.T())

	suite.ctrThingID = suite.ThingCfg.DeviceID + ":edge:containers"
	suite.ctrThingURL = util.GetThingURL(suite.Cfg.DigitalTwinAPIAddress, suite.ctrThingID)
	suite.ctrFactoryFeatureURL = util.GetFeatureURL(suite.ctrThingURL, ctrFactoryFeatureID)

	suite.topicCreated = util.GetTwinEventTopic(suite.ctrThingID, protocol.ActionCreated)
	suite.topicModified = util.GetTwinEventTopic(suite.ctrThingID, protocol.ActionModified)
	suite.topicDeleted = util.GetTwinEventTopic(suite.ctrThingID, protocol.ActionDeleted)

	suite.assertCtrFactoryFeature()
}

func getCtrFeatureID(topic string) string {
	result := strings.Split(topic, "/")
	return result[2]
}

func (suite *ctrManagementSuite) getActualCtrStatus(ctrFeatureID string) string {
	ctrPropertyPath := fmt.Sprintf("%s/features/%s/properties/status/state/status", suite.ctrThingURL, ctrFeatureID)
	body, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, ctrPropertyPath, nil)
	require.NoError(suite.T(), err, "failed to get the status property of the container feature: %s", ctrFeatureID)

	return strings.Trim(string(body), "\"")
}

func (suite *ctrManagementSuite) assertCtrFactoryFeature() {
	ctrFactoryDefinition := suite.ctrFactoryFeatureURL + "/definition"
	body, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, ctrFactoryDefinition, nil)

	require.NoError(suite.T(), err, "failed to get the container factory feature definition")
	require.Equal(suite.T(), ctrFactoryFeatureDefinition, string(body), "the container factory definition is not expected")
}

func (suite *ctrManagementSuite) createWSConnection() *websocket.Conn {
	wsConnection, err := util.NewDigitalTwinWSConnection(suite.Cfg)
	require.NoError(suite.T(), err, "failed to create a websocket connection")

	err = util.SubscribeForWSMessages(suite.Cfg, wsConnection, typeEvents, "like(resource:path,'/features/Container:*')")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", typeEvents)

	return wsConnection
}

func (suite *ctrManagementSuite) testCreate(operation string, params interface{}, process func(*protocol.Envelope) (bool, error)) *websocket.Conn {
	wsConnection := suite.createWSConnection()

	_, err := util.ExecuteOperation(suite.Cfg, suite.ctrFactoryFeatureURL, operation, params)
	require.NoError(suite.T(), err, "failed to execute the %s operation", operation)

	err = util.ProcessWSMessages(suite.Cfg, wsConnection, process)
	require.NoError(suite.T(), err, "event for creating the container feature is not received")

	return wsConnection
}

func (suite *ctrManagementSuite) testRemove(connEvents *websocket.Conn, ctrFeatureID string) {
	filter := fmt.Sprintf("like(resource:path,'/features/%s')", ctrFeatureID)
	err := util.SubscribeForWSMessages(suite.Cfg, connEvents, typeEvents, filter)
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", typeEvents)

	_, err = util.ExecuteOperation(suite.Cfg, util.GetFeatureURL(suite.ctrThingURL, ctrFeatureID), "remove", true)
	require.NoError(suite.T(), err, "failed to remove the container feature with ID %s", ctrFeatureID)

	err = util.ProcessWSMessages(suite.Cfg, connEvents, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicDeleted {
			return true, nil
		}
		return false, fmt.Errorf("unknown message received")
	})

	require.NoError(suite.T(), err, "event for removing the container feature is not received")
}
