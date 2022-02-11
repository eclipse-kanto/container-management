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
)

// in the future we will have a more meaningful interface here
type event interface {
}

// eventsSink accepts and sends events
type eventsSink interface {
	write(event event) error
	close() error
}

// channelledEventsSink provides an eventsSink that can be listened on.
//The writer and channel listener must operate in separate goroutines.
type channelledEventsSink struct {
	eventsChannel chan event
	closed        chan struct{}
	once          sync.Once
}

func newChannelledEventsSink(buffer int) *channelledEventsSink {
	return &channelledEventsSink{
		eventsChannel: make(chan event, buffer),
		closed:        make(chan struct{}),
	}
}

func (chanEvSink *channelledEventsSink) done() chan struct{} {
	return chanEvSink.closed
}

func (chanEvSink *channelledEventsSink) write(event event) error {
	select {
	case chanEvSink.eventsChannel <- event:
		return nil
	case <-chanEvSink.closed:
		return errEventsSinkClosed
	}
}

func (chanEvSink *channelledEventsSink) close() error {
	chanEvSink.once.Do(func() {
		close(chanEvSink.closed)
	})

	return nil
}
func (chanEvSink *channelledEventsSink) String() string {
	// avoid a race condition on sync.Once
	return fmt.Sprintf("%+v", map[string]interface{}{
		"eventsChannel": chanEvSink.eventsChannel,
		"closed":        chanEvSink.closed,
	})
}
