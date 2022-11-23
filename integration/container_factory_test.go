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

	"github.com/eclipse-kanto/kanto/integration/util"
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
	ctrFeatureID string
}

var isCtrFeatureCreated bool

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
	params["imageRef"] = influxdbImageRef
	params["start"] = true

	suite.testCtrCreated("create", params)
}

func (suite *ctrFactorySuite) TestCreateWithConfig() {
	params := make(map[string]interface{})
	params["imageRef"] = influxdbImageRef
	params["start"] = true
	params["config"] = make(map[string]interface{})

	suite.testCtrCreated("createWithConfig", params)
}

func (suite *ctrFactorySuite) TestCreateWithConfigPortMapping() {
	wsConnection := suite.createWSConnection()

	defer wsConnection.Close()

	config := make(map[string]interface{})
	config["extraHosts"] = []string{"ctrhost:host_ip"}
	config["portMappings"] = []map[string]interface{}{
		{
			"hostPort":      5000,
			"containerPort": 80,
		},
	}

	params := make(map[string]interface{})
	params["imageRef"] = httpdImageRef
	params["start"] = true
	params["config"] = config

	util.ExecuteOperation(suite.Cfg, suite.ctrFactoryFeatureURL, "createWithConfig", params)

	err := util.ProcessWSMessages(suite.Cfg, wsConnection, suite.processCtrFeatureCreated)
	require.NoError(suite.T(), err, "failed to reach requested URL on host from the running container")

	defer suite.removeCtrFeature(wsConnection, suite.ctrFeatureID)

	body, err := sendHTTPGetRequest()
	require.Equal(suite.T(), httpdResponse, string(body), "HTTP response from the running container is not expected")
}

func (suite *ctrFactorySuite) processCtrFeatureCreated(event *protocol.Envelope) (bool, error) {
	if event.Topic.String() == suite.topicCreated {
		suite.ctrFeatureID = getCtrFeatureID(event.Path)
		return false, nil
	}
	if event.Topic.String() == suite.topicModified {
		if suite.ctrFeatureID == "" {
			return true, fmt.Errorf("event for creating container feature is not received")
		}
		status, check := event.Value.(map[string]interface{})
		if !check {
			return true, fmt.Errorf("error while parsing the property status value from the received event")
		}
		if status["status"].(string) == statusCreated {
			isCtrFeatureCreated = true
			return false, nil
		}
		if isCtrFeatureCreated && status["status"].(string) == statusRunning {
			return true, nil
		}
		return true, fmt.Errorf("event for modify container feature status is not received")
	}
	return false, fmt.Errorf("events for creating container feature are not received")
}

func (suite *ctrFactorySuite) testCtrCreated(operation string, params interface{}) {
	wsConnection := suite.createWSConnection()

	defer wsConnection.Close()

	util.ExecuteOperation(suite.Cfg, suite.ctrFactoryFeatureURL, operation, params)

	err := util.ProcessWSMessages(suite.Cfg, wsConnection, suite.processCtrFeatureCreated)
	require.NoError(suite.T(), err, "error while creating container feature")

	defer suite.removeCtrFeature(wsConnection, suite.ctrFeatureID)

	require.Equal(suite.T(), statusRunning, suite.getActualCtrStatus(suite.ctrFeatureID), "container status is not expected")
}

func sendHTTPGetRequest() ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, httpdRequestURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s %s request failed: %s", http.MethodGet, httpdRequestURL, resp.Status)
	}

	return io.ReadAll(resp.Body)
}
