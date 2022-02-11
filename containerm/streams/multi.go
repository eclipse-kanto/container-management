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

// Package name changed also removed not needed logic and added custom code to handle the specific use case, Bosch.IO GmbH, 2020

package streams

import (
	"io"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

// multiWriter allows caller to broadcast data to several writers.
type multiWriter struct {
	sync.Mutex
	writers []io.WriteCloser
}

// Add registers one writer into MultiWriter.
func (mw *multiWriter) Add(writer io.WriteCloser) {
	mw.Lock()
	mw.writers = append(mw.writers, writer)
	mw.Unlock()
}

// Write writes data into several writers and never returns error.
func (mw *multiWriter) Write(p []byte) (int, error) {
	mw.Lock()
	var evictIdx []int
	for n, w := range mw.writers {
		if _, err := w.Write(p); err != nil {
			log.DebugErr(err, "failed to write data")

			w.Close()
			evictIdx = append(evictIdx, n)
		}
	}

	for n, i := range evictIdx {
		mw.writers = append(mw.writers[:i-n], mw.writers[i-n+1:]...)
	}
	mw.Unlock()
	return len(p), nil
}

// Close closes all the writers and never returns error.
func (mw *multiWriter) Close() error {
	mw.Lock()
	for _, w := range mw.writers {
		w.Close()
	}
	mw.writers = nil
	mw.Unlock()
	return nil
}
