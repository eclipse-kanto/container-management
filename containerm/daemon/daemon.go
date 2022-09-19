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

package main

import (
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/registry"
)

type daemon struct {
	config         *config
	serviceInfoSet *registry.Set
}

func newDaemon(config *config) (*daemon, error) {
	log.Debug("starting Kanto CM daemon initialization")
	daemon := &daemon{
		config:         config,
		serviceInfoSet: registry.NewServiceInfoSet(),
	}
	log.Debug("successfully created Kanto CM daemon instance")
	return daemon, nil
}
