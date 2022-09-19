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

package registry

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
)

// ServiceRegistryContext holds information for the service's registry context
type ServiceRegistryContext struct {
	Context      context.Context
	Config       interface{}
	Registration *Registration
	services     *Set
}

// NewContext returns a new registry ServiceRegistryContext
func NewContext(ctx context.Context, config interface{}, r *Registration, services *Set) *ServiceRegistryContext {
	return &ServiceRegistryContext{
		Context:      ctx,
		Config:       config,
		Registration: r,
		services:     services,
	}
}

// Get returns the first service by its type
func (i *ServiceRegistryContext) Get(t Type) (interface{}, error) {
	return i.services.Get(t)
}

// GetAll services in the set
func (i *ServiceRegistryContext) GetAll() []*ServiceInfo {
	return i.services.ordered
}

// GetByType returns all services with the specific type.
func (i *ServiceRegistryContext) GetByType(t Type) (map[string]*ServiceInfo, error) {
	p, ok := i.services.byTypeAndID[t]
	if !ok {
		return nil, errors.Wrapf(errdefs.ErrNotFound, "no services registered for %s", t)
	}

	return p, nil
}

// ServiceInfo holds service information
type ServiceInfo struct {
	err          error
	instance     interface{}
	Config       interface{}
	Registration *Registration
}

// Err returns the errors during initialization.
// returns nil if not error was encountered
func (p *ServiceInfo) Err() error {
	return p.err
}

// Instance returns the instance and any initialization error of the plugin
func (p *ServiceInfo) Instance() (interface{}, error) {
	return p.instance, p.err
}

// Set defines a plugin collection, used with RegistryContext.
//
// This maintains ordering and unique indexing over the set.
type Set struct {
	ordered     []*ServiceInfo // order of initialization
	byTypeAndID map[Type]map[string]*ServiceInfo
}

// NewServiceInfoSet returns an initialized service info set
func NewServiceInfoSet() *Set {
	return &Set{
		byTypeAndID: make(map[Type]map[string]*ServiceInfo),
	}
}

// Add a service info to the set
func (ps *Set) Add(p *ServiceInfo) error {
	if byID, typeok := ps.byTypeAndID[p.Registration.Type]; !typeok {
		ps.byTypeAndID[p.Registration.Type] = map[string]*ServiceInfo{
			p.Registration.ID: p,
		}
	} else if _, idok := byID[p.Registration.ID]; !idok {
		byID[p.Registration.ID] = p
	} else {
		return fmt.Errorf("service %v already initialized", p.Registration.ID)
	}

	ps.ordered = append(ps.ordered, p)
	return nil
}

// Get returns the first service instance by its type
func (ps *Set) Get(t Type) (interface{}, error) {
	for _, v := range ps.byTypeAndID[t] {
		return v.Instance()
	}
	return nil, fmt.Errorf("no services registered for %s", t)
}

// GetAll returns all service infos by a type
func (ps *Set) GetAll(t Type) []*ServiceInfo {
	res := []*ServiceInfo{}

	for _, v := range ps.ordered {
		if v.Registration.Type == t {
			res = append(res, v)
		}
	}
	return res
}
