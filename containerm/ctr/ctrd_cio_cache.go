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

import "sync"

// Cache saves the all container's io.
type cioCache struct {
	m *sync.Map
}

// нewCache creates a container's io storage.
func newCache() *cioCache {
	return &cioCache{
		m: &sync.Map{},
	}
}

// Put writes a container's io into storage.
func (c *cioCache) Put(id string, io IO) error {
	c.m.Store(id, io)
	return nil
}

// Get reads a container's io by id.
func (c *cioCache) Get(id string) IO {
	obj, ok := c.m.Load(id)
	if !ok {
		return nil
	}

	if io, ok := obj.(IO); ok {
		return io
	}
	return nil
}

// Remove removes the container's io.
func (c *cioCache) Remove(id string) {
	c.m.Delete(id)
}
