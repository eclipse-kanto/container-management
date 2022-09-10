// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package things

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/events"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
	"sync"
	"time"
)

const (
	// MetricsFeatureID is the feature ID of the container metrics
	MetricsFeatureID               = "Metrics"
	metricsFeatureOperationRequest = "request"
	metricsFeatureAction           = "data"
	metricsFeatureDefinition       = "com.bosch.iot.suite.edge.metric:Metrics:1.0.0"
)

type metricsFeature struct {
	rootThing           model.Thing
	mgr                 mgr.ContainerManager
	cancelEventsHandler context.CancelFunc
	eventsMgr           events.ContainerEventsManager
	previousCPU         map[string]*types.CPUStats
	request             *Request
	disposed            bool
	mutex               sync.Mutex
	ticker              *time.Ticker
	tickerStop          chan bool
}

func newMetricsFeature(rootThing model.Thing, mgr mgr.ContainerManager, eventsMgr events.ContainerEventsManager) managedFeature {
	return &metricsFeature{
		rootThing: rootThing,
		mgr:       mgr,
		eventsMgr: eventsMgr,
	}
}

func (f *metricsFeature) register(ctx context.Context) error {
	log.Debug("initializing Metrics feature")

	if f.cancelEventsHandler == nil {
		f.handleContainerEvents(ctx)
		log.Debug("subscribed for container events")
	}
	return f.rootThing.SetFeature(MetricsFeatureID, f.createFeature())
}

func (f *metricsFeature) dispose() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	log.Debug("disposing Metrics feature")
	f.stopTicker()
	f.disposed = true
	f.request = nil

	if f.cancelEventsHandler != nil {
		log.Debug("unsubscribing from container events")
		f.cancelEventsHandler()
		f.cancelEventsHandler = nil
	}
}

func (f *metricsFeature) featureOperationsHandler(operationName string, args interface{}) (interface{}, error) {
	if operationName == metricsFeatureOperationRequest {
		bytes, err := json.Marshal(args)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		req := new(Request)

		err = json.Unmarshal(bytes, req)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}

		return nil, f.processRequest(req)
	}

	err := log.NewErrorf("unsupported operation %s", operationName)
	log.ErrorErr(err, "unsupported operation %s", operationName)
	return nil, client.NewMessagesSubjectNotFound(err.Error())
}

func (f *metricsFeature) createFeature() model.Feature {
	return client.NewFeature(MetricsFeatureID,
		client.WithFeatureDefinitionFromString(metricsFeatureDefinition),
		client.WithFeatureOperationsHandler(f.featureOperationsHandler),
	)
}
func (f *metricsFeature) processRequest(req *Request) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.disposed {
		return log.NewError("metrics feature is disposed")
	}

	// frequency duration of zero means stop reports
	if req.Frequency.Duration <= 0 {
		// remove the previous ticker if there is such
		f.stopTicker()
		return nil
	}

	if req.Frequency.Duration < time.Second {
		req.Frequency.Duration = time.Second
	}
	f.request = req

	// initialize previous CPU stats to get proper CPU utilization measurement on first report
	f.previousCPU = make(map[string]*types.CPUStats)
	f.walkContainerMetrics(func(originator string, metrics *types.Metrics) {
		if metrics.CPU != nil && f.request.HasFilterForItem(CPUUtilization, originator) {
			f.previousCPU[originator] = metrics.CPU
		}
	})

	if f.ticker == nil {
		f.ticker = time.NewTicker(req.Frequency.Duration)
		f.tickerStop = make(chan bool)
		go func() {
			for {
				select {
				case <-f.tickerStop:
					return
				case <-f.ticker.C:
					f.reportMetrics()
				}
			}
		}()
		log.Info("started metrics with frequency = %s and filter = %v", f.request.Frequency, f.request.Filter)
	} else {
		f.ticker.Reset(req.Frequency.Duration)
		log.Info("reset metrics with frequency = %s and filter = %v", f.request.Frequency, f.request.Filter)
	}
	return nil
}

func (f *metricsFeature) walkContainerMetrics(execute func(string, *types.Metrics)) {
	ctrs, listErr := f.mgr.List(context.Background())
	if listErr != nil {
		log.ErrorErr(listErr, "could not list containers")
		return
	}
	for _, ctr := range ctrs {
		originator := fmt.Sprintf(containerFeatureIDTemplate, ctr.ID)

		if !f.request.HasFilterFor(originator) {
			continue
		}

		metrics, err := f.mgr.Metrics(context.Background(), ctr.ID)
		if err != nil {
			log.ErrorErr(err, "could not get metrics data for container ID = %s", ctr.ID)
			continue
		}
		if metrics == nil {
			log.Debug("no metrics data for container ID = %s", ctr.ID)
			continue
		}
		execute(originator, metrics)
	}
}

func (f *metricsFeature) reportMetrics() {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.disposed {
		return // feature is disposed do not report
	}

	mData := &MetricData{
		Snapshot:  make([]OriginatorMeasurements, 0),
		Timestamp: time.Now().Unix(),
	}

	f.walkContainerMetrics(func(originator string, metrics *types.Metrics) {

		var measurements []Measurement
		if metrics.CPU != nil && f.request.HasFilterForItem(CPUUtilization, originator) {
			if cpuUtilization, err := util.CalculateCPUPercent(metrics.CPU, f.previousCPU[originator]); err == nil {
				measurements = append(measurements, Measurement{ID: CPUUtilization, Value: cpuUtilization})
				f.previousCPU[originator] = metrics.CPU
			} else {
				log.DebugErr(err, "could not calculate CPU utilization for originator = %s", originator)
			}
		}

		if metrics.Memory != nil {
			if f.request.HasFilterForItem(MemoryTotal, originator) {
				measurements = append(measurements, Measurement{ID: MemoryTotal, Value: float64(metrics.Memory.Total)})
			}
			if f.request.HasFilterForItem(MemoryUsed, originator) {
				measurements = append(measurements, Measurement{ID: MemoryUsed, Value: float64(metrics.Memory.Used)})
			}
			if f.request.HasFilterForItem(MemoryUtilization, originator) {
				if memoryUtilization, err := util.CalculateMemoryPercent(metrics.Memory); err == nil {
					measurements = append(measurements, Measurement{ID: MemoryUtilization, Value: memoryUtilization})
				} else {
					log.DebugErr(err, "could not calculate memory utilization for originator = %s", originator)
				}
			}
		}

		if metrics.IO != nil {
			if f.request.HasFilterForItem(IOReadBytes, originator) {
				measurements = append(measurements, Measurement{ID: IOReadBytes, Value: float64(metrics.IO.Read)})
			}
			if f.request.HasFilterForItem(IOWriteBytes, originator) {
				measurements = append(measurements, Measurement{ID: IOWriteBytes, Value: float64(metrics.IO.Write)})
			}
		}

		if metrics.Network != nil {
			if f.request.HasFilterForItem(NetReadBytes, originator) {
				measurements = append(measurements, Measurement{ID: NetReadBytes, Value: float64(metrics.Network.Read)})
			}
			if f.request.HasFilterForItem(NetWriteBytes, originator) {
				measurements = append(measurements, Measurement{ID: NetWriteBytes, Value: float64(metrics.Network.Write)})
			}
		}

		if metrics.PIDs > 0 && f.request.HasFilterForItem(PIDs, originator) {
			measurements = append(measurements, Measurement{ID: PIDs, Value: float64(metrics.PIDs)})
		}

		if len(measurements) > 0 {
			m := OriginatorMeasurements{
				Originator:   originator,
				Measurements: measurements,
			}
			mData.Snapshot = append(mData.Snapshot, m)
		}
	})

	if len(mData.Snapshot) > 0 {
		log.Debug("sending metrics data = %v", mData)
		if err := f.rootThing.SendFeatureMessage(MetricsFeatureID, metricsFeatureAction, mData); err != nil {
			log.ErrorErr(err, "could not send metrics data = %v", mData)
		}

	}
}

func (f *metricsFeature) stopTicker() {
	if f.ticker != nil {
		f.tickerStop <- true
		f.ticker.Stop()
		f.ticker = nil
		log.Info("stopped metrics reporting")
	}
}
