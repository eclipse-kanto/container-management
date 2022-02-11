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
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

func (ctrFactory *containerFactoryFeature) createContainerFeature(ctr *types.Container) error {
	ctrFeature := newContainerFeature(ctr.Image.Name, ctr.Name, ctr, ctrFactory.mgr)
	dittoFeature := ctrFeature.createFeature()
	if err := ctrFactory.rootThing.SetFeature(dittoFeature.GetID(), dittoFeature); err != nil {
		return err
	}
	return nil
}

func (ctrFactory *containerFactoryFeature) removeContainerFeature(ctrID string) error {
	return ctrFactory.rootThing.RemoveFeature(generateContainerFeatureID(ctrID))
}

func (ctrFactory *containerFactoryFeature) updateContainerFeature(id, path string, value interface{}) error {
	return ctrFactory.rootThing.SetFeatureProperty(generateContainerFeatureID(id), path, value)
}

func (ctrFactory *containerFactoryFeature) processContainers(ctrs []*types.Container) {
	storedCtrFeaturesInfo, err := ctrFactory.storageMgr.Restore()
	if err != nil {
		log.WarnErr(err, "could not restore things service persistent info")
	}

	var currentContainers map[string]string
	if ctrs != nil && len(ctrs) > 0 {
		currentContainers = map[string]string{}
		for _, ctr := range ctrs {
			delete(storedCtrFeaturesInfo, ctr.ID)
			if err := ctrFactory.createContainerFeature(ctr); err != nil {
				log.ErrorErr(err, "could not create container feature for container ID = %s", ctr.ID)
			} else {
				currentContainers[ctr.ID] = generateContainerFeatureID(ctr.ID)
			}
		}
	} else {
		log.Debug("no containers available to create/update features for")
	}

	if storedCtrFeaturesInfo != nil && len(storedCtrFeaturesInfo) > 0 {
		for ctrID := range storedCtrFeaturesInfo {
			log.Debug("removing feature from no longer existing container ID = %s", ctrID)
			if err := ctrFactory.removeContainerFeature(ctrID); err != nil {
				log.ErrorErr(err, "could not remove stale container feature for container ID = %s", ctrID)
			}
		}
	}

	log.Debug("storing new container IDs info after sanity check %v", currentContainers)
	ctrFactory.storageMgr.UpdateContainersInfo(currentContainers)
}
