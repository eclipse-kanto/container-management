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

func (suite *ctrManagementSuite) SetupCtrManagementSuite() {
	const ctrFactoryFeatureID = "ContainerFactory"
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
	featureURL := util.GetFeatureURL(suite.ctrThingURL, ctrFeatureID)
	body, err := util.GetFeaturePropertyValue(suite.Cfg, featureURL, "status/state/status")
	require.NoError(suite.T(), err, "failed to get the property status of the container feature: %s", ctrFeatureID)

	return strings.Trim(string(body), "\"")
}

func (suite *ctrManagementSuite) assertCtrFactoryFeature() {
	const ctrFactoryFeatureDefinition = "[\"com.bosch.iot.suite.edge.containers:ContainerFactory:1.3.0\"]"
	ctrFactoryDefinition := suite.ctrFactoryFeatureURL + "/definition"
	body, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, ctrFactoryDefinition, nil)

	require.NoError(suite.T(), err, "failed to get the container factory feature definition")
	require.Equal(suite.T(), ctrFactoryFeatureDefinition, string(body), "the container factory definition is not expected")
}

func (suite *ctrManagementSuite) createWSConnection() *websocket.Conn {
	const filterCtrFeatures = "like(resource:path,'/features/Container:*')"

	wsConnection, err := util.NewDigitalTwinWSConnection(suite.Cfg)
	require.NoError(suite.T(), err, "failed to create a websocket connection")

	err = util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, filterCtrFeatures)
	suite.assertNoError(wsConnection, err, "failed to subscribe for the %s messages", util.StartSendEvents)
	return wsConnection
}

func (suite *ctrManagementSuite) createOperation(operation string, params map[string]interface{}) (*websocket.Conn, string) {
	const (
		propertyStatus = "status"
		statusCreated  = "CREATED"
		statusRunning  = "RUNNING"
	)

	wsConnection := suite.createWSConnection()

	_, err := util.ExecuteOperation(suite.Cfg, suite.ctrFactoryFeatureURL, operation, params)
	suite.assertNoError(wsConnection, err, "failed to execute the %s operation", operation)

	var ctrFeatureID string
	var isCtrFeatureCreated bool

	err = util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicCreated {
			ctrFeatureID = getCtrFeatureID(event.Path)
			return false, nil
		}
		if event.Topic.String() == suite.topicModified {
			if ctrFeatureID == "" {
				return true, fmt.Errorf("event for creating the container feature is not received")
			}
			status, check := event.Value.(map[string]interface{})
			if !check {
				return true, fmt.Errorf("failed to parse the property status value from the received event")
			}
			if status[propertyStatus].(string) == statusCreated {
				isCtrFeatureCreated = true
				return false, nil
			}
			if isCtrFeatureCreated {
				suite.expectedStatus(status[propertyStatus].(string), params[paramStart].(bool))
				return true, nil
			}
			return true, fmt.Errorf("event for modify the container feature status is not received")
		}
		return true, fmt.Errorf("unknown message is received")
	})
	suite.assertNoError(wsConnection, err, "failed to process creating the container feature")
	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)
	return wsConnection, ctrFeatureID
}

func (suite *ctrManagementSuite) expectedStatus(status string, isStarted bool) {
	const (
		statusCreated = "CREATED"
		statusRunning = "RUNNING"
	)
	if isStarted {
		require.Equal(suite.T(), statusRunning, status, "container status is not expected")
		return
	}
	require.Equal(suite.T(), statusCreated, status, "container status is not expected")
}

func (suite *ctrManagementSuite) create(params map[string]interface{}) (*websocket.Conn, string) {
	return suite.createOperation("create", params)
}

func (suite *ctrManagementSuite) createWithConfig(params map[string]interface{}) (*websocket.Conn, string) {
	return suite.createOperation("createWithConfig", params)
}

func (suite *ctrManagementSuite) remove(wsConnection *websocket.Conn, ctrFeatureID string) {
	const (
		filterCtrFeature = "like(resource:path,'/features/%s')"
		operationRemove  = "remove"
	)

	filter := fmt.Sprintf(filterCtrFeature, ctrFeatureID)
	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, filter)
	suite.assertNoError(wsConnection, err, "failed to subscribe for the %s messages", util.StartSendEvents)
	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	_, err = util.ExecuteOperation(suite.Cfg, util.GetFeatureURL(suite.ctrThingURL, ctrFeatureID), "remove", true)
	suite.assertNoError(wsConnection, err, "failed to remove the container feature with ID %s", ctrFeatureID)

	err = util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicDeleted {
			return true, nil
		}
		return true, fmt.Errorf("unknown message is received")
	})
	suite.assertNoError(wsConnection, err, "failed to process removing the container feature")

}

func (suite *ctrManagementSuite) assertNoError(wsConnection *websocket.Conn, err error, message string, messageArs ...interface{}) {
	if err != nil {
		wsConnection.Close()
	}
	require.NoError(suite.T(), err, message, messageArs)
}
