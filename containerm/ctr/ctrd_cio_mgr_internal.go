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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/containerd/containerd/cio"
	"github.com/eclipse-kanto/container-management/containerm/log"
	ioutils "github.com/eclipse-kanto/container-management/containerm/util/io"
)

// newFIFOSet prepares fifo files.
func (mgr *cioMgr) newFIFOSet(processID string, withStdin bool, withTerminal bool) (*cio.FIFOSet, error) {
	//root := "/run/container-management/fifo"
	if err := os.MkdirAll(mgr.fifoRootDir, 0700); err != nil {
		return nil, err
	}

	fifoDir, err := os.MkdirTemp(mgr.fifoRootDir, "")
	if err != nil {
		return nil, err
	}

	cfg := cio.Config{
		Terminal: withTerminal,
		Stdout:   filepath.Join(fifoDir, processID+"-stdout"),
	}

	if withStdin {
		cfg.Stdin = filepath.Join(fifoDir, processID+"-stdin")
	}

	if !withTerminal {
		cfg.Stderr = filepath.Join(fifoDir, processID+"-stderr")
	}

	closeFn := func() error {
		err := os.RemoveAll(fifoDir)
		if err != nil {
			log.WarnErr(err, "failed to remove process(id=%v) fifo dir", processID)
		}
		return err
	}

	return cio.NewFIFOSet(cfg, closeFn), nil
}

func (mgr *cioMgr) createIO(fifoSet *cio.FIFOSet, containerIO IO) (cio.IO, error) {
	cdio, err := cio.NewDirectIO(context.Background(), fifoSet)
	if err != nil {
		return nil, err
	}

	if cdio.Stdin != nil {
		var (
			errClose  error
			stdinOnce sync.Once
		)
		oldStdin := cdio.Stdin
		cdio.Stdin = ioutils.NewWriteCloserWrapper(oldStdin, func() error {
			stdinOnce.Do(func() {
				errClose = oldStdin.Close()
			})
			return errClose
		})
	}

	cntrio, err := containerIO.InitContainerIO(cdio)
	if err != nil {
		cdio.Cancel()
		cdio.Close()
		return nil, err
	}
	return cntrio, nil
}

/*
// keeping the createIO exec handling ready
func (mgr *cioMgr) createIOExec(fifoSet *cio.FIFOSet, cntrID, procID string, closeStdinCh <-chan struct{}, procIOCloser execProcessIOCloser, containerIO *IO) (cio.IO, error) {
	cdio, err := cio.NewDirectIO(context.Background(), fifoSet)
	if err != nil {
		return nil, err
	}

	if cdio.Stdin != nil {
		var (
			errClose  error
			stdinOnce sync.Once
		)
		oldStdin := cdio.Stdin
		cdio.Stdin = ioutils.NewWriteCloserWrapper(oldStdin, func() error {
			stdinOnce.Do(func() {
				errClose = oldStdin.Close()

				// Both the caller and container/exec process holds write side pipe
				// for the stdin. When the caller closes the write pipe, the process doesn't
				// exit until the caller calls the CloseIO.
				go func() {
					<-closeStdinCh
					if err := procIOCloser(cntrID, procID); err != nil {
						// for the CloseIO grpc call, the containerd doesn't
						// return correct status code if the process doesn't exist.
						// for the case, we should use strings.Contains to reduce warning
						// log. it will be fixed in containerd#2747.
						if !strings.Contains(err.Error(), "not found") {
							log.WarnErr(err, "failed to close stdin containerd IO (container:%v, process:%v", cntrID, procID)
						}
					}
				}()
			})
			return errClose
		})
	}

	cntrio, err := containerIO.InitContainerIO(cdio)
	if err != nil {
		cdio.Cancel()
		cdio.Close()
		return nil, err
	}
	return cntrio, nil
}*/

func (mgr *cioMgr) attachIO(fifoSet *cio.FIFOSet, id string) (cio.IO, error) {
	ctrIO := mgr.ioCache.Get(id)
	if ctrIO == nil {
		return nil, log.NewErrorf("no IO resources allocated for id = %s", id)
	}

	if fifoSet == nil {
		return nil, fmt.Errorf("cannot attach to existing fifos")
	}

	cdio, cioErr := cio.NewDirectIO(context.Background(), &cio.FIFOSet{
		Config: cio.Config{
			Terminal: fifoSet.Terminal,
			Stdin:    fifoSet.Stdin,
			Stdout:   fifoSet.Stdout,
			Stderr:   fifoSet.Stderr,
		},
	})
	if cioErr != nil {
		return nil, cioErr
	}

	cntrio, err := ctrIO.InitContainerIO(cdio)
	if err != nil {
		cdio.Cancel()
		cdio.Close()
		return nil, err
	}
	return cntrio, nil
}
