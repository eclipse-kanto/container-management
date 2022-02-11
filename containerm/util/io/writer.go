// Copyright The PouchContainer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package name changed, Bosch.IO GmbH, 2020

package io

import "io"

// noopWriter is an io.Writer on which all Write calls succeed without
// doing anything.
type noopWriter struct{}

func (nw *noopWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (nw *noopWriter) Close() error {
	return nil
}

// NewNoopWriteCloser returns the no-op WriteCloser.
func NewNoopWriteCloser() io.WriteCloser {
	return &noopWriter{}
}

type writeCloserWrapper struct {
	io.Writer
	closeFunc func() error
}

func (w *writeCloserWrapper) Close() error {
	return w.closeFunc()
}

// NewWriteCloserWrapper provides the ability to handle the cleanup during closer.
func NewWriteCloserWrapper(w io.Writer, closeFunc func() error) io.WriteCloser {
	return &writeCloserWrapper{
		Writer:    w,
		closeFunc: closeFunc,
	}
}

// CloseWriter is an interface which represents the implementation closes the
// writing side of writer.
type CloseWriter interface {
	CloseWrite() error
}
