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
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ctrFactorySuite struct {
	ctrManagementSuite
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
	params[paramImageRef] = influxdbImageRef
	params[paramStart] = true

	ctrFeatureID := suite.create(params)
	suite.remove(ctrFeatureID)
}

func (suite *ctrFactorySuite) TestCreateWithConfig() {
	params := make(map[string]interface{})
	params[paramImageRef] = influxdbImageRef
	params[paramStart] = true
	params[paramConfig] = make(map[string]interface{})

	ctrFeatureID := suite.createWithConfig(params)
	suite.remove(ctrFeatureID)
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
	params[paramImageRef] = httpdImageRef
	params[paramStart] = true
	params[paramConfig] = config

	ctrFeatureID := suite.createWithConfig(params)
	defer suite.remove(ctrFeatureID)
	suite.assertHTTPServer()
}

func (suite *ctrFactorySuite) assertHTTPServer() {
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1:5000", nil)
	require.NoError(suite.T(), err, "failed to create an HTTP request to the container")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(suite.T(), err, "failed to get an HTTP response from the container")

	defer resp.Body.Close()

	require.Equal(suite.T(), 200, resp.StatusCode, "HTTP response status code from the container is not expected")

	body, err := io.ReadAll(resp.Body)
	require.NoError(suite.T(), err, "failed to reach the requested URL on the host to the container")
	require.Equal(suite.T(), "<html><body><h1>It works!</h1></body></html>\n", string(body), "HTTP response from the container is not expected")
}
