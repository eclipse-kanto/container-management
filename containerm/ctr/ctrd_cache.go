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

package ctr

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/containerd/containerd"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

type exitInfo struct {
	exitCode  uint32
	exitError error
	exitTime  time.Time
}

type containerInfo struct {
	c             *types.Container
	container     containerd.Container
	task          containerd.Task
	statusChannel <-chan containerd.ExitStatus
	resultChannel chan exitInfo

	skipExitHooks bool
	oomKilled     bool

	ctrInfoLock sync.RWMutex
}

func (ctrInfo *containerInfo) getTask() containerd.Task {
	ctrInfo.ctrInfoLock.RLock()
	defer ctrInfo.ctrInfoLock.RUnlock()
	return ctrInfo.task
}

func (ctrInfo *containerInfo) setTask(newTask containerd.Task) {
	ctrInfo.ctrInfoLock.Lock()
	defer ctrInfo.ctrInfoLock.Unlock()
	ctrInfo.task = newTask
}

type containerInfoCache struct {
	sync.Mutex
	containerdStopped  bool
	cache              map[string]*containerInfo
	containerExitHooks []ContainerExitHook
}

func newContainerInfoCache() *containerInfoCache {
	return &containerInfoCache{
		cache: make(map[string]*containerInfo),
	}
}

func (cacheMgr *containerInfoCache) setExitHooks(hooks ...ContainerExitHook) {
	cacheMgr.Lock()
	defer cacheMgr.Unlock()
	cacheMgr.containerExitHooks = hooks
}

func (ctrInfo *containerInfo) isOOmKilled() bool {
	ctrInfo.ctrInfoLock.RLock()
	defer ctrInfo.ctrInfoLock.RUnlock()
	return ctrInfo.oomKilled
}

func (ctrInfo *containerInfo) setOOMKilled(oomKilled bool) {
	ctrInfo.ctrInfoLock.Lock()
	defer ctrInfo.ctrInfoLock.Unlock()
	ctrInfo.oomKilled = oomKilled
}

func (cacheMgr *containerInfoCache) add(info *containerInfo) {
	cacheMgr.Lock()
	defer cacheMgr.Unlock()
	cacheMgr.cache[info.container.ID()] = info

	go func(cMgr *containerInfoCache, ctrInfo *containerInfo) {
		log.Debug("started watching go func for container id = %s", ctrInfo.c.ID)
		defer log.Debug("releasing cacheMgr's watch goroutine for container %s", ctrInfo.c.ID)

		status := <-ctrInfo.statusChannel
		log.Debug("recevied exit status from container id = %s", ctrInfo.c.ID)

		// connection to ctrd has been lost - exit immediately
		log.Debug("checking if containerd is available")
		if cMgr.isContainerdDead() {
			log.Warn("containerd is not available - exiting all go watch funcs")
			return
		}

		log.Debug("checking if the channel of the container is closed")
		if isContainerdConnectable(status) {
			log.Warn("received exit message with a broken channel identifying that connection to containerd is lost, %+v", status)
			return
		}

		log.Debug("processing exit status")
		code, exitTime, exitErr := status.Result()
		if exitErr != nil {
			log.ErrorErr(exitErr, "error while exiting container id = %s", ctrInfo.c.ID)
		}

		var (
			cleanupOnce sync.Once
			err         error
		)
		cleanupFunc := func() error {
			cleanupOnce.Do(func() {
				log.Debug("container %s exited with code %d", ctrInfo.c.ID, code)

				ctrInfo.ctrInfoLock.Lock()
				defer ctrInfo.ctrInfoLock.Unlock()

				ctx := context.Background()
				log.Debug("closing task's IOs now to release the streams for container id = %s", ctrInfo.c.ID)
				err = ctrInfo.task.CloseIO(ctx, containerd.WithStdinCloser)
				if err != nil {
					log.ErrorErr(err, "could not close task IOs for container id = %s", ctrInfo.c.ID)
				}
				log.Debug("deleting task for container id = %s", ctrInfo.c.ID)
				_, err = ctrInfo.task.Delete(ctx)
				if err != nil {
					log.ErrorErr(err, "error while deleting task for container id = %s, container delete was not successful ", ctrInfo.c.ID)
				} else {
					ctrInfo.task = nil
					log.Debug("deleted the task for container id = %s successfully", ctrInfo.c.ID)
				}

				log.Debug("deleting container for container id = %s", ctrInfo.c.ID)
				err = ctrInfo.container.Delete(ctx)
				if err != nil {
					log.ErrorErr(err, "error while deleting containerd container for container id = %s, container delete was not successful ", ctrInfo.c.ID)
				} else {
					ctrInfo.container = nil
					log.Debug("deleted the underlying container for container id = %s successfully", ctrInfo.c.ID)
				}
			})
			return nil
		}
		ctrInfo.ctrInfoLock.RLock()
		skipCleanup := ctrInfo.skipExitHooks
		ctrInfo.ctrInfoLock.RUnlock()
		if !skipCleanup {
			log.Debug("calling all exit hooks")
			for _, hook := range cMgr.containerExitHooks {
				if err = hook(ctrInfo.c, int64(code), exitErr, ctrInfo.isOOmKilled(), cleanupFunc); err != nil {
					log.ErrorErr(err, "failed to execute exit hook for container %s", ctrInfo.c.ID)
					break
				}
				log.Debug("successfully called exit hook")
			}
			cleanupFunc()
		}

		select {
		case ctrInfo.resultChannel <- exitInfo{
			exitCode:  code,
			exitError: exitErr,
			exitTime:  exitTime,
		}:
			log.Debug("sending exit result of the watch goroutine for container id = %s", ctrInfo.c.ID)
		default:
			log.Debug("process exited internally - no exit result of the watch goroutine for container id = %s will be sent", ctrInfo.c.ID)
		}

	}(cacheMgr, info)
}

func (cacheMgr *containerInfoCache) remove(id string) *containerInfo {
	cacheMgr.Lock()
	defer cacheMgr.Unlock()
	ctrInfo, ok := cacheMgr.cache[id]
	if ok {
		delete(cacheMgr.cache, id)
		return ctrInfo
	}
	return nil
}

func (cacheMgr *containerInfoCache) get(id string) *containerInfo {
	cacheMgr.Lock()
	defer cacheMgr.Unlock()
	ctrInfo, ok := cacheMgr.cache[id]
	if !ok {
		return nil
	}
	return ctrInfo
}

func (cacheMgr *containerInfoCache) getAll() []*containerInfo {
	cacheMgr.Lock()
	defer cacheMgr.Unlock()
	var res []*containerInfo
	for _, ctrInfo := range cacheMgr.cache {
		res = append(res, ctrInfo)
	}
	return res
}

func (cacheMgr *containerInfoCache) setContainerdDead(isDead bool) error {
	cacheMgr.Lock()
	defer cacheMgr.Unlock()

	cacheMgr.containerdStopped = isDead
	return nil
}

func (cacheMgr *containerInfoCache) isContainerdDead() bool {
	cacheMgr.Lock()
	defer cacheMgr.Unlock()

	return cacheMgr.containerdStopped
}

// isContainerdConnectable identifies that the connection to containerd has failed unexpectedly
func isContainerdConnectable(ctrdExitStatus containerd.ExitStatus) bool {
	if ctrdExitStatus.Error() == nil {
		return false
	}

	rpcError := strings.Contains(ctrdExitStatus.Error().Error(), "transport is closing") ||
		strings.Contains(ctrdExitStatus.Error().Error(), "rpc error")

	return ctrdExitStatus.ExitTime().IsZero() && rpcError
}
