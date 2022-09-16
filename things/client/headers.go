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

import (
	"encoding/json"
)

const (
	headerCorrelationID    = "correlation-id"
	headerResponseRequired = "response-required"
	headerChannel          = "ditto-channel"
	headerDryRun           = "ditto-dry-run"
	headerOrigin           = "origin"
	headerOriginator       = "ditto-originator"
	headerETag             = "ETag"
	headerIfMatch          = "If-Match"
	headerIfNoneMatch      = "If-None-Match"
	headerReplyTarget      = "ditto-reply-target"
	headerReplyTo          = "reply-to"
	headerSchemaVersion    = "version"
	headerContentType      = "content-type"
)

type headers struct {
	values map[string]interface{}
}

func (h *headers) GetCorrelationID() string {
	if h.values[headerCorrelationID] == nil {
		return ""
	}
	return h.values[headerCorrelationID].(string)
}
func (h *headers) IsResponseRequired() bool {
	if h.values[headerResponseRequired] == nil {
		return false
	}
	return h.values[headerResponseRequired].(bool)
}
func (h *headers) GetChannel() string {
	if h.values[headerChannel] == nil {
		return ""
	}
	return h.values[headerChannel].(string)
}
func (h *headers) IsDryRun() bool {
	if h.values[headerDryRun] == nil {
		return false
	}
	return h.values[headerDryRun].(bool)
}
func (h *headers) GetOrigin() string {
	if h.values[headerOrigin] == nil {
		return ""
	}
	return h.values[headerOrigin].(string)
}
func (h *headers) GetOriginator() string {
	if h.values[headerOriginator] == nil {
		return ""
	}
	return h.values[headerOriginator].(string)
}
func (h *headers) GetETag() string {
	if h.values[headerETag] == nil {
		return ""
	}
	return h.values[headerETag].(string)
}
func (h *headers) GetIfMatch() string {
	if h.values[headerIfMatch] == nil {
		return ""
	}
	return h.values[headerIfMatch].(string)
}
func (h *headers) GetIfNoneMatch() string {
	if h.values[headerIfNoneMatch] == nil {
		return ""
	}
	return h.values[headerIfNoneMatch].(string)
}
func (h *headers) GetReplyTarget() int64 {
	if h.values[headerReplyTarget] == nil {
		return -1
	}
	return h.values[headerReplyTarget].(int64)
}
func (h *headers) GetReplyTo() string {
	if h.values[headerReplyTo] == nil {
		return ""
	}
	return h.values[headerReplyTo].(string)
}
func (h *headers) GetVersion() int64 {
	if h.values[headerSchemaVersion] == nil {
		return -1
	}
	return h.values[headerSchemaVersion].(int64)
}
func (h *headers) GetContentType() string {
	if h.values[headerContentType] == nil {
		return ""
	}
	return h.values[headerContentType].(string)
}
func (h *headers) GetGeneric(id string) interface{} {
	return h.values[id]
}
func (h *headers) ToMap() map[string]interface{} {
	return h.values
}

func (h *headers) MarshalJSON() ([]byte, error) {
	return json.Marshal(h.values)
}

func (h *headers) UnmarshalJSON(data []byte) error {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	h.values = v
	return nil
}
