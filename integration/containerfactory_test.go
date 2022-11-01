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
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	statusCreated = "CREATED"
	statusRunning = "RUNNING"
	requestURL    = "http://127.0.0.1:5000"
	httpResponse  = "<html><body><h1>It works!</h1></body></html>\n"
)

func (suite *containerManagementSuite) SetupSuite() {
	suite.newTestConnection()
}

func (suite *containerManagementSuite) TearDownSuite() {
	suite.disconnect()
}

func TestContainerFactorySuite(t *testing.T) {
	suite.Run(t, new(containerManagementSuite))
}

func (suite *containerManagementSuite) TestCreateOperation() {
	var containerID string
	chEvent := suite.startEventListener("START-SEND-EVENTS", func(props map[string]interface{}) bool {
		if props["topic"].(string) == suite.topicModify {
			containerID = getContainerID(props["path"].(string))
			if value, ok := props["value"]; ok {
				status, check := value.(map[string]interface{})
				return check && status["status"].(string) == statusCreated
			}
		}
		return false
	})

	params := make(map[string]interface{})
	params["imageRef"] = "docker.io/library/influxdb:1.8.4"
	params["start"] = false

	suite.execCreateCommand("create", params)

	require.True(suite.T(), suite.awaitChan(chEvent), "The event not received.")

	ctrFeture := suite.getContainerFeture(containerID)
	ctrStatusProp := ctrFeture.GetProperty("status").(map[string]interface{})
	require.NotNil(suite.T(), ctrStatusProp, "Container feture property 'status' is nil.")
	ctrStateProp := ctrStatusProp["state"].(map[string]interface{})
	require.Equal(suite.T(), ctrStateProp["status"], statusCreated, "The container state is not expected.")

	suite.execRemoveCommand(containerID)
}

func (suite *containerManagementSuite) TestCreateWithConfigOperation() {
	var containerID string
	chEvent := suite.startEventListener("START-SEND-EVENTS", func(props map[string]interface{}) bool {
		if props["topic"].(string) == suite.topicModify {
			containerID = getContainerID(props["path"].(string))
			if value, ok := props["value"]; ok {
				status, check := value.(map[string]interface{})
				return check && status["status"].(string) == statusRunning
			}
		}
		return false
	})
	config := make(map[string]interface{})
	config["extraHosts"] = []string{"ctrhost:host_ip"}
	config["portMappings"] = []map[string]interface{}{
		{
			"hostPort":      5000,
			"hostPortEnd":   5000,
			"containerPort": 80,
		},
	}

	params := make(map[string]interface{})
	params["imageRef"] = "docker.io/library/httpd:latest"
	params["start"] = true
	params["config"] = config

	suite.execCreateCommand("createWithConfig", params)

	require.True(suite.T(), suite.awaitChan(chEvent), "The event not received.")

	data, _ := suite.doRequest(http.MethodGet, requestURL, nil)
	require.Equal(suite.T(), httpResponse, string(data), "The HTTP response is not expected.")

	suite.execRemoveCommand(containerID)
}
