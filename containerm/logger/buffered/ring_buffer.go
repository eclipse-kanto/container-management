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

package buffered

import (
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/logger"
)

var ringBuffClosedError = log.NewErrorf("closed")

const (
	defaultMaxBytes = 1e6 //1MB
)

// ringBuffer implements a fixed-size buffer which will drop oldest data if full.
type ringBuffer struct {
	syncMux  sync.Mutex
	waitCond *sync.Cond

	isClosed  bool
	msgsQueue *messageQueue

	maxBytes     int64
	currentBytes int64
}

func newRingBuffer(maxBytes int64) *ringBuffer {
	if maxBytes < 0 {
		maxBytes = defaultMaxBytes
	}

	rb := &ringBuffer{
		isClosed:  false,
		msgsQueue: newMessageQueue(),
		maxBytes:  maxBytes,
	}
	rb.waitCond = sync.NewCond(&rb.syncMux)
	return rb
}

func (rb *ringBuffer) push(val *logger.LogMessage) error {
	rb.syncMux.Lock()
	defer rb.syncMux.Unlock()

	if rb.isClosed {
		return ringBuffClosedError
	}

	if val == nil {
		return nil
	}

	msgLength := int64(len(val.Line))
	if (rb.currentBytes + msgLength) > rb.maxBytes {
		rb.waitCond.Broadcast()
		return nil
	}

	rb.msgsQueue.enqueue(val)
	rb.waitCond.Broadcast()
	return nil
}

func (rb *ringBuffer) pop() (*logger.LogMessage, error) {
	rb.syncMux.Lock()
	for rb.msgsQueue.size() == 0 && !rb.isClosed {
		rb.waitCond.Wait()
	}

	if rb.isClosed {
		rb.syncMux.Unlock()
		return nil, ringBuffClosedError
	}

	val := rb.msgsQueue.dequeue()
	rb.currentBytes -= int64(len(val.Line))
	rb.syncMux.Unlock()
	return val, nil
}

func (rb *ringBuffer) drain() []*logger.LogMessage {
	rb.syncMux.Lock()
	defer rb.syncMux.Unlock()

	size := rb.msgsQueue.size()
	vals := make([]*logger.LogMessage, 0, size)

	for i := 0; i < size; i++ {
		vals = append(vals, rb.msgsQueue.dequeue())
	}
	rb.currentBytes = 0
	return vals
}

func (rb *ringBuffer) Close() error {
	rb.syncMux.Lock()
	if rb.isClosed {
		rb.syncMux.Unlock()
		return nil
	}

	rb.isClosed = true
	rb.waitCond.Broadcast()
	rb.syncMux.Unlock()
	return nil
}
