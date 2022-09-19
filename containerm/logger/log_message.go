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

package logger

import (
	"time"
)

// const defaultMessageBufferSize = 2048

// LogMessage represents the log message in the container json log.
type LogMessage struct {
	Source     string
	Line       []byte
	Timestamp  time.Time
	Attributes map[string]string
	Err        error
}

// // LogReaderForwarder is used to pass the log message to the reader.
// type LogReaderForwarder struct {
// 	Messages  chan *LogMessage
// 	ReaderErr chan error

// 	closeOnce            sync.Once
// 	closeNotifierChannel chan struct{}
// }

// // NewLogReaderForwarder creates a new LogReaderForwarder.
// func NewLogReaderForwarder() *LogReaderForwarder {
// 	return &LogReaderForwarder{
// 		Messages:             make(chan *LogMessage, defaultMessageBufferSize),
// 		ReaderErr:            make(chan error, 1),
// 		closeNotifierChannel: make(chan struct{}),
// 	}
// }

// // Close is used to stop the reader forwarding
// func (logRdrForwarder *LogReaderForwarder) Close() {
// 	logRdrForwarder.closeOnce.Do(func() {
// 		close(logRdrForwarder.closeNotifierChannel)
// 	})
// }

// // WatchClose waits for closing the notification channel.
// func (logRdrForwarder *LogReaderForwarder) WatchClose() <-chan struct{} {
// 	return logRdrForwarder.closeNotifierChannel
// }

// // ReadConfig holds the configuration of logs to be read.
// type ReadConfig struct {
// 	Since    time.Time
// 	Until    time.Time
// 	Last     int
// 	DoFollow bool
// }
