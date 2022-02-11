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

package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	pbcontainers "github.com/eclipse-kanto/container-management/containerm/api/services/containers"
)

const (
	// MaxBufSize is the maximum buffer size (in bytes) received in a read chunk or sent in a write chunk.
	MaxBufSize  = 2 * 1024 * 1024
	backoffBase = 10 * time.Millisecond
	backoffMax  = 1 * time.Second
	maxTries    = 5
)

// Reader reads from a byte stream.
type Reader struct {
	ctx         context.Context
	readServer  pbcontainers.Containers_AttachServer
	containerID string
	stdIn       bool
	err         error
	buf         []byte
}

// ContainerID gets the container id of the IO this Reader is reading.
func (r *Reader) ContainerID() string {
	return r.containerID
}

// StdIn gets whether the IO should be Interactive.
func (r *Reader) StdIn() bool {
	return r.stdIn
}

// Read implements io.Reader.
// Read buffers received bytes that do not fit in p.
func (r *Reader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	var backoffDelay time.Duration
	for tries := 0; len(r.buf) == 0 && tries < maxTries; tries++ {
		// No data in buffer.
		req, err := r.readServer.Recv()
		if err != nil {
			r.err = err
			return 0, err
		}
		r.buf = req.DataToWrite
		if len(r.buf) != 0 {
			break
		}

		// back off
		if backoffDelay < backoffBase {
			backoffDelay = backoffBase
		} else {
			backoffDelay = time.Duration(float64(backoffDelay) * 1.3 * (1 - 0.4*rand.Float64()))
		}
		if backoffDelay > backoffMax {
			backoffDelay = backoffMax
		}
		select {
		case <-time.After(backoffDelay):
		case <-r.ctx.Done():
			if err := r.ctx.Err(); err != nil {
				r.err = err
			}
			return 0, r.err
		}
	}

	// Copy from buffer.
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

// Close implements io.Closer.
func (r *Reader) Close() error {
	if r.readServer == nil {
		return nil
	}
	r.readServer = nil
	return nil
}

// NewReader creates a new Reader to read a resource.
func NewReader(ctx context.Context, containerID string, stdIn bool, readServer pbcontainers.Containers_AttachServer) (*Reader, error) {
	return NewReaderAt(ctx, containerID, stdIn, readServer, 0)
}

// NewReaderAt creates a new Reader to read a resource from the given offset.
func NewReaderAt(ctx context.Context, containerID string, stdIn bool, readServer pbcontainers.Containers_AttachServer, offset int64) (*Reader, error) {
	// readClient is set up for Read(). ReadAt() will copy needed fields into its reentrantReader.
	return &Reader{
		ctx:         ctx,
		containerID: containerID,
		stdIn:       stdIn,
		readServer:  readServer,
	}, nil
}

// Writer writes to a byte stream.
type Writer struct {
	ctx         context.Context
	writeServer pbcontainers.Containers_AttachServer
	containerID string
	stdIn       bool
	offset      int64
	err         error
}

// ContainerID gets the container ID of the IO this Writer is writing.
func (w *Writer) ContainerID() string {
	return w.containerID
}

// StdIn checks whether the IO should be attached to STDIN also - i.e. interaction is enabled.
func (w *Writer) StdIn() bool {
	return w.stdIn
}

// Write implements io.Writer.
func (w *Writer) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}

	n := 0
	for n < len(p) {
		bufSize := len(p) - n
		if bufSize > MaxBufSize {
			bufSize = MaxBufSize
		}
		r := pbcontainers.AttachContainerResponse{
			ReadData: p,
		}
		// Bytestream only requires the resourceName to be sent in the first WriteRequest.
		if w.offset == 0 {
			r.Id = w.containerID
			r.StdIn = w.stdIn
		}
		err := w.writeServer.Send(&r)
		if err != nil {
			w.err = err
			return n, err
		}
		w.offset += int64(bufSize)
		n += bufSize
	}
	return n, nil
}

// Close implements io.Closer. It is the caller's responsibility to call Close() when writing is done.
func (w *Writer) Close() error {
	err := w.writeServer.Send(&pbcontainers.AttachContainerResponse{
		ReadData: nil,
	})
	if err != nil {
		w.err = err
		return fmt.Errorf("Send(WriteRequest< FinishWrite >) failed: %v", err)
	}
	w.err = err
	return err
}

// NewWriter creates a new Writer to write a resource.
//
// resourceName specifies the name of the resource.
// The resource will be available after Close has been called.
//
// It is the caller's responsibility to call Close when writing is done.
//
// TODO: There is currently no way to resume a write. Maybe NewWriter should begin with a call to QueryWriteStatus.
func NewWriter(ctx context.Context, containerID string, stdIn bool, writeServer pbcontainers.Containers_AttachServer) (*Writer, error) {
	return &Writer{
		ctx:         ctx,
		writeServer: writeServer,
		containerID: containerID,
		stdIn:       stdIn,
	}, nil
}
