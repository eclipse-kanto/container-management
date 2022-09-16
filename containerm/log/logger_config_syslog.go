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

//go:build !windows && !nacl && !plan9
// +build !windows,!nacl,!plan9

package log

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"log/syslog"
	"os"
	"time"

	"github.com/sirupsen/logrus"
	logrussyslog "github.com/sirupsen/logrus/hooks/syslog"
)

var syslogLevelsMapping = map[string]syslog.Priority{
	"ERROR": syslog.LOG_ERR,
	"WARN":  syslog.LOG_WARNING,
	"INFO":  syslog.LOG_INFO,
	"DEBUG": syslog.LOG_DEBUG,
	"TRACE": syslog.LOG_DEBUG,
}

// Configure applies the full configuration of the logger instance based on the provided configuration
func Configure(cfg *Config) {
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}
	logrus.SetFormatter(formatter)
	logrus.SetLevel(logLevelsMapping[cfg.LogLevel])

	var (
		sysLogErr  error
		sysLogHook *logrussyslog.SyslogHook
	)
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
		logrus.RegisterExitHandler(clear)
	} else if cfg.Syslog {
		sysLogHook, sysLogErr = logrussyslog.NewSyslogHook("", "", syslogLevelsMapping[cfg.LogLevel], "container-management")
		if sysLogErr == nil {
			//TODO add hook for windows events log
			logrus.AddHook(sysLogHook)
		} else {
			logrus.SetOutput(os.Stdout)
		}
	} else {
		logrus.SetOutput(os.Stdout)
	}

	if cfg.Syslog && cfg.LogFile != "" {
		Warn("both sys log and log file are configured for daemon debug - log file will be used only")
	}
	if sysLogErr != nil {
		Warn("could not enable sys log due to error %v", sysLogErr)
	}
}
