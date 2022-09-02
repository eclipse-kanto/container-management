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

package log

import (
	"github.com/sirupsen/logrus"
	"io"
)

var (
	logFileWriteCloser io.WriteCloser
)

// ArgsFunction can be used to skip expensive arguments generation when the level is not enabled
type ArgsFunction logrus.LogFunction

// Config represents the configuration options to be set for logging
type Config struct {
	LogFile       string `json:"log_file,omitempty"`
	LogLevel      string `json:"log_level,omitempty"`
	LogFileSize   int    `json:"log_file_size,omitempty"`
	LogFileCount  int    `json:"log_file_count,omitempty"`
	LogFileMaxAge int    `json:"log_file_max_age,omitempty"`
	Syslog        bool   `json:"syslog,omitempty"`
}

// Trace logs a message at level Trace on the standard logger.
func Trace(format string, args ...interface{}) {
	logrus.Tracef(processFormat(format), args...)
}

// TraceErr logs a message at level Trace on the standard logger.
func TraceErr(err error, format string, args ...interface{}) {
	logrus.Tracef(processFormatWithError(format, err), args...)
}

// TraceFn logs a message at level Trace on the standard logger.
func TraceFn(format string, fn ArgsFunction) {
	if logrus.IsLevelEnabled(logrus.TraceLevel) {
		logrus.Tracef(processFormat(format), fn()...)
	}
}

// Debug logs a message at level Debug on the standard logger.
func Debug(format string, args ...interface{}) {
	logrus.Debugf(processFormat(format), args...)
}

// DebugErr logs a message at level Debug on the standard logger.
func DebugErr(err error, format string, args ...interface{}) {
	logrus.Debugf(processFormatWithError(format, err), args...)
}

// DebugFn logs a message at level Debug on the standard logger.
func DebugFn(format string, fn ArgsFunction) {
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		logrus.Debugf(processFormat(format), fn()...)
	}
}

// Info logs a message at level Info on the standard logger.
func Info(format string, args ...interface{}) {
	logrus.Infof(processFormat(format), args...)
}

// InfoErr logs a message at level Info on the standard logger.
func InfoErr(err error, format string, args ...interface{}) {
	logrus.Infof(processFormatWithError(format, err), args...)
}

// InfoFn logs a message at level Info on the standard logger.
func InfoFn(format string, fn ArgsFunction) {
	if logrus.IsLevelEnabled(logrus.InfoLevel) {
		logrus.Infof(processFormat(format), fn()...)
	}
}

// Warn logs a message at level Warn on the standard logger.
func Warn(format string, args ...interface{}) {
	logrus.Warnf(processFormat(format), args...)
}

// WarnErr logs a message at level Warn on the standard logger.
func WarnErr(err error, format string, args ...interface{}) {
	logrus.Warnf(processFormatWithError(format, err), args...)
}

// WarnFn logs a message at level Warn on the standard logger.
func WarnFn(format string, fn ArgsFunction) {
	if logrus.IsLevelEnabled(logrus.WarnLevel) {
		logrus.Warnf(processFormat(format), fn()...)
	}
}

// Error logs a message at level Error on the standard logger.
func Error(format string, args ...interface{}) {
	logrus.Errorf(processFormat(format), args...)
}

// ErrorErr logs a message at level Error on the standard logger.
func ErrorErr(err error, format string, args ...interface{}) {
	logrus.Errorf(processFormatWithError(format, err), args...)
}

// ErrorFn logs a message at level Error on the standard logger.
func ErrorFn(format string, fn ArgsFunction) {
	if logrus.IsLevelEnabled(logrus.ErrorLevel) {
		logrus.Errorf(processFormat(format), fn()...)
	}
}

// Panic logs a message at level Panic on the standard logger.
func Panic(format string, args ...interface{}) {
	logrus.Panicf(processFormat(format), args...)
}

// PanicErr logs a message at level Panic on the standard logger.
func PanicErr(err error, format string, args ...interface{}) {
	logrus.Panicf(processFormatWithError(format, err), args...)
}

// PanicFn logs a message at level Panic on the standard logger.
func PanicFn(format string, fn ArgsFunction) {
	if logrus.IsLevelEnabled(logrus.PanicLevel) {
		logrus.Panicf(processFormat(format), fn()...)
	}
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatal(format string, args ...interface{}) {
	logrus.Fatalf(processFormat(format), args...)
}

// FatalErr logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func FatalErr(err error, format string, args ...interface{}) {
	logrus.Fatalf(processFormatWithError(format, err), args...)
}

// FatalFn logs a message at level Fatal on the standard logger.
func FatalFn(format string, fn ArgsFunction) {
	if logrus.IsLevelEnabled(logrus.FatalLevel) {
		logrus.Fatalf(processFormat(format), fn()...)
	}
}
