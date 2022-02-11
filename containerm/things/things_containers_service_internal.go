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
	"context"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/things/api/handlers"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
)

func (tMgr *containerThingsMgr) registryHandler(changedType handlers.ThingsRegistryChangedType, thing model.Thing) {
	tMgr.initMutex.Lock()
	defer tMgr.initMutex.Unlock()
	if changedType == handlers.Added {
		log.Debug("added new thing from Things service >>>>> %v", thing)
		tMgr.processThing(thing)
	}
}

func (tMgr *containerThingsMgr) thingsClientInitializedHandler(cl *client.Client, configuration *client.Configuration, err error) {
	tMgr.initMutex.Lock()
	defer tMgr.initMutex.Unlock()
	log.Debug("received things client initialized notification with client configuration: %s and Error info: %s", configuration, err)
	if err != nil {
		log.ErrorErr(err, "Error initializing things client")
		return
	}
	log.Debug("processing things client configuration")
	log.Debug("successfully initialized things manager info with {rootDeviceId:%s,rootDeviceTenantId:%s,rootDeviceAuthId:%s,rootDevicePassword:%s}", configuration.GatewayDeviceID(), configuration.DeviceTenantID(), configuration.DeviceAuthID(), configuration.DevicePassword())

	namespaceID := client.NewNamespacedIDFromString(client.NewNamespacedID(configuration.GatewayDeviceID(), configuration.DeviceName()).String())
	tMgr.containerThingID = namespaceID.String()
	rootThing := tMgr.thingsClient.Get(namespaceID)
	if rootThing == nil {
		log.Error("the root thing device with id = %s is missing in the things client's cache", tMgr.containerThingID)
	} else {
		// add features
		tMgr.processThing(rootThing)
	}
	cl.SetThingsRegistryChangedHandler(tMgr.registryHandler)
}

func (tMgr *containerThingsMgr) processThing(thing model.Thing) {
	if thing.GetID().String() == tMgr.containerThingID {
		ctx := context.Background()

		// dispose all features(their event handlers would be closed)
		tMgr.disposeFeatures()
		tMgr.managedFeatures = make(map[string]managedFeature)

		// handle ContainerFactory
		if tMgr.isFeatureEnabled(ContainerFactoryFeatureID) {
			log.Debug("registering %s feature", ContainerFactoryFeatureID)
			ctrFactory := newContainerFactoryFeature(tMgr.mgr, tMgr.eventsMgr, thing, newContainerFactoryStorage(tMgr.storageRoot, tMgr.containerThingID))
			tMgr.managedFeatures[ContainerFactoryFeatureID] = ctrFactory
		} else {
			log.Debug("ContainerFactory feature is NOT enabled and will not be registered. No Container feature per container instance will also be registered!")
		}
		// handle SoftwareUpdatable
		if tMgr.isFeatureEnabled(SoftwareUpdatableFeatureID) {
			log.Debug("registering %s feature", SoftwareUpdatableFeatureID)
			su := newSoftwareUpdatable(thing, tMgr.mgr, tMgr.eventsMgr)
			tMgr.managedFeatures[SoftwareUpdatableFeatureID] = su
		} else {
			log.Debug("SoftwareUpdatable feature is NOT enabled and will not be registered")
		}

		// register all added features
		for featureID, feature := range tMgr.managedFeatures {
			log.Debug("registering feature %s", featureID)
			if err := feature.register(ctx); err != nil {
				log.ErrorErr(err, "could not register %s feature", featureID)
			}
		}
	} else {
		log.Debug("the thing is not the containers thing - will not process it")
	}
}

func (tMgr *containerThingsMgr) disposeFeatures() {
	for featureID, feature := range tMgr.managedFeatures {
		log.Debug("disposing feature %s", featureID)
		feature.dispose()
	}
}

func (tMgr *containerThingsMgr) isFeatureEnabled(featureID string) bool {
	if tMgr.enabledFeatureIds == nil || len(tMgr.enabledFeatureIds) == 0 {
		return false
	}
	for _, enabled := range tMgr.enabledFeatureIds {
		if enabled == featureID {
			return true
		}
	}
	return false
}
