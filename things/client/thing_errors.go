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

package client

import "fmt"

type dittoError string

const (
	messagesParameterInvalid dittoError = "messages:parameter.invalid"
	messagesSubjectNotFound  dittoError = "messages:subject.notfound"
	messagesExecutionFailed  dittoError = "messages:execution.failed"

	// error codes
	responseStatusBadRequest    = 400
	responseStatusNotFound      = 404
	responseStatusInternalError = 500

	thingErrorStringFormat = "[%d][%s] %s"
)

// ThingError represents the thing error
type ThingError struct {
	ErrorCode dittoError `json:"error"`
	Status    int        `json:"status"`
	Message   string     `json:"message"`
}

func (thErr *ThingError) Error() string {
	return fmt.Sprintf(thingErrorStringFormat, thErr.Status, thErr.ErrorCode, thErr.Message)
}

// NewMessagesParameterInvalidError creates a new thing error message for an invalid parameter
func NewMessagesParameterInvalidError(messageFormat string, args ...interface{}) *ThingError {
	return &ThingError{
		ErrorCode: messagesParameterInvalid,
		Status:    responseStatusBadRequest,
		Message:   fmt.Sprintf(messageFormat, args...),
	}
}

// NewMessagesSubjectNotFound creates a new thing error message for a subject not found
func NewMessagesSubjectNotFound(messageFormat string, args ...interface{}) *ThingError {
	return &ThingError{
		ErrorCode: messagesSubjectNotFound,
		Status:    responseStatusNotFound,
		Message:   fmt.Sprintf(messageFormat, args...),
	}
}

// NewMessagesInternalError creates a new thing error for an internal error
func NewMessagesInternalError(messageFormat string, args ...interface{}) *ThingError {
	return &ThingError{
		ErrorCode: messagesExecutionFailed,
		Status:    responseStatusInternalError,
		Message:   fmt.Sprintf(messageFormat, args...),
	}
}
