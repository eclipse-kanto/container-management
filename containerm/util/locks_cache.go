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

package util

import "sync"

// LocksCache provides cache management for identifiable *sync.RWMutex instances.
type LocksCache struct {
	locks     map[string]*sync.RWMutex
	cacheLock sync.RWMutex
}

// NewLocksCache creates a new LockCache
func NewLocksCache() LocksCache {
	cache := make(map[string]*sync.RWMutex)
	return LocksCache{
		locks:     cache,
		cacheLock: sync.RWMutex{},
	}
}

// GetLock returns the lock by provided key
func (cache *LocksCache) GetLock(key string) *sync.RWMutex {
	cache.cacheLock.Lock()
	defer cache.cacheLock.Unlock()

	if val, ok := cache.locks[key]; ok {
		return val
	}

	nLock := sync.RWMutex{}
	cache.locks[key] = &nLock
	return &nLock
}

// RemoveLock removes the lock by provided key
func (cache *LocksCache) RemoveLock(key string) {
	cache.cacheLock.Lock()
	defer cache.cacheLock.Unlock()

	if _, ok := cache.locks[key]; ok {
		delete(cache.locks, key)
	}
}
