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
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/logger"
)

type jsonLogFileOpts struct {
	maxFiles int
	maxSize  int64
}

func applyJSONLoggerOpts(jsonLogOpts *jsonLogFileOpts, opts ...logger.LogConfigOption) error {
	for _, o := range opts {
		if err := o(jsonLogOpts); err != nil {
			return err
		}
	}
	return nil
}

// WithMaxFiles sets the maximum number of log files per container.
func WithMaxFiles(maxFiles int) logger.LogConfigOption {
	return func(specificConfigOpts interface{}) error {
		if maxFiles < 1 {
			return log.NewError("maxFiles logger config cannot be < 1")
		}
		specificConfigOpts.(*jsonLogFileOpts).maxFiles = maxFiles
		return nil
	}
}

// WithMaxSize sets the maximum size per log file.
func WithMaxSize(maxSize int64) logger.LogConfigOption {
	return func(specificConfigOpts interface{}) error {
		specificConfigOpts.(*jsonLogFileOpts).maxSize = maxSize
		return nil
	}
}
