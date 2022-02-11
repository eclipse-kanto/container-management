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

package ctr

import (
	"io"
	"time"

	"github.com/containerd/containerd/cio"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/logger"
	"github.com/eclipse-kanto/container-management/containerm/logger/buffered"
	"github.com/eclipse-kanto/container-management/containerm/streams"
	errutil "github.com/eclipse-kanto/container-management/containerm/util/error"
)

var (
	logcopierCloseTimeout = 10 * time.Second
	streamCloseTimeout    = 10 * time.Second
)

// wrapcio will wrap the DirectIO and IO.
//
// When the task exits, the containerd client will close the wrapcio.
type wrapcio struct {
	cio.IO

	ctrio IO
}

func (wcio *wrapcio) Wait() {
	wcio.IO.Wait()
	wcio.ctrio.Wait()
}

func (wcio *wrapcio) Close() error {
	wcio.IO.Close()

	return wcio.ctrio.Close()
}

// IO represents the streams and logger per identifiable process
type IO interface {
	// InitContainerIO will start logger and coping data from fifo
	InitContainerIO(dio *cio.DirectIO) (cio.IO, error)
	// SetLogDriver sets log driver to the IO
	SetLogDriver(logDriver logger.LogDriver)
	// SetMaxBufferSize set the max size of buffer.
	SetMaxBufferSize(maxBufferSize int64)
	// SetNonBlock whether to cache the container's logs with buffer
	SetNonBlock(nonBlock bool)
	// Stream is used to export the stream field
	Stream() streams.Stream
	// UseStdin returns whether the STDIN stream should also be attached
	UseStdin() bool
	// Wait wait for coping-data job
	Wait()
	// Close closes the stream and the logger
	Close() error
	// Reset resets the allocated streams
	Reset()
}

type containerIO struct {
	// currently, the IO instances are used only to handler he streams of a container - not an exec process inside it
	// at a later stage - if exec is to be supported, the struct must be enhanced with an additional id filed
	// for either to process or the container - as it has to be possible to identify to which container the exec belongs
	id       string
	useStdin bool
	stream   streams.Stream

	logDriver  logger.LogDriver
	logHandler logger.LogHandler

	nonBlock      bool
	maxBufferSize int64
}

// NewIO return IO instance.
func newIO(id string, withStdin bool) IO {
	s := streams.NewStream()
	if withStdin {
		s.NewStdinInput()
	} else {
		s.NewDiscardStdinInput()
	}

	return &containerIO{
		id:       id,
		useStdin: withStdin,
		stream:   s,
	}
}

// Reset reset the logDriver.
func (ctrio *containerIO) Reset() {
	if err := ctrio.Close(); err != nil {
		log.WarnErr(err, "failed to close during reset IO")
	}

	if ctrio.useStdin {
		ctrio.stream.NewStdinInput()
	} else {
		ctrio.stream.NewDiscardStdinInput()
	}

	ctrio.logDriver = nil
	ctrio.logHandler = nil
}

// SetLogDriver sets log driver to the IO.
func (ctrio *containerIO) SetLogDriver(logdriver logger.LogDriver) {
	ctrio.logDriver = logdriver
}

// SetMaxBufferSize set the max size of buffer.
func (ctrio *containerIO) SetMaxBufferSize(maxBufferSize int64) {
	ctrio.maxBufferSize = maxBufferSize
}

// SetNonBlock whether to cache the container's logs with buffer.
func (ctrio *containerIO) SetNonBlock(nonBlock bool) {
	ctrio.nonBlock = nonBlock
}

// Stream is used to export the stream field.
func (ctrio *containerIO) Stream() streams.Stream {
	return ctrio.stream
}

// UseStdin returns whether the STDIN stream should also be attached
func (ctrio *containerIO) UseStdin() bool {
	return ctrio.useStdin
}

// Wait wait for coping-data job.
func (ctrio *containerIO) Wait() {
	waitCh := make(chan struct{})
	go func() {
		defer close(waitCh)
		ctrio.stream.Wait()
	}()

	select {
	case <-waitCh:
	case <-time.After(streamCloseTimeout):
		log.Warn("stream doesn't exit in time")
	}
}

// Close closes the stream and the logger.
func (ctrio *containerIO) Close() error {
	compoundErr := new(errutil.CompoundError)

	ctrio.Wait()
	if err := ctrio.stream.Close(); err != nil {
		compoundErr.Append(err)
	}

	if ctrio.logDriver != nil {
		if ctrio.logHandler != nil {
			waitChannel := make(chan struct{})
			go func() {
				defer close(waitChannel)
				ctrio.logHandler.Wait()
			}()
			select {
			case <-waitChannel:
			case <-time.After(logcopierCloseTimeout):
				log.Warn("logHandler doesn't exit in time")
			}
		}

		if err := ctrio.logDriver.Close(); err != nil {
			compoundErr.Append(err)
		}
	}

	if compoundErr.Size() > 0 {
		return compoundErr
	}
	return nil
}

// InitContainerIO will start logger and coping data from fifo.
func (ctrio *containerIO) InitContainerIO(dio *cio.DirectIO) (cio.IO, error) {
	if err := ctrio.startLogging(); err != nil {
		return nil, err
	}

	ctrio.stream.CopyPipes(streams.Pipes{
		Stdin:  dio.Stdin,
		Stdout: dio.Stdout,
		Stderr: dio.Stderr,
	})
	return &wrapcio{IO: dio, ctrio: ctrio}, nil
}
func (ctrio *containerIO) startLogging() error {
	if ctrio.logDriver == nil {
		return nil
	}

	if ctrio.nonBlock {
		logDriver, err := buffered.NewBufferedLog(ctrio.logDriver, ctrio.maxBufferSize)
		if err != nil {
			return err
		}
		ctrio.logDriver = logDriver
	}

	ctrio.logHandler = logger.NewLogHandler(ctrio.logDriver, map[string]io.Reader{
		"stdout": ctrio.stream.NewStdoutPipe(),
		"stderr": ctrio.stream.NewStderrPipe(),
	})
	ctrio.logHandler.StartCopyToLogDriver()
	return nil
}
