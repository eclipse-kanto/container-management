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

// LogDriverMode indicates available logging modes.
type LogDriverMode string

// const (
// 	// LogDriverModeBlocking not to use buffer to make logs blocking.
// 	LogDriverModeBlocking LogDriverMode = "blocking"
// 	// LogDriverModeNonBlocking means to use a buffer to make logs non blocking.
// 	LogDriverModeNonBlocking LogDriverMode = "non-blocking"
// )

// LogDriverType represents the different driver types - it must be used as identification instead of a plain string
type LogDriverType string

// LogDriver represents any kind of log drivers, such as jsonfile, syslog.
type LogDriver interface {
	Type() LogDriverType
	WriteLogMessage(msg *LogMessage) error
	Close() error
}

// LogConfigOption provides configuration options for the creation of new LogDriver instances.
type LogConfigOption func(specificConfigOpts interface{}) error

// LogDriverInfo provides container information for log driver.
type LogDriverInfo struct {
	ContainerID      string
	ContainerName    string
	ContainerImageID string
	ContainerRootDir string
	DaemonName       string
}
