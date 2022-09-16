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

package jsonfile

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/logger"
)

// constants for json log file
const (
	jsonLogFileName              = "json.log"
	jsonLogFilePerms os.FileMode = 0644

	JSONFileLogDriverName logger.LogDriverType = "json-file"
)

//marshalHandler is the function of marshalLogMessageToJSONBytes the logMessage
type marshalHandler func(message *logger.LogMessage) ([]byte, error)

// jsonFileLogDriver is uses to log the container's stdout and stderr.
type jsonFileLogDriver struct {
	jsLogMux sync.Mutex

	jsLogFile          *os.File
	permissions        os.FileMode
	isClosed           bool
	marshalHandlerFunc marshalHandler
	maxSize            int64
	currentSize        int64
	maxFile            int
}

// NewJSONFileLog creates a new LogDriver instance that produces JSON-formatted logs.
func NewJSONFileLog(info logger.LogDriverInfo, configOpts ...logger.LogConfigOption) (logger.LogDriver, error) {
	logCfg := &jsonLogFileOpts{}
	if err := applyJSONLoggerOpts(logCfg, configOpts...); err != nil {
		log.ErrorErr(err, "invalid config provided for log driver %s", JSONFileLogDriverName)
		return nil, err
	}

	if _, err := os.Stat(info.ContainerRootDir); err != nil {
		return nil, err
	}
	logPath := filepath.Join(info.ContainerRootDir, jsonLogFileName)

	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, jsonLogFilePerms)
	if err != nil {
		return nil, err
	}

	var currentSize int64
	if len(configOpts) > 0 {
		size, err := f.Seek(0, io.SeekEnd)
		if err != nil {
			return nil, err
		}
		currentSize = size
	}
	return &jsonFileLogDriver{
		jsLogFile:   f,
		permissions: jsonLogFilePerms,
		isClosed:    false,
		marshalHandlerFunc: func(msg *logger.LogMessage) ([]byte, error) {
			return marshalLogMessageToJSONBytes(msg)
		},
		maxSize:     logCfg.maxSize,
		currentSize: currentSize,
		maxFile:     logCfg.maxFiles,
	}, nil
}

func (jsLogDriver *jsonFileLogDriver) Type() logger.LogDriverType {
	return JSONFileLogDriverName
}

func (jsLogDriver *jsonFileLogDriver) WriteLogMessage(msg *logger.LogMessage) error {
	b, err := jsLogDriver.marshalHandlerFunc(msg)
	if err != nil {
		return err
	}

	jsLogDriver.jsLogMux.Lock()
	defer jsLogDriver.jsLogMux.Unlock()
	if err = jsLogDriver.checkRotate(); err != nil {
		return err
	}

	n, err := jsLogDriver.jsLogFile.Write(b)
	if err == nil {
		jsLogDriver.currentSize += int64(n)
	}
	return err
}

func (jsLogDriver *jsonFileLogDriver) Close() error {
	jsLogDriver.jsLogMux.Lock()
	defer jsLogDriver.jsLogMux.Unlock()

	if jsLogDriver.isClosed {
		return nil
	}

	if err := jsLogDriver.jsLogFile.Close(); err != nil {
		return err
	}
	jsLogDriver.isClosed = true
	return nil
}
