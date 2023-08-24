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

	"github.com/eclipse-kanto/container-management/containerm/things"
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

	parseError                    = "failed to parse %s"
	messageProcessError           = "failed to process %s the container feature"
	unexpectedContainerEventError = "received event is not expected"
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
		ctrHasToBeRemoved bool
		exec              func(ctrFeatureID string, wsConnection *websocket.Conn)
	}{
		"test_start_container": {
			ctrHasToBeRemoved: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationStart, emptyParams)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusRunning)
			},
		},
		"test_start_container_that_is_already_running": {
			ctrStart:          true,
			ctrHasToBeRemoved: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedError(ctrFeatureID, operationStart, emptyParams)
			},
		},
		"test_stop_container": {
			ctrStart:          true,
			ctrHasToBeRemoved: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationStop, emptyParams)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusStopped)
			},
		},
		"test_stop_container_that_is_not_running": {
			ctrHasToBeRemoved: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedError(ctrFeatureID, operationStop, emptyParams)
			},
		},
		"test_pause_container": {
			ctrStart:          true,
			ctrHasToBeRemoved: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationPause, emptyParams)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusPaused)
			},
		},
		"test_resume_container": {
			ctrStart:          true,
			ctrHasToBeRemoved: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationPause, emptyParams)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusPaused)
				suite.executeWithExpectedSuccess(ctrFeatureID, operationResume, emptyParams)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusRunning)
			},
		},
		"test_rename_container": {
			ctrHasToBeRemoved: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				newCtrName := "new_ctr_name"
				suite.executeWithExpectedSuccess(ctrFeatureID, operationRename, newCtrName)
				suite.processNameChange(wsConnection, ctrFeatureID, newCtrName)
			},
		},
		"test_remove_container_that_is_not_running": {
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationRemove, false)
				suite.processRemove(wsConnection, ctrFeatureID)
			},
		},
		"test_remove_container_that_is_running_with_force": {
			ctrStart: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedSuccess(ctrFeatureID, operationRemove, true)
				suite.processRemove(wsConnection, ctrFeatureID)
			},
		},
		"test_remove_container_that_is_running_without_force": {
			ctrStart:          true,
			ctrHasToBeRemoved: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				suite.executeWithExpectedError(ctrFeatureID, operationRemove, false)
			},
		},
		"test_stop_container_with_options": {
			ctrStart:          true,
			ctrHasToBeRemoved: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				params := map[string]string{"signal": "SIGINT"}
				suite.executeWithExpectedSuccess(ctrFeatureID, operationStopWithOptions, params)
				suite.processStateChange(wsConnection, ctrFeatureID, ctrStatusStopped)
			},
		},
		"test_update_container": {
			ctrHasToBeRemoved: true,
			exec: func(ctrFeatureID string, wsConnection *websocket.Conn) {
				restartPolicyKey := "restartPolicy"
				newRestartPolicy := map[string]interface{}{"type": "ALWAYS"}
				params := map[string]interface{}{restartPolicyKey: newRestartPolicy}
				suite.executeWithExpectedSuccess(ctrFeatureID, operationUpdate, params)
				suite.processUpdate(wsConnection, ctrFeatureID, restartPolicyKey, newRestartPolicy)
			},
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
			if event.Path != suite.constructStatusPath(ctrFeatureID, "state") {
				return true, fmt.Errorf(unexpectedContainerEventError + event.Path)
			}

			eventValue, err := parseMap(event.Value)
			require.NoError(suite.T(), err, fmt.Sprintf(parseError, "event value"))

			actualStatus, err := parseString(eventValue["status"])
			require.NoError(suite.T(), err, fmt.Sprintf(parseError, "property status"))

			if actualStatus != expectedStatus {
				return true, fmt.Errorf("expected container status - %s, got container status - %s", expectedStatus, actualStatus)
			}

			return true, nil
		}
		return false, fmt.Errorf(unknownMessageError, event.Topic.String())
	})
	require.NoError(suite.T(), err, fmt.Sprintf(messageProcessError, "updating the state of"))
}

func (suite *ctrInstanceSuite) processNameChange(wsConnection *websocket.Conn, ctrFeatureID string, expectedName string) {
	err := util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicModified {
			if event.Path != suite.constructStatusPath(ctrFeatureID, "name") {
				return true, fmt.Errorf(unexpectedContainerEventError)
			}

			actualName, err := parseString(event.Value)
			require.NoError(suite.T(), err, fmt.Sprintf(parseError, "property name"))

			if actualName != expectedName {
				return true, fmt.Errorf("expected container name - %s, got container name - %s", expectedName, actualName)
			}

			return true, nil
		}
		return false, fmt.Errorf(unknownMessageError, event.Topic.String())
	})
	require.NoError(suite.T(), err, fmt.Sprintf(messageProcessError, "updating the name of"))
}

func (suite *ctrInstanceSuite) processRemove(wsConnection *websocket.Conn, ctrFeatureID string) {
	err := util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicDeleted {
			if event.Path != fmt.Sprintf("/features/%s", ctrFeatureID) {
				return true, fmt.Errorf(unexpectedContainerEventError)
			}
			return true, nil
		} else if event.Topic.String() == suite.topicModified {
			// state change to DEAD and update of SoftwareUpdatable installedDependencies is expected before deleted
			if event.Path == suite.constructStatusPath(ctrFeatureID, "state") ||
				event.Path == suite.constructStatusPath(things.SoftwareUpdatableFeatureID, "installedDependencies") {
				return false, nil
			}
		}
		return false, fmt.Errorf(unknownMessageError, event.Topic.String())
	})
	require.NoError(suite.T(), err, fmt.Sprintf(messageProcessError, "removing"))
}

func (suite *ctrInstanceSuite) processUpdate(wsConnection *websocket.Conn, ctrFeatureID string, expectedKey string, expectedValue map[string]interface{}) {
	err := util.ProcessWSMessages(suite.Cfg, wsConnection, func(event *protocol.Envelope) (bool, error) {
		if event.Topic.String() == suite.topicModified {
			if event.Path != suite.constructStatusPath(ctrFeatureID, "config") {
				return true, fmt.Errorf(unexpectedContainerEventError)
			}

			eventValue, err := parseMap(event.Value)
			require.NoError(suite.T(), err, fmt.Sprintf(parseError, "event value"))

			actualValue, err := parseMap(eventValue[expectedKey])
			require.NoError(suite.T(), err, fmt.Sprintf(parseError, fmt.Sprintf("value of key \"%s\"", expectedKey)))

			if !reflect.DeepEqual(expectedValue, actualValue) {
				return true, fmt.Errorf("expected value - %s, got value - %s", expectedValue, actualValue)
			}

			return true, nil
		}
		return false, fmt.Errorf(unknownMessageError, event.Topic.String())
	})
	require.NoError(suite.T(), err, fmt.Sprintf(messageProcessError, "updating the configuration of"))
}
