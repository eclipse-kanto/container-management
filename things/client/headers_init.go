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

package client

import "github.com/eclipse-kanto/container-management/things/client/protocol"

// HeaderOpt represents header options
type HeaderOpt func(hdrs *headers) error

func applyOptsHeader(headers *headers, opts ...HeaderOpt) error {
	for _, o := range opts {
		if err := o(headers); err != nil {
			return err
		}
	}
	return nil
}

// NewHeaders creates a new protocol header
func NewHeaders(opts ...HeaderOpt) protocol.Headers {
	res := &headers{}
	res.values = make(map[string]interface{})
	if err := applyOptsHeader(res, opts...); err != nil {
		return nil
	}
	return res
}

// WithCorrelationID sets a header value for correlation id
func WithCorrelationID(correlationID string) HeaderOpt {
	return func(hdrs *headers) error {
		if correlationID != "" {
			hdrs.values[headerCorrelationID] = correlationID
		}
		return nil
	}
}

// WithReplyTo sets a header value for reply to
func WithReplyTo(replyTo string) HeaderOpt {
	return func(hdrs *headers) error {
		if replyTo != "" {
			hdrs.values[headerReplyTo] = replyTo
		}
		return nil
	}
}

// WithResponseRequired sets a header value for response required
func WithResponseRequired(isResponseRequired bool) HeaderOpt {
	return func(hdrs *headers) error {
		hdrs.values[headerResponseRequired] = isResponseRequired
		return nil
	}
}

// WithContentType sets a header value for content type
func WithContentType(contentType string) HeaderOpt {
	return func(hdrs *headers) error {
		if contentType != "" {
			hdrs.values[headerContentType] = contentType
		}
		return nil
	}
}
