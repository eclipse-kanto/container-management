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
	// "golang.org/x/net/websocket"
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
	metricsUrl  string
	firstCtrID  string
	secondCtrID string
}

func (suite *ctrMetricsSuite) SetupSuite() {
	suite.SetupCtrManagementSuite()
	suite.metricsUrl = util.GetFeatureURL(suite.ctrThingURL, metricsFeatureID)

	ctrParams := map[string]interface{}{
		paramImageRef: httpdImageRef,
		paramStart:    true,
	}

	suite.firstCtrID = suite.create(ctrParams)
	suite.secondCtrID = suite.create(ctrParams)

	suite.assertContainerMetricsFeatures()
}

func (suite *ctrMetricsSuite) TearDownSuite() {
	suite.stopMetricsRequest()
	suite.remove(suite.firstCtrID)
	suite.remove(suite.secondCtrID)
	suite.TearDown()
}

func TestContainerMetricsSuite(t *testing.T) {
	suite.Run(t, new(ctrMetricsSuite))
}

func (suite *ctrMetricsSuite) TestRequestMetricsForAllContainers() {
	params := make(map[string]interface{})
	params[paramFrequency] = "3s"

	err := suite.testMetrics(params, suite.firstCtrID, suite.secondCtrID)
	require.NoError(suite.T(), err, "error while receiving metrics for all container")
}

func (suite *ctrMetricsSuite) TestRequestMetricsForFirstContainer() {
	filter := things.Filter{}
	filter.ID = []string{"cpu.*", "memory.*", "io.*", "net.*", "pids"}
	filter.Originator = suite.firstCtrID
	params := make(map[string]interface{})
	params[paramFrequency] = "5s"
	params[paramFilter] = []things.Filter{filter}

	err := suite.testMetrics(params, suite.firstCtrID)
	require.NoErrorf(suite.T(), err, "error while receiving metrics for '%s' container", suite.firstCtrID)
}

func (suite *ctrMetricsSuite) TestFilterNotMatching() {
	params := make(map[string]interface{})
	params[paramFrequency] = "2s"
	filter := things.Filter{}
	filter.Originator = "test.process"
	params[paramFilter] = []things.Filter{filter}

	err := suite.testMetrics(params, filter.Originator)
	assert.Errorf(suite.T(), err,
		"metrics event for non existing originator '%s' should not be received", filter.Originator)

	filter.ID = []string{"test.io", "test.cpu", "test.memory", "test.net"}
	filter.Originator = suite.secondCtrID
	params[paramFrequency] = "1s"
	params[paramFilter] = []things.Filter{filter}

	err = suite.testMetrics(params, filter.Originator)
	assert.Error(suite.T(), err, "metrics event for non existing measurements test.* should not be received")
}

func (suite *ctrMetricsSuite) TestInvalidRequestMetrics() {
	params := make(map[string]interface{})
	params[paramFrequency] = "invalid frequency"
	_, err := util.ExecuteOperation(suite.Cfg, suite.metricsUrl, actionRequest, params)
	assert.Errorf(suite.T(), err, "error while sending metrics request with invalid params %v", params)

	params[paramFrequency] = "2s"
	invalidAction := "invalidRequest"
	_, err = util.ExecuteOperation(suite.Cfg, suite.metricsUrl, invalidAction, params)
	assert.Errorf(suite.T(), err, "error while sending metrics request with wrong topic %v", invalidAction)
}

func (suite *ctrMetricsSuite) assertContainerMetricsFeatures() {
	_, err := util.SendDigitalTwinRequest(suite.Cfg, http.MethodGet, suite.metricsUrl, nil)
	require.NoError(suite.T(), err, "error while getting the metrics feature")
}

func (suite *ctrMetricsSuite) stopMetricsRequest() {
	stopParams := map[string]interface{}{
		paramFrequency: "0s",
	}
	suite.executeMetrics(stopParams)
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

	pathData := util.GetFeatureOutboxMessagePath(metricsFeatureID, actionData)
	topicData := util.GetLiveMessageTopic(suite.ctrThingID, protocol.TopicAction(actionData))
	timestamp := time.Now().Unix()
	actualOriginators := make(map[string]bool)

	suite.executeMetrics(params)

	result := util.ProcessWSMessages(suite.Cfg, wsConnection, func(msg *protocol.Envelope) (bool, error) {
		if msg.Path != pathData {
			return false, nil
		}

		if msg.Topic.String() != topicData {
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
