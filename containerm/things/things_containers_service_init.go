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

package things

import (
	"time"

	"github.com/eclipse-kanto/container-management/containerm/events"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/registry"
	"github.com/eclipse-kanto/container-management/things/client"
)

const (
	containersThingName = "edge:containers"
)

func newThingsContainerManager(mgr mgr.ContainerManager, eventsMgr events.ContainerEventsManager,
	brokerURL string,
	keepAlive time.Duration,
	disconnectTimeout time.Duration,
	username string,
	password string,
	storagePath string,
	enabledFeatureIds []string,
	connectTimeout time.Duration,
	acknowledgeTimeout time.Duration,
	subscribeTimeout time.Duration,
	unsubscribeTimeout time.Duration) *containerThingsMgr {
	thingsMgr := &containerThingsMgr{
		storageRoot:       storagePath,
		mgr:               mgr,
		eventsMgr:         eventsMgr,
		enabledFeatureIds: enabledFeatureIds,
		managedFeatures:   map[string]managedFeature{},
	}

	thingsClientOpts := client.NewConfiguration()
	thingsClientOpts.WithBroker(brokerURL).
		WithDisconnectTimeout(disconnectTimeout).
		WithKeepAlive(keepAlive).
		WithClientUsername(username).
		WithClientPassword(password).
		WithInitHook(thingsMgr.thingsClientInitializedHandler).
		WithDeviceName(containersThingName).
		WithConnectTimeout(connectTimeout).
		WithAcknowledgeTimeout(acknowledgeTimeout).
		WithSubscribeTimeout(subscribeTimeout).
		WithUnsubscribeTimeout(unsubscribeTimeout)

	thingsMgr.thingsClient = client.NewClient(thingsClientOpts)
	return thingsMgr

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
	initOpts := registryCtx.Config.([]ContainerThingsManagerOpt)
	tOpts := &thingsOpts{}
	err = applyOptsThings(tOpts, initOpts...)
	if err != nil {
		return nil, err
	}
	return newThingsContainerManager(mgrService.(mgr.ContainerManager), eventsMgr.(events.ContainerEventsManager),
		tOpts.broker,
		tOpts.keepAlive,
		tOpts.disconnectTimeout,
		tOpts.clientUsername,
		tOpts.clientPassword,
		tOpts.storagePath,
		tOpts.featureIds,
		tOpts.connectTimeout,
		tOpts.acknowledgeTimeout,
		tOpts.subscribeTimeout,
		tOpts.unsubscribeTimeout), nil
}
