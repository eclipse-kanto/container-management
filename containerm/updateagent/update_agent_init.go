// Copyright (c) 2023 Contributors to the Eclipse Foundation
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

package updateagent

import (
	"time"

	"github.com/eclipse-kanto/container-management/containerm/events"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/registry"

	"github.com/eclipse-kanto/update-manager/api"
	"github.com/eclipse-kanto/update-manager/api/agent"
	"github.com/eclipse-kanto/update-manager/mqtt"
)

func newUpdateAgent(mgr mgr.ContainerManager, eventsMgr events.ContainerEventsManager,
	domainName string,
	systemContainers []string,
	verboseInventory bool,
	brokerURL string,
	keepAlive time.Duration,
	disconnectTimeout time.Duration,
	clientUsername string,
	clientPassword string,
	connectTimeout time.Duration,
	acknowledgeTimeout time.Duration,
	subscribeTimeout time.Duration,
	unsubscribeTimeout time.Duration,
	tlsConfig *tlsConfig) (api.UpdateAgent, error) {

	mqttClient := mqtt.NewUpdateAgentClient(domainName, &mqtt.ConnectionConfig{
		BrokerURL:          brokerURL,
		KeepAlive:          keepAlive.Milliseconds(),
		DisconnectTimeout:  disconnectTimeout.Milliseconds(),
		ClientUsername:     clientUsername,
		ClientPassword:     clientPassword,
		ConnectTimeout:     connectTimeout.Milliseconds(),
		AcknowledgeTimeout: acknowledgeTimeout.Milliseconds(),
		SubscribeTimeout:   subscribeTimeout.Milliseconds(),
		UnsubscribeTimeout: unsubscribeTimeout.Milliseconds(),
	})

	return agent.NewUpdateAgent(mqttClient, newUpdateManager(mgr, eventsMgr, domainName, systemContainers, verboseInventory)), nil
}

// newUpdateManager instantiates a new update manager instance
func newUpdateManager(mgr mgr.ContainerManager, eventsMgr events.ContainerEventsManager,
	domainName string, systemContainers []string, verboseContainers bool) api.UpdateManager {
	return &containersUpdateManager{
		domainName:        domainName,
		systemContainers:  systemContainers,
		verboseContainers: verboseContainers,

		mgr:                   mgr,
		eventsMgr:             eventsMgr,
		createUpdateOperation: newOperation,
	}
}

func registryInit(registryCtx *registry.ServiceRegistryContext) (interface{}, error) {
	eventsMgr, err := registryCtx.Get(registry.EventsManagerService)
	if err != nil {
		return nil, err
	}
	mgrService, err := registryCtx.Get(registry.ContainerManagerService)
	if err != nil {
		return nil, err
	}

	// init options processing
	uaOpts := &updateAgentOpts{}
	if err := applyOptsUpdateAgent(uaOpts, registryCtx.Config.([]ContainersUpdateAgentOpt)...); err != nil {
		return nil, err
	}
	return newUpdateAgent(mgrService.(mgr.ContainerManager), eventsMgr.(events.ContainerEventsManager),
		uaOpts.domainName,
		uaOpts.systemContainers,
		uaOpts.verboseInventory,
		uaOpts.broker,
		uaOpts.keepAlive,
		uaOpts.disconnectTimeout,
		uaOpts.clientUsername,
		uaOpts.clientPassword,
		uaOpts.connectTimeout,
		uaOpts.acknowledgeTimeout,
		uaOpts.subscribeTimeout,
		uaOpts.unsubscribeTimeout,
		uaOpts.tlsConfig,
	)
}
