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

	"github.com/eclipse-kanto/container-management/containerm/log"
	"golang.org/x/sync/errgroup"
)

// Pipes is used to present any downstream pipe, for example, containerd's cio.
type Pipes struct {
	Stdin io.WriteCloser

	Stdout io.ReadCloser

	Stderr io.ReadCloser
}

// AttachConfig is used to describe how to attach the client's stream to
// the process's stream.
type AttachConfig struct {
	Terminal bool

	// CloseStdin means if the stdin of client's stream is closed by the
	// caller, the stdin of process's stream should be closed.
	CloseStdin bool

	// UseStdin/UseStdout/UseStderr can be used to check the client's stream
	// is nil or not. It is hard to check io.Write/io.ReadCloser != nil
	// directly, because they might be specific type, which means
	// (typ != nil) always is true.
	UseStdin, UseStdout, UseStderr bool

	Stdin          io.ReadCloser
	Stdout, Stderr io.Writer
}

// CopyPipes will watchs the data pipe's channel, like sticked to the pipe.
//
// NOTE: don't assign the specific type to the Pipes because the Std* != nil
// always return true.
func (s *cIOStream) CopyPipes(p Pipes) {
	copyfn := func(styp string, w io.WriteCloser, r io.ReadCloser) {
		s.Add(1)
		go func() {
			log.Debug("start to copy %s from pipe", styp)
			defer log.Debug("stop copy %s from pipe", styp)

			defer s.Done()
			defer r.Close()

			if _, err := io.Copy(w, r); err != nil {
				log.ErrorErr(err, "failed to copy pipe data")
			}
		}()
	}

	if p.Stdout != nil {
		copyfn("stdout", s.Stdout(), p.Stdout)
	}

	if p.Stderr != nil {
		copyfn("stderr", s.Stderr(), p.Stderr)
	}

	if s.stdin != nil && p.Stdin != nil {
		go func() {
			log.Debug("start to copy stdin from pipe")
			defer log.Debug("stop copy stdin from pipe")

			io.Copy(p.Stdin, s.stdin)
			if err := p.Stdin.Close(); err != nil {
				log.ErrorErr(err, "failed to close pipe stdin")
			}
		}()
	}
}

// Attach will use stream defined by AttachConfig to attach the Stream.
func (s *cIOStream) Attach(ctx context.Context, cfg *AttachConfig) <-chan error {
	var (
		group          errgroup.Group
		stdout, stderr io.ReadCloser
	)

	if cfg.UseStdin {
		group.Go(func() error {
			log.Debug("start to attach stdin to stream")
			defer log.Debug("stop attach stdin to stream")

			defer func() {
				if cfg.CloseStdin {
					s.StdinPipe().Close()
				}
			}()

			_, err := io.Copy(s.StdinPipe(), cfg.Stdin)
			if err == io.ErrClosedPipe {
				err = nil
			}
			return err
		})
	}

	attachFn := func(styp string, w io.Writer, r io.ReadCloser) error {
		log.Debug("start to attach %s to stream", styp)
		defer log.Debug("stop attach %s to stream", styp)

		defer func() {
			// NOTE: when the stdout/stderr is closed, the stdin
			// should be closed. for example, caller types the exit
			// command, the stdout will be closed. in this case,
			// the stdin should be closed. Otherwise, the caller
			// will wait for close signal forever.
			if cfg.UseStdin {
				cfg.Stdin.Close()
			}
			r.Close()
		}()

		_, err := io.Copy(w, r)
		if err == io.ErrClosedPipe {
			err = nil
		}
		return err
	}

	if cfg.UseStdout {
		stdout = s.NewStdoutPipe()
		group.Go(func() error {
			return attachFn("stdout", cfg.Stdout, stdout)
		})
	}

	if cfg.UseStderr {
		stderr = s.NewStderrPipe()
		group.Go(func() error {
			return attachFn("stderr", cfg.Stderr, stderr)
		})
	}

	var (
		errCh      = make(chan error, 1)
		groupErrCh = make(chan error, 1)
	)

	go func() {
		defer close(groupErrCh)
		groupErrCh <- group.Wait()
	}()

	go func() {
		defer log.Debug("the goroutine for attaching is done")
		defer close(errCh)

		select {
		case <-ctx.Done():
			if cfg.UseStdin {
				cfg.Stdin.Close()
			}

			// NOTE: the stdout writer will be evicted from stream in
			// next Write call.
			if cfg.UseStdout {
				stdout.Close()
			}

			// NOTE: the stderr writer will be evicted from stream in
			// next Write call.
			if cfg.UseStderr {
				stderr.Close()
			}

			if err := group.Wait(); err != nil {
				errCh <- err
				return
			}
			errCh <- ctx.Err()
		case err := <-groupErrCh:
			errCh <- err
		}
	}()
	return errCh
}
