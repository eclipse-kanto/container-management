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

package events

import (
	"context"
	"sync"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

//EventsManagerServiceLocalID represents the ID of the local events manager service
const EventsManagerServiceLocalID = "container-management.service.local.v1.service-events-manager"

func init() {
	registry.Register(&registry.Registration{
		ID:       EventsManagerServiceLocalID,
		Type:     registry.EventsManagerService,
		InitFunc: registryInit,
	})
}

type eventsMgr struct {
	broadcaster  *eventsSinkDispatcher
	publishMutex sync.Mutex
}

func (eMgr *eventsMgr) Publish(ctx context.Context, eventType types.EventType, eventAction types.EventAction, source *types.Container) error {
	eMgr.publishMutex.Lock()
	defer eMgr.publishMutex.Unlock()

	if source.State == nil {
		return log.NewErrorf("container info missing - cannot publish event")
	}
	msg := &types.Event{
		Type:   eventType,
		Action: eventAction,
		Source: util.CopyContainer(source),
		Time:   time.Now().UTC().Unix(),
	}
	err := eMgr.broadcaster.write(msg)
	if err != nil {
		log.ErrorErr(err, "could not publish event: %+v", msg)
	}
	log.Debug("published event %+v", msg)
	return err
}

func (eMgr *eventsMgr) Subscribe(ctx context.Context) (<-chan *types.Event, <-chan error) {
	var (
		eventsEmitter               = make(chan *types.Event)
		errorsEmitter               = make(chan error, 1)
		broadcasterChan             = newChannelledEventsSink(0)
		broadcasterQueue            = newQueueEventsSink(broadcasterChan)
		resultsSink      eventsSink = broadcasterQueue
	)

	clearResources := func() {
		close(errorsEmitter)
		eMgr.broadcaster.remove(resultsSink)
		broadcasterQueue.close()
		broadcasterChan.close()
	}

	eMgr.broadcaster.add(resultsSink)

	go func() {
		defer clearResources()

		var err error
	eventsLoop:
		for {
			select {
			case internalEvent := <-broadcasterChan.eventsChannel:
				event, ok := internalEvent.(*types.Event)
				if !ok {
					err = log.NewErrorf("invalid message received: %#v", internalEvent)
					log.DebugErr(err, "invalid message received")
					break
				}
				select {
				case eventsEmitter <- event:
					log.Debug("sent event to subscriber %+v", event)
				case <-ctx.Done():
					log.Debug("subscriber context is done")
					break eventsLoop
				}
			case <-ctx.Done():
				log.Debug("subscriber context is done")
				break eventsLoop
			}
		}
		if err == nil {
			if ctxErr := ctx.Err(); ctxErr != context.Canceled {
				log.DebugErr(ctxErr, "subscriber context has an error")
				err = ctxErr
			}
		}
		errorsEmitter <- err
	}()

	return eventsEmitter, errorsEmitter
}
