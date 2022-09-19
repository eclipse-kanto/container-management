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

	"github.com/pkg/errors"
)

const (
	errorFormat             = "\n\t Error: %v \n\t %+v"
	errorFormatNoStackTrace = "\n\t Error: %v"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// NewError returns an error message.
func NewError(message string) error {
	return errors.New(message)
}

// NewErrorf formats according to a format specifier and returns an error message.
func NewErrorf(format string, args ...interface{}) error {
	return errors.New(fmt.Sprintf(format, args...))
}

func prepareError(error error) string {
	if err, ok := error.(stackTracer); ok {
		return fmt.Sprintf(errorFormat, error, err.(stackTracer).StackTrace()[1:])
	}
	return fmt.Sprintf(errorFormatNoStackTrace, error)
}
