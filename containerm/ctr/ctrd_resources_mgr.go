// Copyright (c) 2022 Contributors to the Eclipse Foundation
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
	"github.com/eclipse-kanto/container-management/containerm/log"
	"sync"
	"time"
)

type resourcesManager interface {
	Watch(resourceID string, duration time.Duration, expiredHandler watchExpired) error
	Dispose()
}

var alreadyWatchedError = log.NewError("resource already watched")

type watchExpired func(ctx context.Context, id string) error

type resourcesInfo struct {
	id             string
	timer          *time.Timer
	expiredHandler watchExpired
}

type resMgr struct {
	watchCache     map[string]resourcesInfo
	watchCacheLock sync.RWMutex
	mgrCtx         context.Context
	mgrCtxCancel   context.CancelFunc
}

func newResourceManager() resourcesManager {
	mgr := &resMgr{
		watchCache:     make(map[string]resourcesInfo),
		watchCacheLock: sync.RWMutex{},
	}
	mgr.mgrCtx, mgr.mgrCtxCancel = context.WithCancel(context.Background())
	return mgr
}

func (resMrg *resMgr) Watch(resourceID string, duration time.Duration, expiredHandler watchExpired) error {
	if err := resMrg.mgrCtx.Err(); err == context.Canceled {
		log.Debug("resource manager is cancelled - will not process watch request for resource = %s", resourceID)
		return err
	}
	resMrg.watchCacheLock.RLock()
	defer resMrg.watchCacheLock.RUnlock()

	if _, ok := resMrg.watchCache[resourceID]; ok {
		return alreadyWatchedError
	}

	info := resourcesInfo{
		id:             resourceID,
		timer:          time.NewTimer(duration),
		expiredHandler: expiredHandler,
	}
	resMrg.watchCache[info.id] = info
	go func(ctx context.Context, info resourcesInfo) {
		select {
		case <-info.timer.C:
			if info.expiredHandler != nil {
				if err := info.expiredHandler(ctx, info.id); err != nil {
					log.WarnErr(err, "could not perform resource clean up for resource %s", info.id)
				}
			}
			resMrg.cleanCache(info.id)
		case <-ctx.Done():
			log.Debug("cancelled monitoring for resource %s", info.id)
		}
	}(resMrg.mgrCtx, info)
	return nil
}

func (resMrg *resMgr) Dispose() {
	resMrg.mgrCtxCancel()
	resMrg.watchCacheLock.RLock()
	defer resMrg.watchCacheLock.RUnlock()

	for infoKey, info := range resMrg.watchCache {
		log.Debug("stopping monitoring for resource %s", infoKey)
		info.timer.Stop()
	}
}
func (resMrg *resMgr) cleanCache(id string) {
	resMrg.watchCacheLock.Lock()
	defer resMrg.watchCacheLock.Unlock()
	info, ok := resMrg.watchCache[id]
	if ok {
		delete(resMrg.watchCache, id)
		log.Debug("removed watch cache for resource %s", info.id)
	} else {
		log.Debug("no watch cache to remove for resource %s", info.id)
	}
}
