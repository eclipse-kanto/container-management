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
	"container/list"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

// queueEventsSync accepts all messages into a queueEventsSync for asynchronous consumption by an eventsSink
type queueEventsSync struct {
	wrappedEventsSink eventsSink
	events            *list.List
	syncCondition     *sync.Cond
	qEvSinkMutex      sync.Mutex
	closed            bool
}

// newQueueEventsSink returns a queueEventsSync to the provided eventsSink
func newQueueEventsSink(dst eventsSink) *queueEventsSync {
	eq := queueEventsSync{
		wrappedEventsSink: dst,
		events:            list.New(),
	}

	eq.syncCondition = sync.NewCond(&eq.qEvSinkMutex)
	go eq.run()
	return &eq
}
func (qEvSink *queueEventsSync) write(event event) error {
	qEvSink.qEvSinkMutex.Lock()
	defer qEvSink.qEvSinkMutex.Unlock()

	if qEvSink.closed {
		return errEventsSinkClosed
	}

	qEvSink.events.PushBack(event)
	qEvSink.syncCondition.Signal() // signal waiters

	return nil
}
func (qEvSink *queueEventsSync) close() error {
	qEvSink.qEvSinkMutex.Lock()
	defer qEvSink.qEvSinkMutex.Unlock()

	if qEvSink.closed {
		return nil
	}

	qEvSink.closed = true
	qEvSink.syncCondition.Signal() // signal flushes the queue
	qEvSink.syncCondition.Wait()   // wait for signal from last flush
	return qEvSink.wrappedEventsSink.close()
}
func (qEvSink *queueEventsSync) run() {
	for {
		event := qEvSink.next()

		if event == nil {
			return // nil block means queueEventsSync is closed
		}
		if err := qEvSink.wrappedEventsSink.write(event); err != nil {
			log.ErrorErr(err, "dropping event from event queueEventsSync for eventsSink %+v", qEvSink.wrappedEventsSink)
		}
	}
}
func (qEvSink *queueEventsSync) next() event {
	qEvSink.qEvSinkMutex.Lock()
	defer qEvSink.qEvSinkMutex.Unlock()

	for qEvSink.events.Len() < 1 {
		if qEvSink.closed {
			qEvSink.syncCondition.Broadcast()
			return nil
		}

		qEvSink.syncCondition.Wait()
	}

	front := qEvSink.events.Front()
	block := front.Value.(event)
	qEvSink.events.Remove(front)

	return block
}
