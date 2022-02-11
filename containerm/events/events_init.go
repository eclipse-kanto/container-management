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

package events

import (
	"github.com/eclipse-kanto/container-management/containerm/registry"
)

func newEventsManager() ContainerEventsManager {
	return &eventsMgr{
		broadcaster: newEventSinksDispatcher(),
	}
}

func registryInit(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
	return newEventsManager(), nil
}
