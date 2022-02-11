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

package io

import "io"

type readCloserWrapper struct {
	io.Reader
	closeFunc func() error
}

func (r *readCloserWrapper) Close() error {
	return r.closeFunc()
}

// NewReadCloserWrapper provides the ability to handle the cleanup during closer.
func NewReadCloserWrapper(r io.Reader, closeFunc func() error) io.ReadCloser {
	return &readCloserWrapper{
		Reader:    r,
		closeFunc: closeFunc,
	}
}
