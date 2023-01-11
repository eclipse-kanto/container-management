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

	suite.assertCtrFeatureDefinition(suite.ctrFactoryFeatureURL, "[\"com.bosch.iot.suite.edge.containers:ContainerFactory:1.3.0\"]")
}

func (suite *ctrManagementSuite) assertCtrFeatureDefinition(featureURL, expectedCtrDefinition string) {
	actualCtrDefinition := featureURL + "/definition"
	body, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, actualCtrDefinition, nil)

	require.NoError(suite.T(), err, "failed to get the container feature feature definition")
	require.Equal(suite.T(), expectedCtrDefinition, string(body), "the container feature definition is not expected")
}

func (suite *ctrManagementSuite) createWSConnection() *websocket.Conn {
	var (
		wsConnection *websocket.Conn
		err          error
	)
	defer func() {
		if wsConnection != nil {
			suite.closeOnError(wsConnection, err, "failed to subscribe for the %s messages", util.StartSendEvents)
		}
	}()

	wsConnection, err = util.NewDigitalTwinWSConnection(suite.Cfg)
	require.NoError(suite.T(), err, "failed to create a websocket connection")

	err = util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "like(resource:path,'/features/Container:*')")
	return wsConnection
}

func (suite *ctrManagementSuite) createOperation(operation string, params map[string]interface{}) string {
	wsConnection := suite.createWSConnection()

	defer func() {
		if wsConnection != nil {
			util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)
			wsConnection.Close()
		}
	}()

	ctrID, err := util.ExecuteOperation(suite.Cfg, suite.ctrFactoryFeatureURL, operation, params)
	suite.closeOnError(wsConnection, err, "failed to execute the %s operation", operation)

	var (
		ctrFeatureID   string
		isCtrCreated   bool
		eventValue     map[string]interface{}
		propertyStatus string
		isCtrStarted   bool
	)

	err = util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicCreated {
			ctrFeatureID = getCtrFeatureID(event.Path)
			err := suite.assertCtrID(ctrFeatureID, string(ctrID))
			require.NoError(suite.T(), err, "container ID is not expected")
			definition, err := getCtrDefinition(event.Value)
			require.NoError(suite.T(), err, "failed to parse property definition")
			require.Equal(suite.T(), "com.bosch.iot.suite.edge.containers:Container:1.5.0", definition, "container feature definition is not expected")
			return false, nil
		}
		if event.Topic.String() == suite.topicModified {
			if ctrFeatureID == "" {
				return true, fmt.Errorf("event for creating the container feature is not received")
			}
			propertyStatePath := fmt.Sprintf("/features/%s/properties/status/state", ctrFeatureID)
			if propertyStatePath != event.Path {
				return true, fmt.Errorf("received event is not expected")
			}
			eventValue, err = parseMap(event.Value)
			require.NoError(suite.T(), err, "failed to parse event value")

			propertyStatus, err = parseString(eventValue["status"])
			require.NoError(suite.T(), err, "failed to parse property status")

			isCtrStarted, err = parseBool(params[paramStart])
			require.NoError(suite.T(), err, "failed to parse param start")

			if propertyStatus == "CREATED" {
				if isCtrStarted {
					isCtrCreated = true
					return false, nil
				}
				return true, nil
			}
			if propertyStatus == "RUNNING" {
				if isCtrStarted && isCtrCreated {
					return true, nil
				}
				return true, fmt.Errorf("container status is not expected")
			}
			return true, fmt.Errorf("event for an unexpected container status is received")
		}
		return true, fmt.Errorf("unknown message is received")
	})
	suite.closeOnError(wsConnection, err, "failed to process creating the container feature")
	return ctrFeatureID
}

func getCtrFeatureID(path string) string {
	result := strings.Split(path, "/")
	if len(result) < 3 {
		return ""
	}
	return result[2]
}

func (suite *ctrManagementSuite) assertCtrID(ctrFeatureID, ctrID string) error {
	s := strings.Split(ctrFeatureID, ":")
	if len(s) < 2 {
		return fmt.Errorf("failed to get container ID from container feature ID")
	}
	s1 := strings.Trim(ctrID, "\"")
	require.Equal(suite.T(), s1, s[1], "container ID is not expected")
	return nil
}

func parseMap(value interface{}) (map[string]interface{}, error) {
	property, check := value.(map[string]interface{})
	if !check {
		return nil, fmt.Errorf("failed to parse the property to map")
	}
	return property, nil
}

func getCtrDefinition(value interface{}) (string, error) {
	eventValue, err := parseMap(value)
	if err != nil {
		return "", err
	}
	property, check := eventValue["definition"].([]interface{})
	if !check {
		return "", fmt.Errorf("failed to parse the property definition")
	}
	if len(property) != 1 {
		return "", fmt.Errorf("property definition type is not expected")
	}
	return parseString(property[0])
}

func parseString(value interface{}) (string, error) {
	property, check := value.(string)
	if !check {
		return "", fmt.Errorf("failed to parse the property to string")
	}
	return property, nil
}

func parseBool(value interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	property, check := value.(bool)
	if !check {
		return false, fmt.Errorf("failed to parse the property")
	}
	return property, nil
}

func (suite *ctrManagementSuite) create(params map[string]interface{}) string {
	return suite.createOperation("create", params)
}

func (suite *ctrManagementSuite) createWithConfig(params map[string]interface{}) string {
	return suite.createOperation("createWithConfig", params)
}

func (suite *ctrManagementSuite) remove(ctrFeatureID string) {
	if ctrFeatureID == "" {
		return
	}
	wsConnection := suite.createWSConnection()

	defer func() {
		if wsConnection != nil {
			util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)
			wsConnection.Close()
		}
	}()

	filter := fmt.Sprintf("like(resource:path,'/features/%s')", ctrFeatureID)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, filter)
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	_, err = util.ExecuteOperation(suite.Cfg, util.GetFeatureURL(suite.ctrThingURL, ctrFeatureID), "remove", true)
	require.NoError(suite.T(), err, "failed to remove the container feature with ID %s", ctrFeatureID)

	err = util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicDeleted {
			return true, nil
		}
		return true, fmt.Errorf("unknown message is received")
	})
	require.NoError(suite.T(), err, "failed to process removing the container feature")

}

func (suite *ctrManagementSuite) isCtrFeatureAvailable(ctrFeatureID string) string {
	ctrFeatureURL := util.GetFeatureURL(suite.ctrThingURL, ctrFeatureID)
	_, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, ctrFeatureURL, nil)
	if err != nil {
		return ""
	}
	return ctrFeatureURL
}

func (suite *ctrManagementSuite) closeOnError(wsConnection *websocket.Conn, err error, message string, messageArs ...interface{}) {
	if err != nil {
		wsConnection.Close()
	}
	require.NoError(suite.T(), err, message, messageArs)
}
