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
	"bufio"
	"io"
	"sync"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

// LogHandler is responsible for handling the streams data to be logged.
type LogHandler interface {
	// StartCopyToLogDriver starts to copy the streams data that will be logged.
	StartCopyToLogDriver()
	// Wait blocks until goroutines have finished.
	Wait()
}

type containerLogHandler struct {
	sync.WaitGroup
	sources              map[string]io.Reader
	destinationLogDriver LogDriver
}

// NewLogHandler creates a new LogHandler
func NewLogHandler(destinationLogDriver LogDriver, sources map[string]io.Reader) LogHandler {
	return &containerLogHandler{
		sources:              sources,
		destinationLogDriver: destinationLogDriver,
	}
}

func (logHandler *containerLogHandler) StartCopyToLogDriver() {
	for source, r := range logHandler.sources {
		logHandler.Add(1)
		go logHandler.copyToLogDriver(source, r)
	}
}

func (logHandler *containerLogHandler) copyToLogDriver(source string, reader io.Reader) {
	defer log.Debug("finish %s stream type LogHandler for %s", source, logHandler.destinationLogDriver.Type())
	defer logHandler.Done()

	var (
		bytes []byte
		err   error

		isFirstSegment = true
		isPrefix       bool
		createdAt      time.Time

		defaultBufferSize = 16 * 1024
	)

	br := bufio.NewReaderSize(reader, defaultBufferSize)
	for {
		bytes, isPrefix, err = br.ReadLine()
		if err != nil {
			if err != io.EOF {
				log.ErrorErr(err, "failed to copy into %v-%v", logHandler.destinationLogDriver.Type(), source)
			}
			return
		}

		// The partial content will share the same timestamp.
		if isFirstSegment {
			createdAt = time.Now().UTC()
		}

		if isPrefix {
			isFirstSegment = false
		} else {
			isFirstSegment = true
			bytes = append(bytes, '\n')
		}

		if err = logHandler.destinationLogDriver.WriteLogMessage(&LogMessage{
			Source:    source,
			Line:      bytes,
			Timestamp: createdAt,
		}); err != nil {
			log.ErrorErr(err, "failed to copy into %v-%v", logHandler.destinationLogDriver.Type(), source)
		}
	}
}
