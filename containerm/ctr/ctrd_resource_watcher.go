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

type resourcesWatcher interface {
	Watch(resourceID string, duration time.Duration, expiredHandler watchExpired) error
	Dispose()
}

var errAlreadyWatched = log.NewError("resource already watched")

type watchExpired func(ctx context.Context, id string) error

type watchInfo struct {
	resourceID     string
	timer          *time.Timer
	expiredHandler watchExpired
}

type resWatcher struct {
	sync.Mutex
	watchCache       map[string]watchInfo
	watchCacheLock   sync.RWMutex
	watcherCtx       context.Context
	watcherCtxCancel context.CancelFunc
}

func newResourcesWatcher(ctx context.Context) resourcesWatcher {
	watcher := &resWatcher{
		watchCache:     make(map[string]watchInfo),
		watchCacheLock: sync.RWMutex{},
	}
	watcher.watcherCtx, watcher.watcherCtxCancel = context.WithCancel(ctx)
	return watcher
}

func (watcher *resWatcher) Watch(resourceID string, duration time.Duration, expiredHandler watchExpired) error {
	watcher.Lock()
	defer watcher.Unlock()
	if err := watcher.watcherCtx.Err(); err == context.Canceled {
		log.Debug("resource manager is cancelled - will not process watch request for resource = %s", resourceID)
		return err
	}
	watcher.watchCacheLock.RLock()
	defer watcher.watchCacheLock.RUnlock()
	log.Debug("processing monitoring requested for resource %s", resourceID)

	if _, ok := watcher.watchCache[resourceID]; ok {
		return errAlreadyWatched
	}

	info := watchInfo{
		resourceID:     resourceID,
		timer:          time.NewTimer(duration),
		expiredHandler: expiredHandler,
	}
	watcher.watchCache[info.resourceID] = info
	go func(ctx context.Context, info watchInfo) {
		select {
		case <-info.timer.C:
			if info.expiredHandler != nil {
				if err := info.expiredHandler(ctx, info.resourceID); err != nil {
					log.WarnErr(err, "error while handling monitoring expiry for resource %s", info.resourceID)
				}
			}
			watcher.cleanCache(info.resourceID)
			log.Debug("successfully processed expired resource %s", info.resourceID)
		case <-ctx.Done():
			log.Debug("cancelled monitoring for resource %s", info.resourceID)
		}
		log.Debug("finished watch process for resource %s", info.resourceID)
	}(watcher.watcherCtx, info)
	log.Debug("successfully scheduled expiry monitoring for resource %s", info.resourceID)
	return nil
}

func (watcher *resWatcher) Dispose() {
	watcher.Lock()
	defer watcher.Unlock()
	watcher.watcherCtxCancel()

	watcher.watchCacheLock.RLock()
	defer watcher.watchCacheLock.RUnlock()

	for infoKey, info := range watcher.watchCache {
		log.Debug("stopping monitoring for resource %s", infoKey)
		info.timer.Stop()
	}
}
func (watcher *resWatcher) cleanCache(id string) {
	watcher.watchCacheLock.Lock()
	defer watcher.watchCacheLock.Unlock()
	info, ok := watcher.watchCache[id]
	if ok {
		delete(watcher.watchCache, id)
		log.Debug("removed watch cache for resource %s", info.resourceID)
	} else {
		log.Debug("no watch cache to remove for resource %s", info.resourceID)
	}
}
