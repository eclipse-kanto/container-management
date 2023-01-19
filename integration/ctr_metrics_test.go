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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/things"
	"github.com/eclipse-kanto/kanto/integration/util"
	"github.com/eclipse/ditto-clients-golang/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	metricsFeatureID = "Metrics"
	actionRequest    = "request"
	actionData       = "data"
	paramFrequency   = "frequency"
	paramFilter      = "filter"
)

type ctrMetricsSuite struct {
	ctrManagementSuite
	metricsUrl         string
	firstCtrFeatureID  string
	secondCtrFeatureID string
}

func (suite *ctrMetricsSuite) SetupSuite() {
	suite.SetupCtrManagementSuite()
	suite.metricsUrl = util.GetFeatureURL(suite.ctrThingURL, metricsFeatureID)

	suite.assertContainerMetricsFeatures()

	ctrParams := map[string]interface{}{
		paramImageRef: httpdImageRef,
		paramStart:    true,
	}

	suite.firstCtrFeatureID = suite.create(ctrParams)
	suite.secondCtrFeatureID = suite.create(ctrParams)
}

func (suite *ctrMetricsSuite) TearDownSuite() {
	suite.remove(suite.firstCtrFeatureID)
	suite.remove(suite.secondCtrFeatureID)
	suite.TearDown()
}

func TestContainerMetricsSuite(t *testing.T) {
	suite.Run(t, new(ctrMetricsSuite))
}

func (suite *ctrMetricsSuite) TestRequestMetricsForAllContainers() {
	err := suite.testMetrics(map[string]interface{}{paramFrequency: "3s"}, suite.firstCtrFeatureID, suite.secondCtrFeatureID)
	require.NoError(suite.T(), err, "error while receiving metrics for all container")
}

func (suite *ctrMetricsSuite) TestRequestMetricsForFirstContainer() {
	filter := things.Filter{
		ID:         []string{"cpu.*", "memory.*", "io.*", "net.*", "pids"},
		Originator: suite.firstCtrFeatureID,
	}

	err := suite.testMetrics(createParams("5s", filter), suite.firstCtrFeatureID)
	require.NoErrorf(suite.T(), err, "error while receiving metrics for '%s' container", suite.firstCtrFeatureID)
}

func (suite *ctrMetricsSuite) TestFilterNotMatching() {
	filter := things.Filter{Originator: "test.process"}

	err := suite.testMetrics(createParams("2s", filter), filter.Originator)
	assert.Errorf(suite.T(), err,
		"metrics event for non existing originator '%s' should not be received", filter.Originator)

	filter.ID = []string{"test.io", "test.cpu", "test.memory", "test.net"}
	filter.Originator = suite.secondCtrFeatureID

	err = suite.testMetrics(createParams("2s", filter), filter.Originator)
	assert.Error(suite.T(), err, "metrics event for non existing measurements test.* should not be received")
}

func (suite *ctrMetricsSuite) TestInvalidFrequency() {
	params := make(map[string]interface{})
	params[paramFrequency] = "invalid frequency"
	_, err := util.ExecuteOperation(suite.Cfg, suite.metricsUrl, actionRequest, params)
	assert.Errorf(suite.T(), err, "error while sending metrics request with invalid params %v", params)
}

func (suite *ctrMetricsSuite) TestInvalidAction() {
	invalidAction := "invalidRequest"
	_, err := util.ExecuteOperation(suite.Cfg, suite.metricsUrl, invalidAction,
		map[string]interface{}{paramFrequency: "2s"})
	assert.Errorf(suite.T(), err, "error while sending metrics request with wrong topic %s", invalidAction)
}

func (suite *ctrMetricsSuite) assertContainerMetricsFeatures() {
	_, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, suite.metricsUrl, nil)
	require.NoError(suite.T(), err, "error while getting the metrics feature")

	suite.assertCtrFeatureDefinition(suite.metricsUrl, "[\"com.bosch.iot.suite.edge.metric:Metrics:1.0.0\"]")
}

func (suite *ctrMetricsSuite) stopMetricsRequest() {
	suite.executeMetrics(map[string]interface{}{paramFrequency: "0s"})
}

func (suite *ctrMetricsSuite) executeMetrics(params map[string]interface{}) {
	_, err := util.ExecuteOperation(suite.Cfg, suite.metricsUrl, actionRequest, params)
	assert.NoErrorf(suite.T(), err, "error while sending metrics request for containers with params %v", params)
}

func (suite *ctrMetricsSuite) testMetrics(params map[string]interface{}, expectedOriginators ...string) error {
	wsConnection, err := util.NewDigitalTwinWSConnection(suite.Cfg)
	defer wsConnection.Close()

	require.NoError(suite.T(), err, "failed to create websocket connection")

	err = util.SubscribeForWSMessages(suite.Cfg, wsConnection, util.StartSendMessages, "")
	defer func() {
		suite.stopMetricsRequest()
		util.UnsubscribeFromWSMessages(suite.Cfg, wsConnection, util.StopSendMessages)
	}()
	require.NoError(suite.T(), err, "unable to listen for events by using a websocket connection")

	timestamp := time.Now().Unix()
	actualOriginators := make(map[string]bool)

	suite.executeMetrics(params)

	result := util.ProcessWSMessages(suite.Cfg, wsConnection, func(msg *protocol.Envelope) (bool, error) {
		if msg.Path != util.GetFeatureOutboxMessagePath(metricsFeatureID, actionData) {
			return false, nil
		}

		if msg.Topic.String() != util.GetLiveMessageTopic(suite.ctrThingID, protocol.TopicAction(actionData)) {
			return false, nil
		}

		data, err := json.Marshal(msg.Value)
		if err != nil {
			return true, err
		}

		metric := new(things.MetricData)
		if err := json.Unmarshal(data, metric); err != nil {
			return true, err
		}

		if metric.Timestamp < timestamp {
			return true, fmt.Errorf("Invalid timestamp: %v", metric.Timestamp)
		}

		for _, m := range metric.Snapshot {
			for _, originator := range expectedOriginators {
				if originator == m.Originator {
					actualOriginators[originator] = true
					break
				}
			}

			if _, ok := actualOriginators[m.Originator]; !ok {
				return true, fmt.Errorf("Invalid originator: %s", m.Originator)
			}

			for _, mm := range m.Measurements {
				if !allowedPrefixID(mm.ID) {
					return true, fmt.Errorf("Invalid metrics ID: %s", mm.ID)
				}
			}
		}

		return len(expectedOriginators) == len(actualOriginators), nil
	})

	return result
}

func allowedPrefixID(id string) bool {
	allowedIDs := []string{"cpu.", "memory.", "io.", "net.", "pids"}
	for _, allowedID := range allowedIDs {
		if strings.HasPrefix(id, allowedID) {
			return true
		}
	}
	return false
}

func createParams(frequency string, filter things.Filter) map[string]interface{} {
	return map[string]interface{}{
		paramFrequency: frequency,
		paramFilter:    []things.Filter{filter},
	}
}
