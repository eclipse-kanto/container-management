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
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/logger"
)

// bufferedLog is uses to cache the container's logs with ringBuffer.
type bufferedLog struct {
	ringBuffer *ringBuffer
	logDriver  logger.LogDriver
}

// NewBufferedLog creates a new LogDriver instance that enables buffered containers logs handling.
func NewBufferedLog(logDriver logger.LogDriver, maxBytes int64) (logger.LogDriver, error) {
	bl := &bufferedLog{
		logDriver:  logDriver,
		ringBuffer: newRingBuffer(maxBytes),
	}
	go bl.run()
	return bl, nil
}

func (bl *bufferedLog) Type() logger.LogDriverType {
	return bl.logDriver.Type()
}

func (bl *bufferedLog) WriteLogMessage(msg *logger.LogMessage) error {
	return bl.ringBuffer.push(msg)
}

func (bl *bufferedLog) Close() error {
	bl.ringBuffer.Close()
	for _, msg := range bl.ringBuffer.drain() {
		if err := bl.logDriver.WriteLogMessage(msg); err != nil {
			log.Debug("failed to write log %v when closing with log driver %s", msg, bl.logDriver.Type())
		}
	}

	return bl.logDriver.Close()
}

// write logs continuously with specified log driver from ringBuffer.
func (bl *bufferedLog) run() {
	for {
		msg, err := bl.ringBuffer.pop()
		if err != nil {
			return
		}

		if err := bl.logDriver.WriteLogMessage(msg); err != nil {
			log.Debug("failed to write log %v with log driver %s", msg, bl.logDriver.Type())
		}
	}
}
