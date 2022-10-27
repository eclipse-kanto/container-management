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

package things

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/events"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
)

const (
	// ContainerFactoryFeatureID is the feature ID of the container factory
	ContainerFactoryFeatureID = "ContainerFactory"
	// containerFactoryFeatureDefinition is the feature definition of the container factory
	containerFactoryFeatureDefinition = "com.bosch.iot.suite.edge.containers:ContainerFactory:1.2.0"
	// containerFactoryFeatureOperationCreate is the name of the operation that the feature implements
	// based on the Vorto model provided in the feature's definition of the create operation
	containerFactoryFeatureOperationCreate = "create"
	// containerFactoryFeatureOperationCreateWithConfig is the name of the operation that the feature implements
	// based on the Vorto model provided in the feature's definition of the createWithConfig operation
	containerFactoryFeatureOperationCreateWithConfig = "createWithConfig"
)

type containerFactoryFeature struct {
	mgr                 mgr.ContainerManager
	cancelEventsHandler context.CancelFunc
	storageMgr          containerStorage
	eventsMgr           events.ContainerEventsManager
	rootThing           model.Thing
	eventsHandlingLock  sync.Mutex
}
type createArgs struct {
	ImageRef string `json:"imageRef"`
	Start    bool   `json:"start"`
}

type createWithConfigArgs struct {
	ImageRef string         `json:"imageRef"`
	Name     string         `json:"name,omitempty"`
	Config   *configuration `json:"config,omitempty"`
	Start    bool           `json:"start"`
}

func newContainerFactoryFeature(mgr mgr.ContainerManager, eventsMgr events.ContainerEventsManager, rootThing model.Thing, storageMgr containerStorage) managedFeature {
	return &containerFactoryFeature{
		mgr:        mgr,
		storageMgr: storageMgr,
		eventsMgr:  eventsMgr,
		rootThing:  rootThing,
	}
}

func (ctrFactory *containerFactoryFeature) register(ctx context.Context) error {
	log.Debug("initializing ContainerFactory feature")

	log.Debug("initializing container features")
	ctrs, err := ctrFactory.mgr.List(ctx)
	if err != nil {
		log.ErrorErr(err, "could not list containers for features sanity check")
	} else {
		ctrFactory.processContainers(ctrs)
	}
	if ctrFactory.cancelEventsHandler == nil {
		ctrFactory.handleContainerEvents(ctx)
		log.Debug("subscribed for container events")
	}
	return ctrFactory.rootThing.SetFeature(ContainerFactoryFeatureID, ctrFactory.createFeature())
}

func (ctrFactory *containerFactoryFeature) dispose() {
	log.Debug("disposing ContainerFactory feature")
	if ctrFactory.cancelEventsHandler != nil {
		log.Debug("unsubscribing from container events")
		ctrFactory.cancelEventsHandler()
		ctrFactory.cancelEventsHandler = nil
	}
}

func (ctrFactory *containerFactoryFeature) createFeature() model.Feature {
	return client.NewFeature(ContainerFactoryFeatureID,
		client.WithFeatureDefinitionFromString(containerFactoryFeatureDefinition),
		client.WithFeatureOperationsHandler(ctrFactory.featureOperationsHandler))
}

func (ctrFactory *containerFactoryFeature) featureOperationsHandler(operationName string, args interface{}) (interface{}, error) {
	ctx := context.Background()
	switch operationName {
	case containerFactoryFeatureOperationCreate:
		bytes, err := json.Marshal(args)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		cArgs := &createArgs{}
		err = json.Unmarshal(bytes, cArgs)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		return nil, ctrFactory.create(ctx, cArgs.ImageRef, cArgs.Start)
	case containerFactoryFeatureOperationCreateWithConfig:
		bytes, err := json.Marshal(args)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		cArgs := &createWithConfigArgs{}
		err = json.Unmarshal(bytes, cArgs)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		return nil, ctrFactory.createWithConfig(ctx, cArgs.ImageRef, cArgs.Name, cArgs.Config, cArgs.Start)
	default:
		err := log.NewErrorf("unsupported operation %s", operationName)
		log.ErrorErr(err, "unsupported operation %s", operationName)
		return nil, client.NewMessagesSubjectNotFound(err.Error())
	}
}

func (ctrFactory *containerFactoryFeature) create(ctx context.Context, imageRef string, start bool) error {
	if imageRef == "" {
		return log.NewError("imageRef must be set")
	}
	ctr := &types.Container{
		Image: types.Image{
			Name: imageRef,
		},
	}
	var (
		resCtr *types.Container
		err    error
	)
	if resCtr, err = ctrFactory.mgr.Create(ctx, ctr); err != nil {
		log.ErrorErr(err, "failed to create container")
		return err
	}
	if start {
		if err := ctrFactory.mgr.Start(ctx, resCtr.ID); err != nil {
			log.ErrorErr(err, "could not auto start container ID = %s", ctr.ID)
		}
	}
	return nil
}

func (ctrFactory *containerFactoryFeature) createWithConfig(ctx context.Context, imageRef, name string, cfg *configuration, start bool) error {
	if imageRef == "" {
		return log.NewError("imageRef must be set")
	}
	var ctr *types.Container
	if cfg != nil {
		ctr = toAPIContainerConfig(cfg)
	} else {
		ctr = &types.Container{}
	}
	ctr.Name = name
	ctr.Image = types.Image{Name: imageRef, DecryptConfig: ctr.Image.DecryptConfig}

	var (
		resCtr *types.Container
		err    error
	)
	if resCtr, err = ctrFactory.mgr.Create(ctx, ctr); err != nil {
		log.ErrorErr(err, "failed to create container")
		return err
	}
	if start {
		if err := ctrFactory.mgr.Start(ctx, resCtr.ID); err != nil {
			log.ErrorErr(err, "could not auto start container ID = %s", ctr.ID)
		}
	}
	return nil
}
