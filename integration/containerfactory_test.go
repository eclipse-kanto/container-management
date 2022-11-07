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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	statusCreated               = "CREATED"
	statusRunning               = "RUNNING"
	requestURL                  = "http://127.0.0.1:5000"
	httpResponse                = "<html><body><h1>It works!</h1></body></html>\n"
	ctrFactoryFeatureDefinition = "com.bosch.iot.suite.edge.containers:ContainerFactory:1.2.0"
)

type ctrFactorySuite struct {
	containerManagementSuite
	ctrFeatureID string
}

func (suite *ctrFactorySuite) SetupSuite() {
	suite.connect()
	ctrFactoryFeature := suite.getCtrFeature(ctrFactoryFeatureID)
	require.NotNil(suite.T(), ctrFactoryFeature, "ContainerFactory feature must not be nil.")

	ctrFactoryFeatureDef := ctrFactoryFeature.GetDefinition()
	require.NotNil(suite.T(), ctrFactoryFeatureDef, "ContainerFactory feature definition must not bi nil.")
	require.Equal(suite.T(), ctrFactoryFeatureDefinition, ctrFactoryFeatureDef[0].String(), "ContainerFactory feature definition is not equals as expected.")
}

func (suite *ctrFactorySuite) TearDownSuite() {
	suite.disconnect()
}

func TestContainerFactorySuite(t *testing.T) {
	suite.Run(t, new(ctrFactorySuite))
}

func (suite *ctrFactorySuite) TestCreateOperation() {
	chEvent := suite.isCreated()

	params := make(map[string]interface{})
	params["imageRef"] = "docker.io/library/influxdb:1.8.4"
	params["start"] = true

	suite.execCreateCommand("create", params)

	require.True(suite.T(), suite.awaitChan(chEvent), "The created events are not received.")
	require.Equal(suite.T(), statusRunning, suite.getActualCtrState(), "The container state is not expected.")

	chEvent = suite.isDeleted()
	suite.execRemoveCommand(suite.ctrFeatureID)
	require.True(suite.T(), suite.awaitChan(chEvent), "The deleted event not received.")
}

func (suite *ctrFactorySuite) TestCreateWithConfigOperation() {
	chEvent := suite.isCreated()

	params := make(map[string]interface{})
	params["imageRef"] = "docker.io/library/influxdb:1.8.4"
	params["start"] = true
	params["config"] = make(map[string]interface{})

	suite.execCreateCommand("createWithConfig", params)

	require.True(suite.T(), suite.awaitChan(chEvent), "The event not received.")
	require.Equal(suite.T(), statusRunning, suite.getActualCtrState(), "The container state is not expected.")

	chEvent = suite.isDeleted()
	suite.execRemoveCommand(suite.ctrFeatureID)
	require.True(suite.T(), suite.awaitChan(chEvent), "The deleted event not received.")
}

func (suite *ctrFactorySuite) TestCreateWithConfigPortMapping() {
	chEvent := suite.isCreated()

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

	require.True(suite.T(), suite.awaitChan(chEvent), "The event not received.")

	data, _ := suite.doRequest(http.MethodGet, requestURL, nil)
	require.Equal(suite.T(), httpResponse, string(data), "The HTTP response is not expected.")

	chEvent = suite.isDeleted()
	suite.execRemoveCommand(suite.ctrFeatureID)
	require.True(suite.T(), suite.awaitChan(chEvent), "The deleted event not received.")
}

func (suite *ctrFactorySuite) isCreated() chan bool {
	return suite.startEventListener("START-SEND-EVENTS", "/features/Container:*", func(props map[string]interface{}) bool {
		if props["topic"].(string) == suite.topicCreated {
			suite.ctrFeatureID = getCtrFeatureID(props["path"].(string))
			return false
		}
		if props["topic"].(string) == suite.topicModify {
			if suite.ctrFeatureID == "" {
				return false
			}
			if value, ok := props["value"]; ok {
				status, check := value.(map[string]interface{})
				if status["status"].(string) == statusCreated {
					return false
				}
				return check && status["status"].(string) == statusRunning
			}
		}
		return false
	})
}

func (suite *ctrFactorySuite) isDeleted() chan bool {
	filter := fmt.Sprintf("/features/%s", suite.ctrFeatureID)
	return suite.startEventListener("START-SEND-EVENTS", filter, func(props map[string]interface{}) bool {
		return props["topic"].(string) == suite.topicDeleted
	})
}

func (suite *ctrFactorySuite) getActualCtrState() string {
	ctrPropertyPath := fmt.Sprintf("%s/features/%s/properties/status/state/status", suite.ctrThingURL, suite.ctrFeatureID)
	data, _ := suite.doRequest(http.MethodGet, ctrPropertyPath, nil)
	return strings.Trim(string(data), "\"")
}
