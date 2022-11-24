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
	"io"
	"net/http"
	"testing"

	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	httpdRequestURL = "http://127.0.0.1:5000"
	httpdResponse   = "<html><body><h1>It works!</h1></body></html>\n"
)

type ctrFactorySuite struct {
	ctrManagementSuite
	ctrFeatureID        string
	isCtrFeatureCreated bool
}

func (suite *ctrFactorySuite) SetupSuite() {
	suite.SetupCtrManagementSuite()
}

func (suite *ctrFactorySuite) TearDownSuite() {
	suite.TearDown()
}

func TestCtrFactorySuite(t *testing.T) {
	suite.Run(t, new(ctrFactorySuite))
}

func (suite *ctrFactorySuite) TestCreate() {
	params := make(map[string]interface{})
	params[imageRefParam] = influxdbImageRef
	params[startParam] = true

	wsConnection := suite.testCreate(createOperation, params, suite.processCtrFeatureCreated)
	defer wsConnection.Close()

	defer suite.testRemove(wsConnection, suite.ctrFeatureID)

	require.Equal(suite.T(), statusRunning, suite.getActualCtrStatus(suite.ctrFeatureID), "container status is not expected")
}

func (suite *ctrFactorySuite) TestCreateWithConfig() {
	params := make(map[string]interface{})
	params[imageRefParam] = influxdbImageRef
	params[startParam] = true
	params[configParam] = make(map[string]interface{})

	wsConnection := suite.testCreate(createWithConfigOperation, params, suite.processCtrFeatureCreated)
	defer wsConnection.Close()

	defer suite.testRemove(wsConnection, suite.ctrFeatureID)

	require.Equal(suite.T(), statusRunning, suite.getActualCtrStatus(suite.ctrFeatureID), "container status is not expected")
}

func (suite *ctrFactorySuite) TestCreateWithConfigPortMapping() {
	config := make(map[string]interface{})
	config["extraHosts"] = []string{"ctrhost:host_ip"}
	config["portMappings"] = []map[string]interface{}{
		{
			"hostPort":      5000,
			"containerPort": 80,
		},
	}

	params := make(map[string]interface{})
	params[imageRefParam] = httpdImageRef
	params[startParam] = true
	params[configParam] = config

	wsConnection := suite.testCreate(createWithConfigOperation, params, suite.processCtrFeatureCreated)
	defer wsConnection.Close()

	defer suite.testRemove(wsConnection, suite.ctrFeatureID)

	suite.assertHTTPServer()
}

func (suite *ctrFactorySuite) processCtrFeatureCreated(event *protocol.Envelope) (bool, error) {
	if event.Topic.String() == suite.topicCreated {
		suite.ctrFeatureID = getCtrFeatureID(event.Path)
		return false, nil
	}
	if event.Topic.String() == suite.topicModified {
		if suite.ctrFeatureID == "" {
			return true, fmt.Errorf("event for creating the container feature is not received")
		}
		status, check := event.Value.(map[string]interface{})
		if !check {
			return true, fmt.Errorf("failed to parsing the property status value from the received event")
		}
		if status[ctrStatusProperty].(string) == statusCreated {
			suite.isCtrFeatureCreated = true
			return false, nil
		}
		if suite.isCtrFeatureCreated && status[ctrStatusProperty].(string) == statusRunning {
			return true, nil
		}
		return true, fmt.Errorf("event for modify the container feature status is not received")
	}
	return false, fmt.Errorf("unknown message received")
}

func (suite *ctrFactorySuite) assertHTTPServer() {
	req, err := http.NewRequest(http.MethodGet, httpdRequestURL, nil)
	require.NoError(suite.T(), err, "failed to create an HTTP request from the container")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(suite.T(), err, "failed to get an HTTP response from the container")

	defer resp.Body.Close()

	require.Equal(suite.T(), 200, resp.StatusCode, "HTTP response status code from the container is not expected")

	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err, "failed to reach the requested URL on the host from the container")
	require.Equal(suite.T(), httpdResponse, string(body), "HTTP response from the container is not expected")
}
