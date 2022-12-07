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
	suite.Setup(suite.T())

	suite.ctrThingID = suite.ThingCfg.DeviceID + ":edge:containers"
	suite.ctrThingURL = util.GetThingURL(suite.Cfg.DigitalTwinAPIAddress, suite.ctrThingID)
	suite.ctrFactoryFeatureURL = util.GetFeatureURL(suite.ctrThingURL, "ContainerFactory")

	suite.topicCreated = util.GetTwinEventTopic(suite.ctrThingID, protocol.ActionCreated)
	suite.topicModified = util.GetTwinEventTopic(suite.ctrThingID, protocol.ActionModified)
	suite.topicDeleted = util.GetTwinEventTopic(suite.ctrThingID, protocol.ActionDeleted)

	suite.assertCtrFactoryFeature()
}

func getCtrFeatureID(path string) string {
	result := strings.Split(path, "/")
	if len(result) < 3 {
		return ""
	}
	return result[2]
}

func (suite *ctrManagementSuite) getActualCtrStatus(ctrFeatureID string) string {
	featureURL := util.GetFeatureURL(suite.ctrThingURL, ctrFeatureID)
	body, err := util.GetFeaturePropertyValue(suite.Cfg, featureURL, "status/state/status")
	require.NoError(suite.T(), err, "failed to get the property status of the container feature: %s", ctrFeatureID)

	return strings.Trim(string(body), "\"")
}

func (suite *ctrManagementSuite) assertCtrFactoryFeature() {
	ctrFactoryDefinition := suite.ctrFactoryFeatureURL + "/definition"
	body, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, ctrFactoryDefinition, nil)

	require.NoError(suite.T(), err, "failed to get the container factory feature definition")
	require.Equal(suite.T(), "[\"com.bosch.iot.suite.edge.containers:ContainerFactory:1.3.0\"]", string(body), "the container factory definition is not expected")
}

func (suite *ctrManagementSuite) createWSConnection() *websocket.Conn {
	wsConnection, err := util.NewDigitalTwinWSConnection(suite.Cfg)
	require.NoError(suite.T(), err, "failed to create a websocket connection")

	err = util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "like(resource:path,'/features/Container:*')")
	suite.closeOnError(wsConnection, err, "failed to subscribe for the %s messages", util.StartSendEvents)
	return wsConnection
}

func (suite *ctrManagementSuite) createOperation(operation string, params map[string]interface{}) (*websocket.Conn, string) {
	wsConnection := suite.createWSConnection()

	_, err := util.ExecuteOperation(suite.Cfg, suite.ctrFactoryFeatureURL, operation, params)
	suite.closeOnError(wsConnection, err, "failed to execute the %s operation", operation)

	const propertyStatus = "status"
	var ctrFeatureID string
	var isCtrFeatureCreated bool
	var eventValue map[string]interface{}

	err = util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicCreated {
			ctrFeatureID = getCtrFeatureID(event.Path)
			if eventValue, err = parseEventValue(event.Value.(map[string]interface{})); err != nil {
				return true, err
			}
			if eventValue["definition"].([]interface{})[0] != "com.bosch.iot.suite.edge.containers:Container:1.5.0" {
				return true, fmt.Errorf("container feature definition is not expected")
			}
			return false, nil
		}
		if event.Topic.String() == suite.topicModified {
			if ctrFeatureID == "" {
				return true, fmt.Errorf("event for creating the container feature is not received")
			}
			ctrState := fmt.Sprintf("/features/%s/properties/status/state", ctrFeatureID)
			if ctrState != event.Path {
				return true, fmt.Errorf("received event is not expected")
			}
			if eventValue, err = parseEventValue(event.Value.(map[string]interface{})); err != nil {
				return true, err
			}
			if eventValue[propertyStatus].(string) == "CREATED" {
				if params[paramStart].(bool) {
					isCtrFeatureCreated = true
					return false, nil
				}
				return true, nil
			}
			if eventValue[propertyStatus].(string) == "RUNNING" {
				if params[paramStart].(bool) && isCtrFeatureCreated {
					return true, nil
				}
				return true, fmt.Errorf("container status is not expected")
			}
			return true, fmt.Errorf("event for modify the container feature status is not received")
		}
		return true, fmt.Errorf("unknown message is received")
	})
	suite.closeOnError(wsConnection, err, "failed to process creating the container feature")
	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)
	return wsConnection, ctrFeatureID
}

func parseEventValue(eventValue interface{}) (map[string]interface{}, error) {
	property, check := eventValue.(map[string]interface{})
	if !check {
		return nil, fmt.Errorf("failed to parse the property event value")
	}
	return property, nil
}

func (suite *ctrManagementSuite) create(params map[string]interface{}) (*websocket.Conn, string) {
	return suite.createOperation("create", params)
}

func (suite *ctrManagementSuite) createWithConfig(params map[string]interface{}) (*websocket.Conn, string) {
	return suite.createOperation("createWithConfig", params)
}

func (suite *ctrManagementSuite) remove(wsConnection *websocket.Conn, ctrFeatureID string) {
	filter := fmt.Sprintf("like(resource:path,'/features/%s')", ctrFeatureID)
	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, filter)
	suite.closeOnError(wsConnection, err, "failed to subscribe for the %s messages", util.StartSendEvents)
	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	_, err = util.ExecuteOperation(suite.Cfg, util.GetFeatureURL(suite.ctrThingURL, ctrFeatureID), "remove", true)
	suite.closeOnError(wsConnection, err, "failed to remove the container feature with ID %s", ctrFeatureID)

	err = util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicDeleted {
			return true, nil
		}
		return true, fmt.Errorf("unknown message is received")
	})
	suite.closeOnError(wsConnection, err, "failed to process removing the container feature")

}

func (suite *ctrManagementSuite) closeOnError(wsConnection *websocket.Conn, err error, message string, messageArs ...interface{}) {
	if err != nil {
		wsConnection.Close()
	}
	require.NoError(suite.T(), err, message, messageArs)
}
