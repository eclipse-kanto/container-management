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

package types

// LogDriver represents the different supported LogDriver types
type LogDriver string

const (
	// LogConfigDriverNone represents a special type to disable container logs handling
	LogConfigDriverNone LogDriver = "none"
	// LogConfigDriverJSONFile represents a LogDriver type that supports JSON-formatted logging
	LogConfigDriverJSONFile LogDriver = "json-file" // the default
)

// LogDriverConfiguration represents a log driver configuration
type LogDriverConfiguration struct {
	Type LogDriver `json:"type,omitempty"`
	// driver config - applicable for json-file only
	MaxFiles int    `json:"max_files,omitempty"`
	MaxSize  string `json:"max_size,omitempty"`
	RootDir  string `json:"root_dir,omitempty"`
}

// LogMode indicates available logging modes
type LogMode string

const (
	// LogModeBlocking specifies a blocking LogMode
	LogModeBlocking LogMode = "blocking" // the default
	// LogModeNonBlocking specifies a non-blocking LogMode
	LogModeNonBlocking LogMode = "non-blocking"
)

// LogModeConfiguration represents log mode configuration
type LogModeConfiguration struct {
	Mode LogMode `json:"mode,omitempty"`
	// applicable for non-blocking mode
	MaxBufferSize string `json:"max_buffer_size,omitempty"`
}

// LogConfiguration represents log configuration
type LogConfiguration struct {
	DriverConfig *LogDriverConfiguration `json:"driver_config,omitempty"`
	ModeConfig   *LogModeConfiguration   `json:"mode_config,omitempty"`
}
