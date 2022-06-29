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

// Package name and imports changed also removed not needed logic and added custom code to handle the specific use case, Bosch.IO GmbH, 2020

package client

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
	readClient  pbcontainers.Containers_AttachClient
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
		resp, err := r.readClient.Recv()
		if err != nil {
			r.err = err
			return 0, err
		}
		r.buf = resp.ReadData
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
	if r.readClient == nil {
		return nil
	}
	err := r.readClient.CloseSend()
	r.readClient = nil
	return err
}

// NewReader creates a new Reader to read a resource.
func NewReader(ctx context.Context, containerID string, stdIn bool, ctrClient pbcontainers.Containers_AttachClient) (*Reader, error) {
	return NewReaderAt(ctx, containerID, stdIn, 0, ctrClient)
}

// NewReaderAt creates a new Reader to read a resource from the given offset.
func NewReaderAt(ctx context.Context, containerID string, stdIn bool, offset int64, ctrClient pbcontainers.Containers_AttachClient) (*Reader, error) {
	// readClient is set up for Read(). ReadAt() will copy needed fields into its reentrantReader.

	return &Reader{
		ctx:         ctx,
		containerID: containerID,
		stdIn:       stdIn,
		readClient:  ctrClient,
	}, nil
}

// Writer writes to a byte stream.
type Writer struct {
	ctx         context.Context
	writeClient pbcontainers.Containers_AttachClient
	containerID string
	stdIn       bool
	offset      int64
	err         error
}

// ContainerID gets the resource name this Writer is writing.
func (w *Writer) ContainerID() string {
	return w.containerID
}

// StdIn gets the resource name this Writer is writing.
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
		r := pbcontainers.AttachContainerRequest{
			WriteOffset: w.offset,
			FinishWrite: false,
			DataToWrite: p[n : n+bufSize],
		}
		// Bytestream only requires the resourceName to be sent in the first WriteRequest.
		if w.offset == 0 {
			r.Id = w.containerID
			r.StdIn = w.stdIn
		}
		err := w.writeClient.Send(&r)
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
	err := w.writeClient.Send(&pbcontainers.AttachContainerRequest{
		Id:          w.containerID,
		WriteOffset: w.offset,
		FinishWrite: true,
		DataToWrite: nil,
	})
	if err != nil {
		w.err = err
		return fmt.Errorf("Send(WriteRequest< FinishWrite >) failed: %v", err)
	}
	resp, err := w.writeClient.Recv()
	if err != nil {
		w.err = err
		return fmt.Errorf("CloseAndRecv: %v", err)
	}
	if resp == nil {
		err = fmt.Errorf("expected a response on close, got %v", resp)
	} else if resp.WriteCommittedSize != w.offset {
		err = fmt.Errorf("server only wrote %d bytes, want %d", resp.WriteCommittedSize, w.offset)
	}
	err = w.writeClient.CloseSend()
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
// Note: There is currently no way to resume a writer. Maybe NewWriter should begin with a call to QueryWriteStatus.
func NewWriter(ctx context.Context, containerID string, stdIn bool, ctrClient pbcontainers.Containers_AttachClient) (*Writer, error) {
	return &Writer{
		ctx:         ctx,
		writeClient: ctrClient,
		containerID: containerID,
		stdIn:       stdIn,
	}, nil
}
