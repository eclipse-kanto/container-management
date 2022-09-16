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

// // This utility file contains the logic for reading the log messages in real time.

// import (
// 	"bytes"
// 	"context"
// 	"io"
// 	"os"
// 	"time"

// 	"github.com/eclipse-kanto/container-management/containerm/log"
// 	"github.com/eclipse-kanto/container-management/containerm/logger"

// 	"github.com/fsnotify/fsnotify"
// )

// const (
// 	blockSize = 1024
// 	endOfLine = '\n'
// )

// type unmarshaler func(r io.Reader) func() (*logger.LogMessage, error)

// var watchFileTimeout = 200 * time.Millisecond

// // acts like `tail -f`.
// func follow(forwarder *logger.LogReaderForwarder, unmarshaler unmarshaler, file *os.File, readConfig *logger.ReadConfig) {
// 	fileChangeWatcher, err := startFileChangeWatching(file.Name())
// 	if err != nil {
// 		forwarder.ReaderErr <- err
// 		return
// 	}

// 	defer func() {
// 		fileChangeWatcher.Remove(file.Name())
// 		fileChangeWatcher.Close()
// 	}()

// 	ctx, cancel := context.WithCancel(context.TODO())
// 	defer cancel()

// 	go func() {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case <-forwarder.WatchClose():
// 			cancel()
// 		}
// 	}()

// 	decodeOneLine := unmarshaler(file)

// 	doneError := log.NewError("done")

// 	// NOTE: avoid to use time.After in select. We need local-global timeout
// 	watchTimeout := time.NewTimer(time.Second)
// 	defer watchTimeout.Stop()

// 	// errorHandler will watch the file if the err is io.EOF so that
// 	// the loop can continue to readLogs the file. Or just return the error.
// 	errorHandler := func(err error) error {
// 		if err != io.EOF {
// 			return err
// 		}

// 		for {
// 			watchTimeout.Reset(watchFileTimeout)

// 			select {
// 			case <-ctx.Done():
// 				return doneError
// 			case fileChangedEvent := <-fileChangeWatcher.Events:
// 				switch fileChangedEvent.Op {
// 				case fsnotify.Write:
// 					decodeOneLine = unmarshaler(file)
// 					return nil
// 				case fsnotify.Remove:
// 					return doneError
// 				default:
// 					log.Debug("unexpected file change during watching file %s: %v", file.Name(), fileChangedEvent.Op)
// 					return doneError
// 				}
// 			case fileChangedError := <-fileChangeWatcher.Errors:
// 				log.DebugErr(fileChangedError, "unexpected error during watching file %v", file.Name())
// 				return err
// 			case <-watchTimeout.C:
// 				_, statError := os.Stat(file.Name())
// 				if statError != nil {
// 					if os.IsNotExist(statError) {
// 						return doneError
// 					}
// 					log.DebugErr(statError, "unexpected error during watching file %s", file.Name())
// 					return doneError
// 				}
// 			}
// 		}
// 	}

// 	// continue to readLogs log
// 	for {
// 		msg, err := decodeOneLine()
// 		if err != nil {
// 			if err = errorHandler(err); err != nil {
// 				if err == doneError {
// 					return
// 				}

// 				forwarder.ReaderErr <- err
// 				return
// 			}
// 			continue
// 		}

// 		if !readConfig.Since.IsZero() && msg.Timestamp.Before(readConfig.Since) {
// 			continue
// 		}

// 		if !readConfig.Until.IsZero() && msg.Timestamp.After(readConfig.Until) {
// 			return
// 		}

// 		select {
// 		case <-ctx.Done():
// 			return
// 		case forwarder.Messages <- msg:
// 		}
// 	}
// }

// //  watch the change of a file
// func startFileChangeWatching(filePath string) (*fsnotify.Watcher, error) {
// 	watcher, err := fsnotify.NewWatcher()
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err := watcher.Add(filePath); err != nil {
// 		return nil, err
// 	}
// 	return watcher, nil
// }

// //  read the log message until the io.EOF or limited by config
// func tail(watcher *logger.LogReaderForwarder, unmarshaler unmarshaler, r io.Reader, cfg *logger.ReadConfig) {
// 	decodeOneLine := unmarshaler(r)

// 	for {
// 		message, err := decodeOneLine()
// 		if err != nil {
// 			if err != io.EOF {
// 				watcher.ReaderErr <- err
// 			}
// 			return
// 		}

// 		if !cfg.Since.IsZero() && message.Timestamp.Before(cfg.Since) {
// 			continue
// 		}

// 		if !cfg.Until.IsZero() && message.Timestamp.After(cfg.Until) {
// 			return
// 		}

// 		select {
// 		case <-watcher.WatchClose():
// 			return
// 		case watcher.Messages <- message:
// 		}
// 	}
// }

// // seek the offset in file by the number lines
// func getOffset(readSeeker io.ReadSeeker, n int) (int64, error) {
// 	if n <= 0 {
// 		return 0, nil
// 	}

// 	size, err := readSeeker.Seek(0, io.SeekEnd)
// 	if err != nil {
// 		return 0, err
// 	}

// 	var (
// 		block    = -1
// 		count    = 0
// 		left     = int64(0)
// 		readBuff []byte

// 		readN int64
// 	)

// 	for {
// 		readN = int64(blockSize)
// 		left = size + int64(block*blockSize)
// 		if left < 0 {
// 			readN = int64(blockSize) + left
// 			left = 0
// 		}

// 		readBuff = make([]byte, readN)
// 		if _, err := readSeeker.Seek(left, io.SeekStart); err != nil {
// 			return 0, err
// 		}

// 		if _, err := readSeeker.Read(readBuff); err != nil {
// 			return 0, err
// 		}

// 		// if the line is enough or the file doesn't contain such lines
// 		count += bytes.Count(readBuff, []byte{endOfLine})
// 		if count > n || left == 0 {
// 			break
// 		}
// 		block--
// 	}

// 	for count > n {
// 		if idx := bytes.IndexByte(readBuff, endOfLine); idx >= 0 {
// 			left += int64(idx) + 1
// 			readBuff = readBuff[idx+1:]
// 		}
// 		count--
// 	}
// 	return left, nil
// }
