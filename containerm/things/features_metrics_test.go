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

package things

import (
	"fmt"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/things/client"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

var testOriginator = fmt.Sprintf(containerFeatureIDTemplate, testContainerID)

func TestFeatureOperationsHandlerRequestError(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	setupManagerMock(controller)
	setupEventsManagerMock(controller)
	setupThingMock(controller)

	testMetrics := newMetricsFeature(mockThing, mockContainerManager, mockEventsManager)
	tests := map[string]struct {
		operation     string
		args          interface{}
		expectedError error
	}{
		"test_metrics_operations_handler_invalid_operation": {
			operation:     "invalid",
			expectedError: client.NewMessagesSubjectNotFound("unsupported operation invalid"),
		},
		"test_metrics_operations_handler_error_unmarshalling": {
			operation:     metricsFeatureOperationRequest,
			args:          testUnmarshableArg,
			expectedError: client.NewMessagesParameterInvalidError("json: cannot unmarshal string into Go value of type things.Request"),
		},
		"test_metrics_operations_handler_error_marshalling": {
			operation:     metricsFeatureOperationRequest,
			args:          testInvalidArg,
			expectedError: client.NewMessagesParameterInvalidError("json: unsupported type: chan int"),
		},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			result, resultErr := testMetrics.(*metricsFeature).featureOperationsHandler(testCase.operation, testCase.args)
			testutil.AssertEqual(t, testCase.expectedError, resultErr)
			testutil.AssertNil(t, result)
		})
	}
}

func TestFeatureOperationsHandlerRequest(t *testing.T) {
	const testEventsTimeout = 5 * time.Second

	controller := gomock.NewController(t)
	defer controller.Finish()

	setupManagerMock(controller)
	setupEventsManagerMock(controller)
	setupThingMock(controller)

	testMetrics := newMetricsFeature(mockThing, mockContainerManager, mockEventsManager)
	startMetricsReq := &Request{Frequency: Duration{time.Second}, Filter: []Filter{
		{ID: nil, Originator: testOriginator},
	}}
	stopMetricsReq := &Request{Frequency: Duration{0}}

	testWg := &sync.WaitGroup{}
	mockInitialMetrics()
	mockFirstMetricsReport(testWg)
	mockSecondMetricsReport(testWg)

	testWg.Add(2)
	result, resultErr := testMetrics.(*metricsFeature).featureOperationsHandler(metricsFeatureOperationRequest, startMetricsReq)
	defer func() {
		_, err := testMetrics.(*metricsFeature).featureOperationsHandler(metricsFeatureOperationRequest, stopMetricsReq)
		testutil.AssertNil(t, err)
	}()
	testutil.AssertNil(t, result)
	testutil.AssertNil(t, resultErr)
	testutil.AssertWithTimeout(t, testWg, testEventsTimeout)
}

func mockInitialMetrics() {
	initialMetrics := &types.Metrics{
		CPU: &types.CPUStats{
			Used:  10000,
			Total: 100000,
		},
	}
	mockContainerManager.EXPECT().List(gomock.Any()).Return([]*types.Container{testContainer}, nil).Times(1)
	mockContainerManager.EXPECT().Metrics(gomock.Any(), testContainerID).Return(initialMetrics, nil).Times(1)
}

func mockFirstMetricsReport(wg *sync.WaitGroup) {
	m := &types.Metrics{
		CPU: &types.CPUStats{
			Used:  15000,
			Total: 150000,
		},
		Memory: &types.MemoryStats{
			Used:  1024 * 1024 * 1024,
			Total: 8 * 1024 * 1024 * 1024,
		},
		IO: &types.IOStats{
			Read:  1024,
			Write: 1024,
		},
		Network: &types.IOStats{
			Read:  1024,
			Write: 1024,
		},
		PIDs: uint64(5),
	}
	mData := &MetricData{
		Snapshot: []OriginatorMeasurements{
			{Originator: testOriginator, Measurements: toMeasurements(m, float64(10))},
		},
	}
	mockContainerManager.EXPECT().List(gomock.Any()).Return([]*types.Container{testContainer}, nil).Times(1)
	mockContainerManager.EXPECT().Metrics(gomock.Any(), testContainerID).Return(m, nil).Times(1)
	mockThing.EXPECT().SendFeatureMessage(MetricsFeatureID, metricsFeatureAction, matchesMetricsData(mData)).Do(func(_, _ string, _ interface{}) { wg.Done() })
}

func mockSecondMetricsReport(wg *sync.WaitGroup) {
	m := &types.Metrics{
		CPU: &types.CPUStats{
			Used:  25000,
			Total: 200000,
		},
		Memory: &types.MemoryStats{
			Used:  2 * 1024 * 1024 * 1024,
			Total: 8 * 1024 * 1024 * 1024,
		},
		IO: &types.IOStats{
			Read:  2048,
			Write: 2048,
		},
		Network: &types.IOStats{
			Read:  2048,
			Write: 2048,
		},
		PIDs: uint64(8),
	}

	mData := &MetricData{
		Snapshot: []OriginatorMeasurements{
			{Originator: testOriginator, Measurements: toMeasurements(m, float64(20))},
		},
	}
	mockContainerManager.EXPECT().List(gomock.Any()).Return([]*types.Container{testContainer}, nil).Times(1)
	mockContainerManager.EXPECT().Metrics(gomock.Any(), testContainerID).Return(m, nil).Times(1)
	mockThing.EXPECT().SendFeatureMessage(MetricsFeatureID, metricsFeatureAction, matchesMetricsData(mData)).Do(func(_, _ string, _ interface{}) { wg.Done() })
}

func toMeasurements(m *types.Metrics, cpuPercentage float64) []Measurement {
	return []Measurement{
		{ID: CPUUtilization, Value: cpuPercentage},
		{ID: MemoryTotal, Value: float64(m.Memory.Total)},
		{ID: MemoryUsed, Value: float64(m.Memory.Used)},
		{ID: MemoryUtilization, Value: float64(m.Memory.Used) / float64(m.Memory.Total) * 100},
		{ID: IOReadBytes, Value: float64(m.IO.Read)},
		{ID: IOWriteBytes, Value: float64(m.IO.Write)},
		{ID: NetReadBytes, Value: float64(m.Network.Read)},
		{ID: NetWriteBytes, Value: float64(m.Network.Write)},
		{ID: PIDs, Value: float64(m.PIDs)},
	}
}

type metricsDataMatcher struct {
	data *MetricData
	msg  string
}

func matchesMetricsData(data *MetricData) gomock.Matcher {
	return &metricsDataMatcher{
		data: data,
	}
}

func (m *metricsDataMatcher) Matches(x interface{}) bool {
	switch x.(type) {
	case *MetricData:
		actual := x.(*MetricData)
		// do not match Timestamp
		m.data.Timestamp = actual.Timestamp
		m.msg = fmt.Sprintf("expected %+v , got %+v", m.data, actual)
		return reflect.DeepEqual(m.data, actual)
	default:
		return false
	}
}

func (m *metricsDataMatcher) String() string {
	return m.msg
}
