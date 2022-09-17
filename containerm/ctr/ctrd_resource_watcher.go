// Copyright (c) 2022 Contributors to the Eclipse Foundation
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
	watchCache          map[string]watchInfo
	watchCacheLock      sync.RWMutex
	watcherCtx          context.Context
	watcherCtxCancel    context.CancelFunc
	watchCacheWaitGroup *sync.WaitGroup
}

func newResourcesWatcher(ctx context.Context) resourcesWatcher {
	watcher := &resWatcher{
		watchCache:          make(map[string]watchInfo),
		watchCacheLock:      sync.RWMutex{},
		watchCacheWaitGroup: &sync.WaitGroup{},
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
	watcher.watchCacheWaitGroup.Add(1)
	go func(ctx context.Context, info watchInfo) {
		defer watcher.watchCacheWaitGroup.Done()
		select {
		case <-info.timer.C:
			if info.expiredHandler != nil {
				if err := info.expiredHandler(ctx, info.resourceID); err != nil {
					log.WarnErr(err, "error while handling monitoring expiry for resource %s", info.resourceID)
				}
			}
			watcher.cleanCache(info.resourceID, false)
			log.Debug("successfully processed expired resource %s", info.resourceID)
		case <-ctx.Done():
			watcher.cleanCache(info.resourceID, true)
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
	log.Debug("resource watcher is disposing")
	watcher.watcherCtxCancel()

	log.Debug("waiting for monitoring routines to finish")
	watcher.watchCacheWaitGroup.Wait()

	log.Debug("resource watcher disposed")
}
func (watcher *resWatcher) cleanCache(id string, withStop bool) {
	watcher.watchCacheLock.Lock()
	defer watcher.watchCacheLock.Unlock()
	info, ok := watcher.watchCache[id]
	if ok {
		if withStop && info.timer.Stop() {
			log.Debug("stopped monitoring timer for resource %s", info.resourceID)
		}
		delete(watcher.watchCache, id)
		log.Debug("removed watch cache for resource %s", info.resourceID)
	} else {
		log.Warn("no watch cache found for resource %s", info.resourceID)
	}
}
