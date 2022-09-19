// Copyright (c) 2021 Contributors to the Eclipse Foundation
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
	"context"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

func (su *softwareUpdatable) handleContainerEvents(ctx context.Context) {
	subscribeCtx, subscrCtxCancelFunc := context.WithCancel(ctx)
	su.cancelEventsHandler = subscrCtxCancelFunc

	eventsChannel, errorChannel := su.eventsMgr.Subscribe(subscribeCtx)
	go func(ctx context.Context) error {
	eventsLoop:
		for {
			select {
			case ctrEvent := <-eventsChannel:
				if ctrEvent.Type == types.EventTypeContainers {
					switch ctrEvent.Action {
					case types.EventActionContainersCreated:
						su.addInstalledDependency(dependencyDescription(&ctrEvent.Source))
					case types.EventActionContainersRemoved:
						su.removeInstalledDependency(dependencyDescription(&ctrEvent.Source))
					default:
						log.Debug("an event that is not related to SoftwareUpdatable inventory has been received")
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
