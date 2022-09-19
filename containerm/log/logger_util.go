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

package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	gwLogPrefixFormat           = "[container-management][%s:%d][pkg:%s][func:%s] "
	gwLogMessageFormat          = "%s %s" //%prefix %format origin
	gwLogMessageWithErrorFormat = gwLogMessageFormat + " %s "
)

var (
	fpcs             = make([]uintptr, 1)
	logLevelsMapping = map[string]logrus.Level{
		"ERROR": logrus.ErrorLevel,
		"WARN":  logrus.WarnLevel,
		"INFO":  logrus.InfoLevel,
		"DEBUG": logrus.DebugLevel,
		"TRACE": logrus.TraceLevel,
	}
)

func processFormat(formatOrigin string) string {
	return fmt.Sprintf(gwLogMessageFormat, preparePrefix(), formatOrigin)
}

func processFormatWithError(formatOrigin string, err error) string {
	return fmt.Sprintf(gwLogMessageWithErrorFormat, preparePrefix(), formatOrigin, prepareError(err))
}

func preparePrefix() string {
	var (
		fileName    = "n/a"
		funcPkgName = "n/a"
		funcName    = "n/a"
		fileLine    = -1
	)

	// Skip 4 levels to get the caller
	n := runtime.Callers(4, fpcs)
	if n != 0 {
		caller := runtime.FuncForPC(fpcs[0] - 1)
		if caller != nil {
			// Print the file name and line number
			fileName, fileLine = caller.FileLine(fpcs[0] - 1)
			fileName = filepath.Base(fileName)

			splitted := strings.Split(caller.Name(), ".")
			funcPkgName = splitted[0]
			funcName = splitted[1]
		}
	}
	return fmt.Sprintf(gwLogPrefixFormat, fileName, fileLine, funcPkgName, funcName)
}

func clear() {
	if logFileWriteCloser != nil {
		if err := logFileWriteCloser.Close(); err != nil {
			WarnErr(err, "failed to close log file output on logger exit")
		}
	}
}
