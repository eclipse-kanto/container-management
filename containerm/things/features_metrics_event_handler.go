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
	"fmt"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

func (f *metricsFeature) handleContainerEvents(ctx context.Context) {
	subscribeCtx, subscrCtxCancelFunc := context.WithCancel(ctx)
	f.cancelEventsHandler = subscrCtxCancelFunc
	eventsChannel, errorChannel := f.eventsMgr.Subscribe(subscribeCtx)
	go func(ctx context.Context) error {
	eventsLoop:
		for {
			select {
			case ctrEvent := <-eventsChannel:
				if ctrEvent.Type == types.EventTypeContainers {
					ctrID := ctrEvent.Source.ID
					switch ctrEvent.Action {
					case types.EventActionContainersRunning:
						f.handleEventRunning(ctx, ctrID)
					default:
						log.Debug("container changed event received that does not affect the Metrics feature")
					}
				}
			case err := <-errorChannel:
				log.ErrorErr(err, "received Error from subscription")
			case <-ctx.Done():
				log.Debug("subscribe context is done - exiting subscribe events loop")
				break eventsLoop

			}
		}
		return nil
	}(subscribeCtx)
}

func (f *metricsFeature) handleEventRunning(ctx context.Context, ctrID string) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if f.disposed {
		return // feature is disposed do not report
	}

	originator := fmt.Sprintf(containerFeatureIDTemplate, ctrID)
	if f.request != nil && f.request.HasFilterForItem(CPUUtilization, originator) {
		metrics, err := f.mgr.Metrics(ctx, ctrID)
		if err != nil {
			log.ErrorErr(err, "could not get metrics for container with ID=%s", ctrID)
			return
		}

		if metrics != nil && metrics.CPU != nil {
			f.previousCPU[originator] = metrics.CPU
		}
	}
}
