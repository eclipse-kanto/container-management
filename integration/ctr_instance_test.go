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

const (
	operationStart           = "start"
	operationStop            = "stop"
	operationPause           = "pause"
	operationResume          = "resume"
	operationRename          = "rename"
	operationRemove          = "remove"
	operationUpdate          = "update"
	operationStopWithOptions = "stopWithOptions"
)

type ctrInstanceSuite struct {
	ctrManagementSuite
}

func (suite *ctrInstanceSuite) SetupSuite() {
	suite.SetupCtrManagementSuite()
}

func (suite *ctrInstanceSuite) TearDownSuite() {
	suite.TearDown()
}

func (suite *ctrInstanceSuite) setupCtrInstanceTest(ctrStart bool) (string, *websocket.Conn) {
	ctrFeatureID := suite.createContainer(ctrStart)
	wsConnection, err := util.NewDigitalTwinWSConnection(suite.Cfg)
	require.NoError(suite.T(), err, "failed to create a websocket connection")
	err = util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendEvents, "")
	require.NoError(suite.T(), err, "failed to subscribe for the %s messages", util.StartSendEvents)
	return ctrFeatureID, wsConnection
}

func (suite *ctrInstanceSuite) tearDownCtrInstanceTest(ctrFeatureID string, wsConnection *websocket.Conn) {
	if wsConnection != nil {
		util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendEvents)
		wsConnection.Close()
	}
	if ctrFeatureID != "" {
		suite.remove(ctrFeatureID)
	}
}

func TestCtrInstanceSuite(t *testing.T) {
	suite.Run(t, new(ctrInstanceSuite))
}

func (suite *ctrInstanceSuite) TestCtrInstanceOperations() {
	tests := map[string]struct {
		ctrStart          bool
		exec              func(ctrFeatureID string, wsConnection *websocket.Conn)
		ctrHasToBeRemoved bool
	}{
		"test_start_container": {
			ctrStart: false,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationStart, emptyParams)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusRunning)
			},
			ctrHasToBeRemoved: true,
		},
		"test_start_container_that_is_already_started": {
			ctrStart: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedError(ctrFeatureID, operationStart, emptyParams)
			},
			ctrHasToBeRemoved: true,
		},
		"test_stop_container": {
			ctrStart: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationStop, emptyParams)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusStopped)
			},
			ctrHasToBeRemoved: true,
		},
		"test_stop_container_that_is_already_stopped": {
			ctrStart: false,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedError(ctrFeatureID, operationStop, emptyParams)
			},
			ctrHasToBeRemoved: true,
		},
		"test_pause_container": {
			ctrStart: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationPause, emptyParams)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusPaused)
			},
			ctrHasToBeRemoved: true,
		},
		"test_resume_container": {
			ctrStart: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationPause, emptyParams)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusPaused)
				suite.executeWithExpectedSuccess(ctrFeatureID, operationResume, emptyParams)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusRunning)
			},
			ctrHasToBeRemoved: true,
		},
		"test_rename_container": {
			ctrStart: false,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				newCtrName := "new_ctr_name"
				suite.executeWithExpectedSuccess(ctrFeatureID, operationRename, newCtrName)
				suite.processNameChange(wsConnection, ctrFeatureID, newCtrName)
			},
			ctrHasToBeRemoved: true,
		},
		"test_remove_stopped_container": {
			ctrStart: false,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationRemove, false)
				suite.processRemove(wsConnection, ctrFeatureID)
			},
			ctrHasToBeRemoved: false,
		},
		"test_remove_started_container_with_force": {
			ctrStart: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationRemove, true)
				suite.processRemove(wsConnection, ctrFeatureID)
			},
			ctrHasToBeRemoved: false,
		},
		"test_remove_started_container_without_force": {
			ctrStart: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedError(ctrFeatureID, operationRemove, false)
			},
			ctrHasToBeRemoved: true,
		},
		"test_stop_container_with_options": {
			ctrStart: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				params := map[string]string{"signal": "SIGINT"}
				suite.executeWithExpectedSuccess(ctrFeatureID, operationStopWithOptions, params)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusStopped)
			},
			ctrHasToBeRemoved: true,
		},
		"test_update_container": {
			ctrStart: false,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				restartPolicyKey := "restartPolicy"
				newRestartPolicy := map[string]interface{}{"type": "ALWAYS"}
				params := map[string]interface{}{restartPolicyKey: newRestartPolicy}
				suite.executeWithExpectedSuccess(ctrFeatureID, operationUpdate, params)
				suite.processUpdate(wsConnection, ctrFeatureID, restartPolicyKey, newRestartPolicy)
			},
			ctrHasToBeRemoved: true,
		},
	}

	for testName, testCase := range tests {
		suite.Run(testName, func() {
			ctrFeatureID, wsConnection := suite.setupCtrInstanceTest(testCase.ctrStart)
			defer func() {
				if testCase.ctrHasToBeRemoved {
					suite.tearDownCtrInstanceTest(ctrFeatureID, wsConnection)
				} else {
					suite.tearDownCtrInstanceTest("", wsConnection)
				}
			}()

			testCase.exec(ctrFeatureID, wsConnection)
		})
	}
}

func (suite *ctrInstanceSuite) createContainer(start bool) string {
	params := map[string]interface{}{
		paramImageRef: influxdbImageRef,
		paramStart:    start,
	}
	return suite.create(params)
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
			if event.Path != fmt.Sprintf("/features/%s/properties/status/state", ctrFeatureID) {
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
			if event.Path != fmt.Sprintf("/features/%s/properties/status/name", ctrFeatureID) {
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
		if event.Topic.String() == suite.topicDeleted {
			if event.Path != fmt.Sprintf("/features/%s", ctrFeatureID) {
				return true, fmt.Errorf("received event for unexpected container")
			}

			return true, nil
		}
		return true, fmt.Errorf("unknown message is received")
	})
	require.NoError(suite.T(), err, "failed to process removing the container feature")
}

func (suite *ctrInstanceSuite) processUpdate(wsConnection *websocket.Conn, ctrFeatureID string, expectedKey string, expectedValue map[string]interface{}) {
	err := util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicModified {
			if event.Path != fmt.Sprintf("/features/%s/properties/status/config", ctrFeatureID) {
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
