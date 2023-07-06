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
	"context"
	"sync"

	"github.com/eclipse-kanto/container-management/containerm/events"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/version"

	"github.com/eclipse-kanto/update-manager/api"
	"github.com/eclipse-kanto/update-manager/api/types"
)

const (
	updateManagerName = "Eclipse Kanto Containers Update Agent"
	parameterDomain   = "domain"
)

type containersUpdateManager struct {
	domainName             string
	systemContainers       []string
	verboseInventoryReport bool

	mgr       mgr.ContainerManager
	eventsMgr events.ContainerEventsManager

	applyLock             sync.Mutex
	eventCallback         api.UpdateManagerCallback
	createUpdateOperation createUpdateOperation
	operation             UpdateOperation
}

// Name returns the name of this update manager, e.g. "containers".
func (updMgr *containersUpdateManager) Name() string {
	return updMgr.domainName
}

// Apply triggers the update operation with the given activity ID and desired state with containers.
// First, it validates the received desired state specification and identifies the actions to be applied.
// If errors are detected, then IDENTIFICATION_FAILED feedback status is reported and operation finishes unsuccessfully.
// Otherwise, IDENTIFIED feedback status with identified actions is reported and it will wait for further commands to proceed.
func (updMgr *containersUpdateManager) Apply(ctx context.Context, activityID string, desiredState *types.DesiredState) {
	updMgr.applyLock.Lock()
	defer updMgr.applyLock.Unlock()

	log.Debug("processing desired state - start")
	// create operation instance
	updMgr.operation = updMgr.createUpdateOperation(updMgr, activityID, desiredState)

	// identification phase
	updMgr.operation.Feedback(types.StatusIdentifying, "", "")
	if err := updMgr.operation.Identify(); err != nil {
		updMgr.operation.Feedback(types.StatusIdentificationFailed, err.Error(), "")
		return
	}
	updMgr.operation.Feedback(types.StatusIdentified, "", "")

	log.Debug("processing desired state - identification phase completed, waiting for commands...")
}

// Command processes received desired state command.
func (updMgr *containersUpdateManager) Command(ctx context.Context, activityID string, command *types.DesiredStateCommand) {
	if command == nil {
		log.Error("Skipping received command for activityId %s, but no payload.", activityID)
		return
	}
	updMgr.applyLock.Lock()
	defer updMgr.applyLock.Unlock()

	operation := updMgr.operation
	if operation == nil {
		log.Warn("Ignoring received command %s for baseline %s and activityId %s, but no operation in progress.", command.Command, command.Baseline, activityID)
		return
	}
	if operation.GetActivityID() != activityID {
		log.Warn("Ignoring received command %s for baseline %s and activityId %s, but not matching operation in progress [%s].",
			command.Command, command.Baseline, activityID, operation.GetActivityID())
		return
	}
	operation.Execute(command.Command, command.Baseline)
}

// Get returns the current state as an inventory graph.
// The inventory graph includes a root software node (type APPLICATION) representing the update agent itself and a list of software nodes (type CONTAINER) representing the available containers.
func (updMgr *containersUpdateManager) Get(ctx context.Context, activityID string) (*types.Inventory, error) {
	return toInventory(updMgr.asSoftwareNode(), updMgr.getCurrentContainers()), nil
}

func toInventory(swNodeAgent *types.SoftwareNode, swNodeContainers []*types.SoftwareNode) *types.Inventory {
	swNodes := []*types.SoftwareNode{swNodeAgent}
	associations := []*types.Association{}
	if len(swNodeContainers) > 0 {
		swNodes = append(swNodes, swNodeContainers...)

		for _, swNodeContainer := range swNodeContainers {
			swNodeContainer.ID = swNodeAgent.Parameters[0].Value + ":" + swNodeContainer.ID
			associations = append(associations, &types.Association{
				SourceID: swNodeAgent.ID,
				TargetID: swNodeContainer.ID,
			})
		}
	}
	return &types.Inventory{
		SoftwareNodes: swNodes,
		Associations:  associations,
	}
}

func (updMgr *containersUpdateManager) asSoftwareNode() *types.SoftwareNode {
	return &types.SoftwareNode{
		InventoryNode: types.InventoryNode{
			ID:      updMgr.Name() + "-update-agent",
			Version: version.ProjectVersion,
			Name:    updateManagerName,
			Parameters: []*types.KeyValuePair{
				{
					Key:   parameterDomain,
					Value: updMgr.Name(),
				},
			},
		},
		Type: types.SoftwareTypeApplication,
	}
}

func (updMgr *containersUpdateManager) getCurrentContainers() []*types.SoftwareNode {
	_, err := updMgr.mgr.List(context.Background())
	if err != nil {
		log.ErrorErr(err, "could not list all existing containers")
		return nil
	}
	// TODO implement function fromContainers(containers, updMgr.verboseInventoryReport)
	return nil
}

// Dispose releases all resources used by this instance
func (updMgr *containersUpdateManager) Dispose() error {
	return nil
}

// WatchEvents subscribes for events that update the current state inventory
func (updMgr *containersUpdateManager) WatchEvents(ctx context.Context) {
	// no container events handled yet - current state inventory reported only on initial start or explicit get request
}

// SetCallback sets the callback instance that is used for desired state feedback / current state notifications.
// It is set when the update agent instance is started
func (updMgr *containersUpdateManager) SetCallback(callback api.UpdateManagerCallback) {
	updMgr.eventCallback = callback
}
