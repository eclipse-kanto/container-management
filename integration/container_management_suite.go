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
	"github.com/eclipse/ditto-clients-golang/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/websocket"
)

type containerManagementSuite struct {
	suite.Suite
	suiteInit            util.SuiteInitializer
	ctrThingID           string
	ctrThingURL          string
	ctrFactoryFeatureURL string
	topicCreated         string
	topicModify          string
	topicDeleted         string
}

type containerFeature struct {
	Definition []string               `json:"definition,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

const (
	ctrFactoryFeatureID = "ContainerFactory"
)

func (suite *containerManagementSuite) init() {
	suite.suiteInit = util.SuiteInitializer{}
	suite.suiteInit.Setup(suite.T())

	edgeDeviceCfg, err := util.GetThingConfiguration(suite.suiteInit.Cfg, suite.suiteInit.MQTTClient)
	require.NoError(suite.T(), err, "failed to get thing configuration")

	suite.ctrThingID = edgeDeviceCfg.DeviceID + ":edge:containers"
	suite.ctrThingURL = fmt.Sprintf("%s/api/2/things/%s", strings.TrimSuffix(suite.suiteInit.Cfg.DigitalTwinAPIAddress, "/"), suite.ctrThingID)
	suite.ctrFactoryFeatureURL = fmt.Sprintf("%s/features/%s", suite.ctrThingURL, ctrFactoryFeatureID)
	ctrThingID := model.NewNamespacedIDFrom(suite.ctrThingID)
	suite.topicCreated = fmt.Sprintf("%s/%s/things/twin/events/created", ctrThingID.Namespace, ctrThingID.Name)
	suite.topicModify = fmt.Sprintf("%s/%s/things/twin/events/modified", ctrThingID.Namespace, ctrThingID.Name)
	suite.topicDeleted = fmt.Sprintf("%s/%s/things/twin/events/deleted", ctrThingID.Namespace, ctrThingID.Name)
}

func getCtrFeatureID(topic string) string {
	result := strings.Split(topic, "/")
	return result[2]
}

func (suite *containerManagementSuite) execCreateCommand(command string, params map[string]interface{}) {
	url := fmt.Sprintf("%s/inbox/messages/%s", suite.ctrFactoryFeatureURL, command)

	_, err := util.SendDigitalTwinRequest(suite.suiteInit.Cfg, http.MethodPost, url, params)
	require.NoError(suite.T(), err, "error while creating container feature")
}

func (suite *containerManagementSuite) execRemoveCommand(ctrFeatureID string) {
	ctrFeatureURL := fmt.Sprintf("%s/features/%s", suite.ctrThingURL, ctrFeatureID)
	url := fmt.Sprintf("%s/inbox/messages/remove", ctrFeatureURL)
	_, err := util.SendDigitalTwinRequest(suite.suiteInit.Cfg, http.MethodPost, url, true)
	require.NoError(suite.T(), err, "error while removing container feature")
}

func (suite *containerManagementSuite) startListening(conn *websocket.Conn, eventType, filter string) {
	err := websocket.Message.Send(conn, fmt.Sprintf("%s?filter=like(resource:path,'%s')", eventType, filter))
	require.NoError(suite.T(), err, "error sending listener request")
	err = util.WaitForWSMessage(suite.suiteInit.Cfg, conn, fmt.Sprintf("%s:ACK", eventType))
	require.NoError(suite.T(), err, "acknowledgement not received in time")
}
