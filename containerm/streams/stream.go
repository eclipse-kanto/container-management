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
	"context"
	"io"
	"sync"

	errutil "github.com/eclipse-kanto/container-management/containerm/util/error"
	ioutils "github.com/eclipse-kanto/container-management/containerm/util/io"
)

// Stream for containers IOs handling
type Stream interface {
	Stdin() io.ReadCloser
	StdinPipe() io.WriteCloser
	NewStdinInput()
	NewDiscardStdinInput()
	Stdout() io.WriteCloser
	AddStdoutWriter(w io.WriteCloser)
	AddStderrWriter(w io.WriteCloser)
	Stderr() io.WriteCloser
	NewStdoutPipe() io.ReadCloser
	NewStderrPipe() io.ReadCloser
	Attach(ctx context.Context, cfg *AttachConfig) <-chan error
	CopyPipes(p Pipes)
	Close() error
	// sync.WaitGroup interface
	Wait()
}

// NewStream returns new streams.
func NewStream() Stream {
	return &cIOStream{
		stdout: &multiWriter{},
		stderr: &multiWriter{},
	}
}

// cIOStream is used to handle container IO.
type cIOStream struct {
	sync.WaitGroup
	stdin          io.ReadCloser
	stdinPipe      io.WriteCloser
	stdout, stderr *multiWriter
}

// Stdin returns the Stdin for reader.
func (s *cIOStream) Stdin() io.ReadCloser {
	return s.stdin
}

// StdinPipe returns the Stdin for writer.
func (s *cIOStream) StdinPipe() io.WriteCloser {
	return s.stdinPipe
}

// NewStdinInput creates pipe for Stdin() and StdinPipe().
func (s *cIOStream) NewStdinInput() {
	s.stdin, s.stdinPipe = io.Pipe()
}

// NewDiscardStdinInput creates a no-op WriteCloser for StdinPipe().
func (s *cIOStream) NewDiscardStdinInput() {
	s.stdin, s.stdinPipe = nil, ioutils.NewNoopWriteCloser()
}

// Stdout returns the Stdout for writer.
func (s *cIOStream) Stdout() io.WriteCloser {
	return s.stdout
}

// Stderr returns the Stderr for writer.
func (s *cIOStream) Stderr() io.WriteCloser {
	return s.stderr
}

// AddStdoutWriter adds the stdout writer.
func (s *cIOStream) AddStdoutWriter(w io.WriteCloser) {
	s.stdout.Add(w)
}

// AddStderrWriter adds the stderr writer.
func (s *cIOStream) AddStderrWriter(w io.WriteCloser) {
	s.stderr.Add(w)
}

// NewStdoutPipe creates pipe and register it into Stdout.
func (s *cIOStream) NewStdoutPipe() io.ReadCloser {
	r, w := io.Pipe()
	s.stdout.Add(w)
	return r
}

// NewStderrPipe creates pipe and register it into Stderr.
func (s *cIOStream) NewStderrPipe() io.ReadCloser {
	r, w := io.Pipe()
	s.stderr.Add(w)
	return r
}

// Close closes streams.
func (s *cIOStream) Close() error {
	cErr := new(errutil.CompoundError)

	if s.stdin != nil {
		if err := s.stdin.Close(); err != nil {
			cErr.Append(err)
		}
	}

	if err := s.stdout.Close(); err != nil {
		cErr.Append(err)
	}

	if err := s.stderr.Close(); err != nil {
		cErr.Append(err)
	}

	if cErr.Size() > 0 {
		return cErr
	}
	return nil
}
