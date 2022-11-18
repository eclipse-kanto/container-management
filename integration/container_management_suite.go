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

	"github.com/eclipse-kanto/kanto/integration/util"
	"github.com/eclipse/ditto-clients-golang/model"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/websocket"
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
	ctrFactoryFeatureID = "ContainerFactory"
)

func (suite *ctrManagementSuite) SetupCommonSuite() {
	suite.Setup(suite.T())

	thingCfg, err := util.GetThingConfiguration(suite.Cfg, suite.MQTTClient)
	require.NoError(suite.T(), err, "failed to get thing configuration")

	suite.ctrThingID = thingCfg.DeviceID + ":edge:containers"
	suite.ctrThingURL = fmt.Sprintf("%s/api/2/things/%s", strings.TrimSuffix(suite.Cfg.DigitalTwinAPIAddress, "/"), suite.ctrThingID)
	suite.ctrFactoryFeatureURL = fmt.Sprintf("%s/features/%s", suite.ctrThingURL, ctrFactoryFeatureID)

	ctrThingID := model.NewNamespacedIDFrom(suite.ctrThingID)
	eventTopic := (&protocol.Topic{}).
		WithNamespace(ctrThingID.Namespace).
		WithEntityName(ctrThingID.Name).
		WithGroup(protocol.GroupThings).
		WithChannel(protocol.ChannelTwin).
		WithCriterion(protocol.CriterionEvents)
	suite.topicCreated = eventTopic.WithAction(protocol.ActionCreated).String()
	suite.topicModify = eventTopic.WithAction(protocol.ActionModified).String()
	suite.topicDeleted = eventTopic.WithAction(protocol.ActionDeleted).String()
}

func getCtrFeatureID(topic string) string {
	result := strings.Split(topic, "/")
	return result[2]
}

func (suite *ctrManagementSuite) create(params map[string]interface{}) {
	url := fmt.Sprintf("%s/inbox/messages/%s", suite.ctrFactoryFeatureURL, "create")
	suite.execCommand(url, params)
}

func (suite *ctrManagementSuite) createWithConfig(params map[string]interface{}) {
	url := fmt.Sprintf("%s/inbox/messages/%s", suite.ctrFactoryFeatureURL, "createWithConfig")
	suite.execCommand(url, params)
}

func (suite *ctrManagementSuite) remove(ctrFeatureID string) {
	ctrFeatureURL := fmt.Sprintf("%s/features/%s", suite.ctrThingURL, ctrFeatureID)
	url := fmt.Sprintf("%s/inbox/messages/remove", ctrFeatureURL)
	suite.execCommand(url, true)
}

func (suite *ctrManagementSuite) execCommand(url string, params interface{}) {
	_, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodPost, url, params)
	require.NoError(suite.T(), err, "error while creating container feature")
}

func (suite *ctrManagementSuite) startListening(conn *websocket.Conn, eventType, filter string) {
	err := websocket.Message.Send(conn, fmt.Sprintf("%s?filter=like(resource:path,'%s')", eventType, filter))
	require.NoError(suite.T(), err, "error sending listener request")
	err = util.WaitForWSMessage(suite.Cfg, conn, fmt.Sprintf("%s:ACK", eventType))
	require.NoError(suite.T(), err, "acknowledgement not received in time")
}
