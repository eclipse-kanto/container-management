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

package things

import "github.com/eclipse-kanto/container-management/containerm/containers/types"

type logDriver string

const (
	jsonFile logDriver = "JSON_FILE"
	none     logDriver = "NONE"
)

type logMode string

const (
	blocking    logMode = "BLOCKING"
	nonBlocking logMode = "NON_BLOCKING"
)

type logConfiguration struct {
	Type          logDriver `json:"type,omitempty"`
	MaxFiles      int       `json:"maxFiles,omitempty"`
	MaxSize       string    `json:"maxSize,omitempty"`
	RootDir       string    `json:"rootDir,omitempty"`
	Mode          logMode   `json:"mode,omitempty"`
	MaxBufferSize string    `json:"maxBufferSize,omitempty"`
}

func toAPILogConfiguration(logConfig *logConfiguration) *types.LogConfiguration {
	return &types.LogConfiguration{
		DriverConfig: &types.LogDriverConfiguration{
			Type:     toAPILogDriver(logConfig.Type),
			MaxFiles: logConfig.MaxFiles,
			MaxSize:  logConfig.MaxSize,
			RootDir:  logConfig.RootDir,
		},
		ModeConfig: &types.LogModeConfiguration{
			Mode:          toAPILogMode(logConfig.Mode),
			MaxBufferSize: logConfig.MaxBufferSize,
		},
	}
}

func fromAPILogConfiguration(logConfig *types.LogConfiguration) *logConfiguration {
	cfg := &logConfiguration{}
	if logConfig.DriverConfig != nil {
		cfg.Type = fromAPILogDriver(logConfig.DriverConfig.Type)
		cfg.MaxFiles = logConfig.DriverConfig.MaxFiles
		cfg.MaxSize = logConfig.DriverConfig.MaxSize
		cfg.RootDir = logConfig.DriverConfig.RootDir
	}
	if logConfig.ModeConfig != nil {
		cfg.Mode = fromAPILogMode(logConfig.ModeConfig.Mode)
		cfg.MaxBufferSize = logConfig.ModeConfig.MaxBufferSize
	}
	return cfg
}

func toAPILogDriver(logType logDriver) types.LogDriver {
	switch logType {
	case jsonFile:
		return types.LogConfigDriverJSONFile
	case none:
		return types.LogConfigDriverNone
	default:
		return types.LogDriver(logType)
	}
}

func fromAPILogDriver(apiLogDriver types.LogDriver) logDriver {
	switch apiLogDriver {
	case types.LogConfigDriverJSONFile:
		return jsonFile
	case types.LogConfigDriverNone:
		return none
	default:
		return logDriver(apiLogDriver)
	}
}

func toAPILogMode(logMode logMode) types.LogMode {
	switch logMode {
	case blocking:
		return types.LogModeBlocking
	case nonBlocking:
		return types.LogModeNonBlocking
	default:
		return types.LogMode(logMode)
	}
}

func fromAPILogMode(apiLogMode types.LogMode) logMode {
	switch apiLogMode {
	case types.LogModeBlocking:
		return blocking
	case types.LogModeNonBlocking:
		return nonBlocking
	default:
		return logMode(apiLogMode)
	}
}
