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

//go:build windows && nacl && plan9
// +build windows,nacl,plan9

package log

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Configure applies the full configuration of the logger instance based on the provided configuration
func Configure(cfg *Config) {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}
	logrus.SetFormatter(formatter)
	logrus.SetLevel(logLevelsMapping[cfg.LogLevel])

	if cfg.LogFile != "" {
		logFileWriteCloser = &lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    cfg.LogFileSize,
			MaxBackups: cfg.LogFileCount,
			MaxAge:     cfg.LogFileMaxAge,
			LocalTime:  true,
			Compress:   true,
		}
		logrus.SetOutput(logFileWriteCloser)
	} else {
		logrus.SetOutput(os.Stdout)
	}

	if cfg.Syslog && cfg.LogFile != "" {
		Warn("both sys log and log file are configured for daemon debug - log file will be used only")
	}

	if mkdirErr != nil && cfg.LogFile != "" {
		Warn("could not load log file %s due to error %v", cfg.LogFile, mkdirErr)
	}

	if cfg.Syslog {
		Warn("syslog is not supported for the target platform - cannot enable it")
	}
}
