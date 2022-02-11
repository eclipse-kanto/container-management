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

package buffered

import (
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/logger"
)

var elementsPool = &sync.Pool{New: func() interface{} { return new(messageQueueElement) }}

type messageQueueElement struct {
	next, previous *messageQueueElement
	message        *logger.LogMessage
}

func (e *messageQueueElement) reset() {
	e.next, e.previous = nil, nil
	e.message = nil
}

type messageQueue struct {
	root  messageQueueElement
	count int
}

func newMessageQueue() *messageQueue {
	messageQueue := new(messageQueue)

	messageQueue.root.next = &messageQueue.root
	messageQueue.root.previous = &messageQueue.root
	messageQueue.count = 0
	return messageQueue
}

func (msgQueue *messageQueue) size() int {
	return msgQueue.count
}

func (msgQueue *messageQueue) enqueue(val *logger.LogMessage) {
	element := elementsPool.Get().(*messageQueueElement)
	element.message = val

	current := msgQueue.root.previous

	current.next = element
	element.previous = current
	element.next = &msgQueue.root
	msgQueue.root.previous = element
	msgQueue.count++
}

func (msgQueue *messageQueue) dequeue() *logger.LogMessage {
	if msgQueue.size() == 0 {
		return nil
	}

	current := msgQueue.root.next
	current.previous.next = current.next
	current.next.previous = current.previous
	val := current.message

	current.reset()
	elementsPool.Put(current)
	msgQueue.count--
	return val
}
