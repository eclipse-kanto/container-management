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
	"fmt"
	"strings"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
)

const (
	containerFeatureIDPrefix                 = "Container:"
	containerFeatureIDTemplate               = containerFeatureIDPrefix + "%s"
	containerFeatureDefinition               = "com.bosch.iot.suite.edge.containers:Container:1.5.0"
	containerFeaturePropertyStatus           = "status"
	containerFeaturePropertyPathStatusState  = containerFeaturePropertyStatus + "/state"
	containerFeaturePropertyPathStatusName   = containerFeaturePropertyStatus + "/name"
	containerFeaturePropertyPathStatusConfig = containerFeaturePropertyStatus + "/config"
	containerFeatureOperationStart           = "start"
	containerFeatureOperationStop            = "stop"
	containerFeatureOperationStopWithOptions = "stopWithOptions"
	containerFeatureOperationPause           = "pause"
	containerFeatureOperationResume          = "resume"
	containerFeatureOperationRemove          = "remove"
	containerFeatureOperationRename          = "rename"
	containerFeatureOperationUpdate          = "update"
)

type containerFeatureStatus struct {
	Name      string         `json:"name,omitempty"`
	ImageRef  string         `json:"imageRef"`
	Config    *configuration `json:"config,omitempty"`
	CreatedAt string         `json:"createdAt"`
	State     *state         `json:"state"`
}

type containerFeature struct {
	id     string
	status *containerFeatureStatus
	mgr    mgr.ContainerManager
}

func newContainerFeature(imageRef, name string, ctr *types.Container, mgr mgr.ContainerManager) *containerFeature {
	return &containerFeature{
		id: generateContainerFeatureID(ctr.ID),
		status: &containerFeatureStatus{
			Name:      name,
			ImageRef:  imageRef,
			Config:    fromAPIContainerConfig(ctr),
			State:     fromAPIContainerState(ctr.State),
			CreatedAt: ctr.Created,
		},
		mgr: mgr,
	}
}

func (ctrFeature *containerFeature) featureOperationsHandler(operationName string, args interface{}) (interface{}, error) {
	ctx := context.Background()
	switch operationName {
	case containerFeatureOperationStart:
		return nil, ctrFeature.start(ctx)
	case containerFeatureOperationPause:
		return nil, ctrFeature.pause(ctx)
	case containerFeatureOperationResume:
		return nil, ctrFeature.resume(ctx)
	case containerFeatureOperationStop:
		return nil, ctrFeature.stop(ctx)
	case containerFeatureOperationRename:
		bytes, err := json.Marshal(args)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		rArgs := ""
		err = json.Unmarshal(bytes, &rArgs)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		return nil, ctrFeature.rename(ctx, rArgs)
	case containerFeatureOperationStopWithOptions:
		bytes, err := json.Marshal(args)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		opts := &stopOptions{}
		err = json.Unmarshal(bytes, opts)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		return nil, ctrFeature.stopWithOptions(ctx, opts)
	case containerFeatureOperationRemove:
		bytes, err := json.Marshal(args)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		rArgs := false
		err = json.Unmarshal(bytes, &rArgs)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		return nil, ctrFeature.remove(ctx, rArgs)
	case containerFeatureOperationUpdate:
		bytes, err := json.Marshal(args)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		opts := &updateOptions{}
		if err = json.Unmarshal(bytes, opts); err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		return nil, ctrFeature.update(ctx, opts)
	default:
		err := log.NewErrorf("unsupported operation %s", operationName)
		log.ErrorErr(err, "unsupported operation %s", operationName)
		return nil, client.NewMessagesSubjectNotFound(err.Error())
	}
}

func (ctrFeature *containerFeature) start(ctx context.Context) error {
	return ctrFeature.mgr.Start(ctx, extractContainerID(ctrFeature.id))
}
func (ctrFeature *containerFeature) pause(ctx context.Context) error {
	return ctrFeature.mgr.Pause(ctx, extractContainerID(ctrFeature.id))
}
func (ctrFeature *containerFeature) resume(ctx context.Context) error {
	return ctrFeature.mgr.Unpause(ctx, extractContainerID(ctrFeature.id))
}
func (ctrFeature *containerFeature) stop(ctx context.Context) error {
	return ctrFeature.mgr.Stop(ctx, extractContainerID(ctrFeature.id), nil)
}
func (ctrFeature *containerFeature) stopWithOptions(ctx context.Context, opts *stopOptions) error {
	return ctrFeature.mgr.Stop(ctx, extractContainerID(ctrFeature.id), &types.StopOpts{
		Timeout: opts.Timeout,
		Force:   opts.Force,
		Signal:  opts.Signal,
	})
}
func (ctrFeature *containerFeature) remove(ctx context.Context, force bool) error {
	return ctrFeature.mgr.Remove(ctx, extractContainerID(ctrFeature.id), force, false, nil)
}

func (ctrFeature *containerFeature) rename(ctx context.Context, name string) error {
	return ctrFeature.mgr.Rename(ctx, extractContainerID(ctrFeature.id), name)
}

func (ctrFeature *containerFeature) update(ctx context.Context, opts *updateOptions) error {
	return ctrFeature.mgr.Update(ctx, extractContainerID(ctrFeature.id), toAPIUpdateOptions(opts))
}

func (ctrFeature *containerFeature) createFeature() model.Feature {
	return client.NewFeature(ctrFeature.id,
		client.WithFeatureDefinitionFromString(containerFeatureDefinition),
		client.WithFeatureProperty(containerFeaturePropertyStatus, ctrFeature.status),
		client.WithFeatureOperationsHandler(ctrFeature.featureOperationsHandler),
	)
}

func generateContainerFeatureID(containerID string) string {
	return fmt.Sprintf(containerFeatureIDTemplate, containerID)
}

func extractContainerID(containerFeatureID string) string {
	return strings.Replace(containerFeatureID, containerFeatureIDPrefix, "", 1)
}
