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

package protocol

// Headers represents the headers in the protocol messages
type Headers interface {
	GetCorrelationID() string
	IsResponseRequired() bool
	GetChannel() string
	IsDryRun() bool
	GetOrigin() string
	GetOriginator() string
	GetETag() string
	GetIfMatch() string
	GetIfNoneMatch() string
	GetReplyTarget() int64
	GetReplyTo() string
	GetVersion() int64
	GetContentType() string
	GetGeneric(id string) interface{}
	ToMap() map[string]interface{}
}
