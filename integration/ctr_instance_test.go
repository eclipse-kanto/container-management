// Copyright (c) 2023 Contributors to the Eclipse Foundation
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
	"reflect"
	"testing"

	"github.com/eclipse-kanto/kanto/integration/util"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/websocket"
)

var emptyParams = make(map[string]interface{})

type ctrInstanceSuite struct {
	ctrManagementSuite
}

func (suite *ctrInstanceSuite) SetupSuite() {
	suite.SetupCtrManagementSuite()
}

func (suite *ctrInstanceSuite) TearDownSuite() {
	suite.TearDown()
}

func TestCtrInstanceSuite(t *testing.T) {
	suite.Run(t, new(ctrInstanceSuite))
}

func (suite *ctrInstanceSuite) TestStartContainer() {
	ctrFeatureID := suite.createStoppedContainer()

	defer suite.remove(ctrFeatureID)

	wsConnection, _ := util.NewDigitalTwinWSConnection(suite.Cfg)

	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	suite.executeWithExpectedSuccess(ctrFeatureID, "start", emptyParams)

	suite.processStateChange(wsConnection, ctrFeatureID, "RUNNING")
}

func (suite *ctrInstanceSuite) TestStartContainerThatIsAlreadyStarted() {
	ctrFeatureID := suite.createStartedContainer()

	defer suite.remove(ctrFeatureID)

	suite.executeWithExpectedError(ctrFeatureID, "start", emptyParams)
}

func (suite *ctrInstanceSuite) TestStopContainer() {
	ctrFeatureID := suite.createStartedContainer()

	defer suite.remove(ctrFeatureID)

	wsConnection, _ := util.NewDigitalTwinWSConnection(suite.Cfg)

	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	suite.executeWithExpectedSuccess(ctrFeatureID, "stop", emptyParams)

	suite.processStateChange(wsConnection, ctrFeatureID, "STOPPED")
}

func (suite *ctrInstanceSuite) TestStopContainerThatIsAlreadyStopped() {
	ctrFeatureID := suite.createStoppedContainer()

	defer suite.remove(ctrFeatureID)

	suite.executeWithExpectedError(ctrFeatureID, "stop", emptyParams)
}

func (suite *ctrInstanceSuite) TestPauseContainer() {
	ctrFeatureID := suite.createStartedContainer()

	defer suite.remove(ctrFeatureID)

	wsConnection, _ := util.NewDigitalTwinWSConnection(suite.Cfg)

	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	suite.executeWithExpectedSuccess(ctrFeatureID, "pause", emptyParams)

	suite.processStateChange(wsConnection, ctrFeatureID, "PAUSED")
}

func (suite *ctrInstanceSuite) TestResumeContainer() {
	ctrFeatureID := suite.createStartedContainer()

	defer suite.remove(ctrFeatureID)

	wsConnection, _ := util.NewDigitalTwinWSConnection(suite.Cfg)

	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	suite.executeWithExpectedSuccess(ctrFeatureID, "pause", emptyParams)

	suite.processStateChange(wsConnection, ctrFeatureID, "PAUSED")

	suite.executeWithExpectedSuccess(ctrFeatureID, "resume", emptyParams)

	suite.processStateChange(wsConnection, ctrFeatureID, "RUNNING")
}

func (suite *ctrInstanceSuite) TestRenameContainer() {
	ctrFeatureID := suite.createStoppedContainer()

	defer suite.remove(ctrFeatureID)

	wsConnection, _ := util.NewDigitalTwinWSConnection(suite.Cfg)

	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	newCtrName := "new_ctr_name"
	suite.executeWithExpectedSuccess(ctrFeatureID, "rename", newCtrName)

	suite.processNameChange(wsConnection, ctrFeatureID, newCtrName)
}

func (suite *ctrInstanceSuite) TestRemoveStoppedContainer() {
	ctrFeatureID := suite.createStoppedContainer()

	wsConnection, _ := util.NewDigitalTwinWSConnection(suite.Cfg)

	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	suite.executeWithExpectedSuccess(ctrFeatureID, "remove", false)

	suite.processRemove(wsConnection, ctrFeatureID)
}

func (suite *ctrInstanceSuite) TestRemoveStartedContainerWithForce() {
	ctrFeatureID := suite.createStartedContainer()

	wsConnection, _ := util.NewDigitalTwinWSConnection(suite.Cfg)

	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	suite.executeWithExpectedSuccess(ctrFeatureID, "remove", true)

	suite.processRemove(wsConnection, ctrFeatureID)
}

func (suite *ctrInstanceSuite) TestRemoveStartedContainerWithoutForce() {
	ctrFeatureID := suite.createStartedContainer()

	defer suite.remove(ctrFeatureID)

	suite.executeWithExpectedError(ctrFeatureID, "remove", false)
}

func (suite *ctrInstanceSuite) TestStopContainerWithOptions() {
	ctrFeatureID := suite.createStartedContainer()

	defer suite.remove(ctrFeatureID)

	wsConnection, _ := util.NewDigitalTwinWSConnection(suite.Cfg)

	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	params := map[string]string{"signal": "SIGINT"}
	suite.executeWithExpectedSuccess(ctrFeatureID, "stopWithOptions", params)

	suite.processStateChange(wsConnection, ctrFeatureID, "STOPPED")
}

func (suite *ctrInstanceSuite) TestUpdateContainer() {
	ctrFeatureID := suite.createStoppedContainer()

	defer suite.remove(ctrFeatureID)

	wsConnection, _ := util.NewDigitalTwinWSConnection(suite.Cfg)

	defer util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)

	err := util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)

	restartPolicyKey := "restartPolicy"
	newRestartPolicy := map[string]interface{}{"type": "ALWAYS"}
	params := map[string]interface{}{restartPolicyKey: newRestartPolicy}
	suite.executeWithExpectedSuccess(ctrFeatureID, "update", params)

	suite.processUpdate(wsConnection, ctrFeatureID, restartPolicyKey, newRestartPolicy)
}

func (suite *ctrInstanceSuite) createStartedContainer() string {
	params := make(map[string]interface{})
	params[paramImageRef] = influxdbImageRef
	params[paramStart] = true

	ctrFeatureID := suite.create(params)
	return ctrFeatureID
}

func (suite *ctrInstanceSuite) createStoppedContainer() string {
	params := make(map[string]interface{})
	params[paramImageRef] = influxdbImageRef
	params[paramStart] = false

	ctrFeatureID := suite.create(params)
	return ctrFeatureID
}

func (suite *ctrInstanceSuite) executeWithExpectedSuccess(ctrFeatureID string, operation string, params interface{}) {
	_, err := util.ExecuteOperation(suite.Cfg, util.GetFeatureURL(suite.ctrThingURL, ctrFeatureID), operation, params)
	require.NoError(suite.T(), err, "failed to perform \"%s\" operation on the container feature with ID %s", operation, ctrFeatureID)
}

func (suite *ctrInstanceSuite) executeWithExpectedError(ctrFeatureID string, operation string, params interface{}) {
	_, err := util.ExecuteOperation(suite.Cfg, util.GetFeatureURL(suite.ctrThingURL, ctrFeatureID), operation, params)
	require.Error(suite.T(), err)
}

func (suite *ctrInstanceSuite) processStateChange(wsConnection *websocket.Conn, ctrFeatureID string, expectedStatus string) {
	err := util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicModified {
			propertyStatePath := fmt.Sprintf("/features/%s/properties/status/state", ctrFeatureID)
			if propertyStatePath != event.Path {
				return true, fmt.Errorf("received event is not expected")
			}

			eventValue, err := parseMap(event.Value)
			require.NoError(suite.T(), err, "failed to parse event value")

			propertyStatus, err := parseString(eventValue["status"])
			require.NoError(suite.T(), err, "failed to parse property status")

			if propertyStatus != expectedStatus {
				return true, fmt.Errorf("event for an unexpected container status is received")
			}

			return true, nil
		}
		return true, fmt.Errorf("unknown message is received")
	})
	require.NoError(suite.T(), err, "failed to process updating the state of the container feature")
}

func (suite *ctrInstanceSuite) processNameChange(wsConnection *websocket.Conn, ctrFeatureID string, expectedName string) {
	err := util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicModified {
			propertyNamePath := fmt.Sprintf("/features/%s/properties/status/name", ctrFeatureID)
			if propertyNamePath != event.Path {
				return true, fmt.Errorf("received event is not expected")
			}

			propertyName, err := parseString(event.Value)
			require.NoError(suite.T(), err, "failed to parse property name")

			if propertyName != expectedName {
				return true, fmt.Errorf("event for an unexpected container status is received")
			}

			return true, nil
		}
		return true, fmt.Errorf("unknown message is received")
	})
	require.NoError(suite.T(), err, "failed to process updating the name of the container feature")
}

func (suite *ctrInstanceSuite) processRemove(wsConnection *websocket.Conn, ctrFeatureID string) {
	err := util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		deletedCtrPath := fmt.Sprintf("/features/%s", ctrFeatureID)
		if deletedCtrPath != event.Path {
			return true, fmt.Errorf("received event for unexpected container")
		}

		if event.Topic.String() != suite.topicDeleted {
			return true, fmt.Errorf("unknown message is received")
		}

		return true, nil
	})
	require.NoError(suite.T(), err, "failed to process removing the container feature")
}

func (suite *ctrInstanceSuite) processUpdate(wsConnection *websocket.Conn, ctrFeatureID string, expectedKey string, expectedValue map[string]interface{}) {
	err := util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicModified {
			statusUpdatePath := fmt.Sprintf("/features/%s/properties/status/config", ctrFeatureID)
			if statusUpdatePath != event.Path {
				return true, fmt.Errorf("received event is not expected")
			}

			eventValue, err := parseMap(event.Value)
			require.NoError(suite.T(), err, "failed to parse event value")

			actualValue, err := parseMap(eventValue[expectedKey])
			require.NoError(suite.T(), err, fmt.Sprintf("failed to parse value of key \"%s\"", expectedKey))

			if !reflect.DeepEqual(expectedValue, actualValue) {
				return true, fmt.Errorf("expected value - %s, got value - %s", expectedValue, actualValue)
			}

			return true, nil
		}
		return true, fmt.Errorf("unknown message is received")
	})
	require.NoError(suite.T(), err, "failed to process updating the config of the container feature")
}
