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
	"net/http"
	"strings"

	"github.com/eclipse-kanto/kanto/integration/util"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type ctrManagementSuite struct {
	suite.Suite
	util.SuiteInitializer
	ctrThingID           string
	ctrThingURL          string
	ctrFactoryFeatureURL string
	topicCreated         string
	topicModify          string
	topicDeleted         string
}

const (
	ctrFactoryFeatureID         = "ContainerFactory"
	ctrFactoryFeatureDefinition = "[\"com.bosch.iot.suite.edge.containers:ContainerFactory:1.2.0\"]"
)

func (suite *ctrManagementSuite) SetupCtrManagementSuite() {
	suite.Setup(suite.T())

	suite.ctrThingID = suite.ThingCfg.DeviceID + ":edge:containers"
	suite.ctrThingURL = util.GetThingURL(suite.Cfg.DigitalTwinAPIAddress, suite.ctrThingID)
	suite.ctrFactoryFeatureURL = util.GetFeatureURL(suite.ctrThingURL, ctrFactoryFeatureID)

	suite.topicCreated = util.GetTwinEventTopic(suite.ctrThingID, protocol.ActionCreated)
	suite.topicModify = util.GetTwinEventTopic(suite.ctrThingID, protocol.ActionModified)
	suite.topicDeleted = util.GetTwinEventTopic(suite.ctrThingID, protocol.ActionDeleted)

	suite.assertCtrFactoryFeature()
}

func getCtrFeatureID(topic string) string {
	result := strings.Split(topic, "/")
	return result[2]
}

func (suite *ctrManagementSuite) getActualCtrStatus(ctrFeatureID string) string {
	ctrPropertyPath := fmt.Sprintf("%s/features/%s/properties/status/state/status", suite.ctrThingURL, ctrFeatureID)
	body, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, ctrPropertyPath, nil)
	require.NoError(suite.T(), err, "error while getting the status property of the container feature: %s", ctrFeatureID)

	return strings.Trim(string(body), "\"")
}

func (suite *ctrManagementSuite) assertCtrFactoryFeature() {
	body, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, suite.ctrFactoryFeatureURL, nil)
	require.NoError(suite.T(), err, "error while getting the container factory feature")

	ctrFactoryDefinition := fmt.Sprintf("%s/definition", suite.ctrFactoryFeatureURL)
	body, err = util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, ctrFactoryDefinition, nil)
	require.NoError(suite.T(), err, "error while getting the container factory feature definition")

	require.Equal(suite.T(), ctrFactoryFeatureDefinition, string(body), "container factory definition is not expected")
}
