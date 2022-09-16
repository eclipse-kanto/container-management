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
	"os"
)

func (jsLogDriver *jsonFileLogDriver) checkRotate() error {
	if jsLogDriver.maxSize == 0 || jsLogDriver.currentSize < jsLogDriver.maxSize {
		// do not rotate
		return nil
	}

	logName := jsLogDriver.jsLogFile.Name()
	if err := jsLogDriver.jsLogFile.Close(); err != nil {
		return err
	}
	// TODO: after rotating logs, a notice should be made to the log reader
	if err := rotate(logName, jsLogDriver.maxFile); err != nil {
		return err
	}
	newFile, err := os.OpenFile(logName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	jsLogDriver.jsLogFile = newFile
	jsLogDriver.currentSize = 0

	return nil
}

// // readMessages reads the log messages and returns LogReaderForwarder
// func (jsLogDriver *jsonFileLogDriver) readMessages(cfg *logger.ReadConfig) *logger.LogReaderForwarder {
// 	watcher := logger.NewLogReaderForwarder()

// 	go func() {
// 		defer close(watcher.Messages)

// 		jsLogDriver.readLogs(watcher, cfg)
// 	}()
// 	return watcher
// }

// // readLogs reads the logs from the provided LogReaderForwarder and ReadConfig
// func (jsLogDriver *jsonFileLogDriver) readLogs(logReaderForwarder *logger.LogReaderForwarder, readConfig *logger.ReadConfig) {
// 	jsLogDriver.jsLogMux.Lock()
// 	logFile, err := os.Open(jsLogDriver.jsLogFile.Name())
// 	jsLogDriver.jsLogMux.Unlock()

// 	if err != nil {
// 		logReaderForwarder.ReaderErr <- err
// 		return
// 	}
// 	defer logFile.Close()

// 	if readConfig.Last > 0 {
// 		offset, err := getOffset(logFile, readConfig.Last)
// 		if err != nil {
// 			logReaderForwarder.ReaderErr <- err
// 			return
// 		}

// 		if _, err := logFile.Seek(offset, io.SeekStart); err != nil {
// 			logReaderForwarder.ReaderErr <- err
// 			return
// 		}
// 	}
// 	tail(logReaderForwarder, performUnmarshal, logFile, readConfig)

// 	if !readConfig.DoFollow {
// 		return
// 	}

// 	follow(logReaderForwarder, performUnmarshal, logFile, readConfig)
// }
