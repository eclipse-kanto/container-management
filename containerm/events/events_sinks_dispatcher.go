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

package events

import (
	"fmt"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

var (
	errEventsSinkClosed = log.NewError("events: eSink closed")
)

// eventsSinkDispatcher sends events to multiple eventSinks
type eventsSinkDispatcher struct {
	eventsSinks    []eventsSink
	events         chan event
	addRequests    chan updateSinksRequest
	removeRequests chan updateSinksRequest

	shutdown chan struct{}
	closed   chan struct{}
	once     sync.Once
}

// newEventSinksDispatcher preconfigures the dispatcher with the provided eventsSinks
func newEventSinksDispatcher(sinks ...eventsSink) *eventsSinkDispatcher {
	dispatcher := eventsSinkDispatcher{
		eventsSinks:    sinks,
		events:         make(chan event),
		addRequests:    make(chan updateSinksRequest),
		removeRequests: make(chan updateSinksRequest),
		shutdown:       make(chan struct{}),
		closed:         make(chan struct{}),
	}

	// start processing
	go dispatcher.run()

	return &dispatcher
}

func (dispatcher *eventsSinkDispatcher) write(event event) error {
	select {
	case dispatcher.events <- event:
	case <-dispatcher.closed:
		return errEventsSinkClosed
	}
	return nil
}

// add the provided eventsSink to the eventsSinkDispatcher.
func (dispatcher *eventsSinkDispatcher) add(sink eventsSink) error {
	return dispatcher.processUpdateRequest(dispatcher.addRequests, sink)
}

// remove the provided eventsSink from the eventsSinkDispatcher.
func (dispatcher *eventsSinkDispatcher) remove(sink eventsSink) error {
	return dispatcher.processUpdateRequest(dispatcher.removeRequests, sink)
}

type updateSinksRequest struct {
	eSink    eventsSink
	response chan error
}

func (dispatcher *eventsSinkDispatcher) processUpdateRequest(ch chan updateSinksRequest, sink eventsSink) error {
	response := make(chan error, 1)
	for {
		select {
		case ch <- updateSinksRequest{
			eSink:    sink,
			response: response}:
			ch = nil
		case err := <-response:
			return err
		case <-dispatcher.closed:
			return errEventsSinkClosed
		}
	}
}
func (dispatcher *eventsSinkDispatcher) close() error {
	dispatcher.once.Do(func() {
		close(dispatcher.shutdown)
	})

	<-dispatcher.closed
	return nil
}
func (dispatcher *eventsSinkDispatcher) run() {
	defer close(dispatcher.closed)
	remove := func(target eventsSink) {
		for i, sink := range dispatcher.eventsSinks {
			if sink == target {
				dispatcher.eventsSinks = append(dispatcher.eventsSinks[:i], dispatcher.eventsSinks[i+1:]...)
				break
			}
		}
	}

	for {
		select {
		case event := <-dispatcher.events:
			for _, sink := range dispatcher.eventsSinks {
				if err := sink.write(event); err != nil {
					if err == errEventsSinkClosed {
						// remove closed eventsSinks
						remove(sink)
						continue
					}
					log.ErrorErr(err, "dropping event %+v for eventsSink %+v", event, sink)
				}
			}
		case request := <-dispatcher.addRequests:
			var found bool
			for _, sink := range dispatcher.eventsSinks {
				if request.eSink == sink {
					found = true
					break
				}
			}

			if !found {
				dispatcher.eventsSinks = append(dispatcher.eventsSinks, request.eSink)
			}
			// dispatcher.eventsSinks[request.eSink] = struct{}{}
			request.response <- nil
		case request := <-dispatcher.removeRequests:
			remove(request.eSink)
			request.response <- nil
		case <-dispatcher.shutdown:
			// close all the underlying eventsSinks
			for _, sink := range dispatcher.eventsSinks {
				if err := sink.close(); err != nil && err != errEventsSinkClosed {
					log.ErrorErr(err, "failed to close eventsSink %+v", sink)
				}
			}
			return
		}
	}
}
func (dispatcher *eventsSinkDispatcher) String() string {
	// avoid a race condition on sync.Once
	return fmt.Sprintf("%+v", map[string]interface{}{"eventsSinks": dispatcher.eventsSinks,
		"events":         dispatcher.events,
		"addRequests":    dispatcher.addRequests,
		"removeRequests": dispatcher.removeRequests,
		"shutdown":       dispatcher.shutdown,
		"closed":         dispatcher.closed})
}
