// Copyright (c) 2021 Contributors to the Eclipse Foundation
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

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

func (ctrFactory *containerFactoryFeature) handleContainerEvents(ctx context.Context) {
	subscribeCtx, subscrCtxCancelFunc := context.WithCancel(ctx)
	ctrFactory.cancelEventsHandler = subscrCtxCancelFunc
	eventsChannel, errorChannel := ctrFactory.eventsMgr.Subscribe(subscribeCtx)
	go func(ctx context.Context) error {
	eventsLoop:
		for {
			select {
			case ctrEvent := <-eventsChannel:
				if ctrEvent.Type == types.EventTypeContainers {
					switch ctrEvent.Action {
					case types.EventActionContainersCreated:
						ctrFactory.handleEventCreated(ctrEvent)
						ctrFactory.handleStateChangedEvent(ctrEvent)
					case types.EventActionContainersRemoved:
						ctrFactory.handleEventRemoved(ctrEvent)
						ctrFactory.handleStateChangedEvent(ctrEvent)
					case types.EventActionContainersRenamed:
						ctrFactory.handleRenameEvent(ctrEvent)
					case types.EventActionContainersUpdated:
						ctrFactory.handleUpdateEvent(ctrEvent)
					default:
						log.Debug("container changed event received that does not affect the Container features set")
						ctrFactory.handleStateChangedEvent(ctrEvent)
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

func (ctrFactory *containerFactoryFeature) handleEventCreated(ctrEvent *types.Event) {
	ctrFactory.eventsHandlingLock.Lock()
	defer ctrFactory.eventsHandlingLock.Unlock()
	if err := ctrFactory.createContainerFeature(&ctrEvent.Source); err != nil {
		log.ErrorErr(err, "could not create feature for container with ID=%s", ctrEvent.Source.ID)
	}
	ctrFactory.storageMgr.StoreContainerInfo(ctrEvent.Source.ID)
}

func (ctrFactory *containerFactoryFeature) handleEventRemoved(ctrEvent *types.Event) {
	ctrFactory.eventsHandlingLock.Lock()
	defer ctrFactory.eventsHandlingLock.Unlock()
	if err := ctrFactory.removeContainerFeature(ctrEvent.Source.ID); err != nil {
		log.ErrorErr(err, "could not remove feature for container with ID=%s", ctrEvent.Source.ID)
	}
	ctrFactory.storageMgr.DeleteContainerInfo(ctrEvent.Source.ID)
}

func (ctrFactory *containerFactoryFeature) handleStateChangedEvent(ctrEvent *types.Event) {
	ctrFactory.eventsHandlingLock.Lock()
	defer ctrFactory.eventsHandlingLock.Unlock()
	err := ctrFactory.updateContainerFeature(ctrEvent.Source.ID, containerFeaturePropertyPathStatusState, fromAPIContainerState(ctrEvent.Source.State))
	if err != nil {
		log.ErrorErr(err, "could not update feature Status/State for container with ID=%s", ctrEvent.Source.ID)
	}
}

func (ctrFactory *containerFactoryFeature) handleRenameEvent(ctrEvent *types.Event) {
	ctrFactory.eventsHandlingLock.Lock()
	defer ctrFactory.eventsHandlingLock.Unlock()

	if err := ctrFactory.updateContainerFeature(ctrEvent.Source.ID, containerFeaturePropertyPathStatusName, ctrEvent.Source.Name); err != nil {
		log.ErrorErr(err, "could not update feature Status/Name for container with ID=%s", ctrEvent.Source.ID)
	}
}

func (ctrFactory *containerFactoryFeature) handleUpdateEvent(ctrEvent *types.Event) {
	ctrFactory.eventsHandlingLock.Lock()
	defer ctrFactory.eventsHandlingLock.Unlock()
	err := ctrFactory.updateContainerFeature(ctrEvent.Source.ID, containerFeaturePropertyPathStatusConfig,
		fromAPIContainerConfig(&ctrEvent.Source))
	if err != nil {
		log.ErrorErr(err, "could not update feature Status/Config for container %+v", &ctrEvent.Source)
	}
}
