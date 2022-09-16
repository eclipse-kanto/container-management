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

package mgr

import (
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

// the cache saves all container' restart managers.
type restartMgrCache struct {
	m *sync.Map
}

// newRestartMgrCache creates a container's restart managers storage.
func newRestartMgrCache() *restartMgrCache {
	return &restartMgrCache{
		m: &sync.Map{},
	}
}

// put writes a container's restart manager into storage.
func (c *restartMgrCache) put(id string, resMan *restartManager) error {
	c.m.Store(id, resMan)
	log.Debug("added restartManager for container id = %s", id)
	return nil
}

// get reads a container's restart manager by id.
func (c *restartMgrCache) get(id string) *restartManager {
	obj, ok := c.m.Load(id)
	if !ok {
		log.Debug("could not load restartManager for container id = %s", id)
		return nil
	}

	if resMan, ok := obj.(*restartManager); ok {
		log.Debug("found restartManager for container id = %s", id)
		return resMan
	}
	log.Debug("could not find restartManager for container id = %s", id)
	return nil
}

// remove removes the container's restart manager.
func (c *restartMgrCache) remove(id string) {
	c.m.Delete(id)
	log.Debug("removed restartManager for container id = %s", id)
}
