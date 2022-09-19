// Copyright (c) 2021 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0

package ctr

import (
	"github.com/containerd/containerd/cio"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/logger"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

// execProcessIOCloser is reponsible for closing the STDIN pipes for processed loaded inside a container in an interactive manner
//type execProcessIOCloser func(containerID, processID string) error

// containerIOManager is responsible for handling all IO resources per container
type containerIOManager interface {
	// ExistsIO checks whether IO resources are already allocated for this id
	ExistsIO(id string) bool
	// Get returns an existing IO instance for the provided id
	GetIO(id string) IO
	// InitIO creates a new IO for the provided container id
	InitIO(id string, withStdin bool) (IO, error)
	// InitExecIO creates a new IO for a new process executed within a container
	//InitExecIO(id string, withStdin bool) (IO, error)
	// ConfigureIO configures the associated IO resources with the provided logging information
	ConfigureIO(id string, logDriver logger.LogDriver, logModeCfg *types.LogModeConfiguration) error
	// ResetIO resets the IO resources for the provided id
	ResetIO(id string)
	// CloseIO closes the IO resources for the provided id
	CloseIO(id string) error
	// ClearIO closes and removes an existing IO instance for the provided id
	ClearIO(id string) error
	// NewCioCreator creates a new IO set for the provided container id
	NewCioCreator(withTerminal bool) cio.Creator
	// NewCioCreator creates a new IO set for the provided id that refers to a process in the provided container id
	// NewCioCreatorExec(containerID string, withTerminal bool, closeStdinCh <-chan struct{}, procIOCloser execProcessIOCloser) cio.Creator
	// NewCioAttach creates a new IO set for attaching to an existing root container process
	NewCioAttach(id string) cio.Attach
}

type cioMgr struct {
	fifoRootDir string
	ioCache     *cioCache
}

// newContainerIOManager creates a new containerIOManager instance
func newContainerIOManager(fifoRootDir string, ioCache *cioCache) containerIOManager {
	return &cioMgr{
		fifoRootDir: fifoRootDir,
		ioCache:     ioCache,
	}
}
func (mgr *cioMgr) ExistsIO(id string) bool {
	return mgr.ioCache.Get(id) != nil
}

func (mgr *cioMgr) GetIO(id string) IO {
	return mgr.ioCache.Get(id)
}

func (mgr *cioMgr) InitIO(id string, withStdin bool) (IO, error) {
	if io := mgr.ioCache.Get(id); io != nil {
		return nil, log.NewErrorf("failed to create containerIO")
	}
	cntrio := newIO(id, withStdin)
	mgr.ioCache.Put(id, cntrio)
	return cntrio, nil
}

// Keeping the code base interactive exec processes execution ready
//func (mgr *cioMgr) InitExecIO(id string, withStdin bool) (IO, error) {
//	if io := mgr.ioCache.Get(id); io != nil {
//		return nil, log.NewError("failed to create containerIO")
//	}
//	cntrio := newIO(id, withStdin)
//	mgr.ioCache.Put(id, cntrio)
//	return cntrio, nil
//}

func (mgr *cioMgr) ConfigureIO(id string, logDriver logger.LogDriver, logModeCfg *types.LogModeConfiguration) error {
	ctrIO := mgr.ioCache.Get(id)
	if ctrIO == nil {
		return log.NewErrorf("no IO resources allocated for id = %s", id)
	}

	if logModeCfg.Mode == types.LogModeNonBlocking {
		bytes, err := util.SizeToBytes(logModeCfg.MaxBufferSize)
		if err != nil {
			return err
		}
		if bytes > 0 {
			ctrIO.SetMaxBufferSize(bytes)
			ctrIO.SetNonBlock(true)
		}
	}
	ctrIO.SetLogDriver(logDriver)

	return nil
}

func (mgr *cioMgr) ResetIO(id string) {
	// release resource
	log.Debug("will reset container IOs for container id = %s", id)
	io := mgr.ioCache.Get(id)
	if io == nil {
		log.Debug("no IOs for container id = %s", id)
		return
	}
	log.Debug("performing reset IOs for container id = %s", id)
	io.Reset()
}

func (mgr *cioMgr) CloseIO(id string) error {
	// release resource
	log.Debug("will close container IOs for container id = %s", id)
	io := mgr.ioCache.Get(id)
	if io == nil {
		log.Debug("no IOs for container id = %s", id)
		return nil
	}
	log.Debug("performing close IOs for container id = %s", id)
	return io.Close()
}

func (mgr *cioMgr) ClearIO(id string) error {
	if err := mgr.CloseIO(id); err != nil {
		return err
	}
	mgr.ioCache.Remove(id)
	return nil
}

func (mgr *cioMgr) NewCioCreator(withTerminal bool) cio.Creator {
	return func(id string) (cio.IO, error) {
		ctrIO := mgr.ioCache.Get(id)
		if ctrIO == nil {
			return nil, log.NewErrorf("no IO resources allocated for id = %s", id)
		}
		log.Debug("creating cio for container ID = %s (withStdin=%v, withTerminal=%v)", id, ctrIO.UseStdin(), withTerminal)
		fifoSet, err := mgr.newFIFOSet(id, ctrIO.UseStdin(), withTerminal)
		if err != nil {
			return nil, err
		}
		return mgr.createIO(fifoSet, ctrIO)
	}
}

// keeping it for future interactive exec support
//func (mgr *cioMgr) NewCioCreatorExec(containerID string, withTerminal bool, closeStdinCh <-chan struct{}, procIOCloser execProcessIOCloser) cio.Creator {
//	log.Warn("NewCioCreatorExec is not supported")
//	return nil
//}

func (mgr *cioMgr) NewCioAttach(id string) cio.Attach {
	return func(fset *cio.FIFOSet) (cio.IO, error) {
		return mgr.attachIO(fset, id)
	}
}
