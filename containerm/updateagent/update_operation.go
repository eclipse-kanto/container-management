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
	"errors"
	"fmt"
	"os"
	"strings"

	ctrtypes "github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"

	"github.com/eclipse-kanto/update-manager/api/types"
)

type containerAction struct {
	desired *ctrtypes.Container
	current *ctrtypes.Container

	feedbackAction *types.Action
	actionType     util.ActionType
}

type baselineAction struct {
	baseline string
	status   types.StatusType
	actions  []*containerAction
}

type operation struct {
	ctx           context.Context
	updateManager *containersUpdateManager
	activityID    string
	desiredState  *internalDesiredState

	allActions      *baselineAction
	baselineActions map[string]*baselineAction
}

// UpdateOperation defines an interface for an update operation process
type UpdateOperation interface {
	GetActivityID() string
	Identify() (bool, error)
	Execute(command types.CommandType, baseline string)
	Feedback(status types.StatusType, message string, baseline string)
}

type createUpdateOperation func(*containersUpdateManager, string, *internalDesiredState) UpdateOperation

func newOperation(updMgr *containersUpdateManager, activityID string, desiredState *internalDesiredState) UpdateOperation {
	return &operation{
		updateManager: updMgr,
		activityID:    activityID,
		desiredState:  desiredState,
	}
}

// GetActivityID returns the activity ID associated with this operation
func (o *operation) GetActivityID() string {
	return o.activityID
}

// Identify executes the IDENTIFYING phase, triggered with the full desired state for the domain
func (o *operation) Identify() (bool, error) {
	if o.ctx == nil {
		o.ctx = context.Background()
	}
	currentContainers, err := o.updateManager.mgr.List(o.ctx)
	if err != nil {
		log.ErrorErr(err, "could not list all existing containers")
		return false, err
	}
	currentContainersMap := util.AsNamedMap(currentContainers)

	allActions := []*containerAction{}
	log.Debug("checking desired vs current containers")
	for _, desired := range o.desiredState.containers {
		id := desired.Name
		if o.isSystemContainer(id) {
			log.Warn("[%s] System container cannot be updated with desired state.", id)
			continue
		}
		current := currentContainersMap[id]
		if current != nil {
			delete(currentContainersMap, id)
		}
		allActions = append(allActions, o.newContainerAction(current, desired))
	}

	destroyActions := o.newDestroyActions(currentContainersMap)
	allActions = append(allActions, destroyActions...)

	// identify baseline actions, e.g. actions that are grouped together as a baseline
	baselineActions := make(map[string]*baselineAction)
	baselineRemoveContainers := o.updateManager.domainName + ":remove-components"
	baselineActions[baselineRemoveContainers] = &baselineAction{
		baseline: baselineRemoveContainers,
		status:   types.StatusIdentified,
		actions:  destroyActions,
	}
	for baseline, containers := range o.desiredState.baselines {
		baselineActions[baseline] = &baselineAction{
			baseline: baseline,
			status:   types.StatusIdentified,
			actions:  filterActions(allActions, containers),
		}
	}
	o.allActions = &baselineAction{
		baseline: "",
		status:   types.StatusIdentified,
		actions:  allActions,
	}
	o.baselineActions = baselineActions

	return len(allActions) > 0, nil
}

func (o *operation) newContainerAction(current *ctrtypes.Container, desired *ctrtypes.Container) *containerAction {
	actionType := util.DetermineUpdateAction(current, desired)
	message := util.GetActionMessage(actionType)

	log.Debug("[%s] %s", desired.Name, message)
	return &containerAction{
		desired: desired,
		current: current,
		feedbackAction: &types.Action{
			Component: &types.Component{
				ID:      o.updateManager.domainName + ":" + desired.Name,
				Version: o.desiredState.findComponent(desired.Name).Version,
			},
			Status:  types.ActionStatusIdentified,
			Message: message,
		},
		actionType: actionType,
	}
}

func (o *operation) newDestroyActions(toBeRemoved map[string]*ctrtypes.Container) []*containerAction {
	destroyActions := []*containerAction{}
	message := util.GetActionMessage(util.ActionDestroy)
	for id, current := range toBeRemoved {
		if o.isSystemContainer(id) {
			continue
		}
		log.Debug("[%s] %s", current.Name, message)
		destroyActions = append(destroyActions, &containerAction{
			desired: nil,
			current: current,
			feedbackAction: &types.Action{
				Component: &types.Component{
					ID:      o.updateManager.domainName + ":" + current.Name,
					Version: findContainerVersion(current.Image.Name),
				},
				Status:  types.ActionStatusIdentified,
				Message: message,
			},
			actionType: util.ActionDestroy,
		})
	}
	return destroyActions
}

func filterActions(actions []*containerAction, containers []*ctrtypes.Container) []*containerAction {
	result := []*containerAction{}
	for _, container := range containers {
		for _, action := range actions {
			if action.desired == container {
				result = append(result, action)
			}
		}
	}
	return result
}

// Execute executes each COMMAND (download, update, activate, etc) phase, triggered per baseline or for all the identified actions
func (o *operation) Execute(command types.CommandType, baseline string) {
	commandHandler, baselineAction := o.getBaselineCommandHandler(baseline, command)
	if baselineAction == nil {
		return
	}
	commandHandler(o, baselineAction)
}

type baselineCommandHandler func(*operation, *baselineAction)

var baselineCommandHandlers = map[types.CommandType]struct {
	expectedBaselineStatus []types.StatusType
	baselineFailureStatus  types.StatusType
	commandHandler         baselineCommandHandler
}{
	types.CommandDownload: {
		expectedBaselineStatus: []types.StatusType{types.StatusIdentified},
		baselineFailureStatus:  types.BaselineStatusDownloadFailure,
		commandHandler:         download,
	},
	types.CommandUpdate: {
		expectedBaselineStatus: []types.StatusType{types.BaselineStatusDownloadSuccess},
		baselineFailureStatus:  types.BaselineStatusUpdateFailure,
		commandHandler:         update,
	},
	types.CommandActivate: {
		expectedBaselineStatus: []types.StatusType{types.BaselineStatusUpdateSuccess},
		baselineFailureStatus:  types.BaselineStatusActivationFailure,
		commandHandler:         activate,
	},
	types.CommandRollback: {
		expectedBaselineStatus: []types.StatusType{types.BaselineStatusDownloadSuccess, types.BaselineStatusUpdateSuccess,
			types.BaselineStatusActivationFailure, types.BaselineStatusActivationSuccess},
		baselineFailureStatus: types.BaselineStatusRollbackFailure,
		commandHandler:        rollback,
	},
	types.CommandCleanup: {
		baselineFailureStatus: types.BaselineStatusCleanup,
		commandHandler:        cleanup,
	},
}

func (o *operation) getBaselineCommandHandler(baseline string, command types.CommandType) (baselineCommandHandler, *baselineAction) {
	handler, ok := baselineCommandHandlers[command]
	if !ok {
		log.Warn("Ignoring unknown command %", command)
		return nil, nil
	}
	var baselineAction *baselineAction
	if baseline == "*" || baseline == "" {
		o.allActions.baseline = baseline
		baselineAction = o.allActions
	} else {
		baselineAction = o.baselineActions[baseline]
	}
	if baselineAction == nil {
		o.Feedback(handler.baselineFailureStatus, "Unknown baseline "+baseline, baseline)
		return nil, nil
	}
	if len(handler.expectedBaselineStatus) > 0 && !hasStatus(handler.expectedBaselineStatus, baselineAction.status) {
		o.Feedback(handler.baselineFailureStatus, fmt.Sprintf("%s is possible only after status %s is reported", command, asStatusString(handler.expectedBaselineStatus)), baseline)
		return nil, nil
	}
	return handler.commandHandler, baselineAction
}

// ActionCreate and ActionRecreate: create new container instance, this will download the container image.
func download(o *operation, baselineAction *baselineAction) {
	var lastAction *containerAction
	var lastActionErr error
	lastActionMessage := ""

	log.Debug("downloading for baseline %s - starting...", baselineAction.baseline)
	defer func() {
		if lastActionErr == nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusDownloadSuccess, lastAction, types.ActionStatusDownloadSuccess, lastActionMessage)
		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusDownloadFailure, lastAction, types.ActionStatusDownloadFailure, lastActionErr.Error())
		}
		log.Debug("downloading for baseline %s - done", baselineAction.baseline)
	}()

	actions := baselineAction.actions
	for _, action := range actions {
		if action.actionType == util.ActionCreate || action.actionType == util.ActionRecreate {
			if lastAction != nil {
				lastAction.feedbackAction.Status = types.ActionStatusDownloadSuccess
				lastAction.feedbackAction.Message = lastActionMessage
			}
			lastAction = action
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusDownloading, action, types.ActionStatusDownloading, action.feedbackAction.Message)
			log.Debug("new container %s to be created...", action.feedbackAction.Component.ID)
			if err := o.createContainer(action.desired); err != nil {
				lastActionErr = err
				return
			}
			lastActionMessage = "New container created."
		}
	}
}

// ActionRecreate, ActionDestroy: stops the current container instance.
// ActionUpdate: update the running container configuration.
func update(o *operation, baselineAction *baselineAction) {
	var lastAction *containerAction
	var lastActionErr error
	lastActionMessage := ""

	log.Debug("updating for baseline %s - starting...", baselineAction.baseline)
	defer func() {
		if lastActionErr == nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusUpdateSuccess, lastAction, types.ActionStatusUpdateSuccess, lastActionMessage)
		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusUpdateFailure, lastAction, types.ActionStatusUpdateFailure, lastActionErr.Error())
		}
		log.Debug("updating for baseline %s - done.", baselineAction.baseline)
	}()

	actions := baselineAction.actions
	for _, action := range actions {
		if action.actionType != util.ActionRecreate && action.actionType != util.ActionDestroy && action.actionType != util.ActionUpdate {
			continue
		}
		if lastAction != nil {
			lastAction.feedbackAction.Status = types.ActionStatusUpdateSuccess
			lastAction.feedbackAction.Message = lastActionMessage
		}
		lastAction = action

		log.Debug("container %s to be updated...", action.feedbackAction.Component.ID)
		o.updateBaselineActionStatus(baselineAction, types.BaselineStatusUpdating, action, types.ActionStatusUpdating, action.feedbackAction.Message)
		if action.actionType == util.ActionRecreate || action.actionType == util.ActionDestroy {
			if err := o.stopContainer(action.current); err != nil {
				lastActionErr = err
				return
			}
			lastActionMessage = "Old container instance is stopped."
		} else { // action.actionType == util.ActionUpdate
			if err := o.updateContainer(action.current, action.desired); err != nil {
				lastActionErr = err
				return
			}
			lastActionMessage = "Container instance is updated with new configuration."
		}
	}
}

// ActionCreate, ActionRecreate: starts the newly created container instance (from DOWNLOAD phase).
// ActionUpdate, ActionCheck: ensure the existing container is running (call start/unpause container).
func activate(o *operation, baselineAction *baselineAction) {
	var lastAction *containerAction
	var lastActionErr error
	lastActionMessage := ""

	log.Debug("activating for baseline %s - starting...", baselineAction.baseline)
	defer func() {
		if lastActionErr == nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusActivationSuccess, lastAction, types.ActionStatusActivationSuccess, lastActionMessage)
		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusActivationFailure, lastAction, types.ActionStatusActivationFailure, lastActionErr.Error())
		}
		log.Debug("activating for baseline %s - done...", baselineAction.baseline)
	}()

	actions := baselineAction.actions
	for _, action := range actions {
		if action.actionType == util.ActionDestroy {
			continue
		}
		if lastAction != nil {
			lastAction.feedbackAction.Status = types.ActionStatusActivationSuccess
			lastAction.feedbackAction.Message = lastActionMessage
		}
		lastAction = action

		log.Debug("container %s to be activated...", action.feedbackAction.Component.ID)
		o.updateBaselineActionStatus(baselineAction, types.BaselineStatusActivating, action, types.ActionStatusActivating, action.feedbackAction.Message)
		if action.actionType == util.ActionCheck || action.actionType == util.ActionUpdate {
			if err := o.ensureRunningContainer(action.current); err != nil {
				lastActionErr = err
				return
			}
			if action.actionType == util.ActionCheck {
				lastActionMessage = "Existing container instance is running."
			} else {
				lastActionMessage = action.feedbackAction.Message
			}
		} else if action.actionType == util.ActionCreate || action.actionType == util.ActionRecreate {
			if err := o.startContainer(action.desired); err != nil {
				lastActionErr = err
				return
			}
			lastActionMessage = "New container instance is started."
		}
	}
}

// ActionCreate: removes the newly created container instance (from DOWNLOAD phase)
// ActionRecreate: removes the newly created container instance (from DOWNLOAD phase) and restarts the old existing container instance.
// ActionUpdate: restores the old configuration to the existing container and ensures it is started.
func rollback(o *operation, baselineAction *baselineAction) {
	var failure bool
	var lastAction *containerAction
	var lastActionMessage string

	log.Debug("rollback for baseline %s - starting...", baselineAction.baseline)
	defer func() {
		if !failure {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusRollbackSuccess, lastAction, types.ActionStatusUpdateFailure, lastActionMessage)
		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusRollbackFailure, lastAction, types.ActionStatusUpdateFailure, lastActionMessage)
		}
		log.Debug("rollback for baseline %s - done.", baselineAction.baseline)
	}()

	actions := baselineAction.actions
	for _, action := range actions {
		if lastAction != nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusRollback, lastAction, types.ActionStatusUpdateFailure, lastActionMessage)
		}
		log.Debug("container %s to be rolled back...", action.feedbackAction.Component.ID)
		lastAction = action
		if action.actionType == util.ActionUpdate {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusRollback, action, types.ActionStatusUpdating, action.feedbackAction.Message)
			if err := o.updateContainer(action.current, action.current); err != nil {
				o.configurationFileRollBack(action.current, &action.current.Mounts[len(action.current.Mounts)-1])
				lastActionMessage = err.Error()
				failure = true
				continue
			}
			if err := o.ensureRunningContainer(action.current); err != nil {
				lastActionMessage = err.Error()
				o.configurationFileRollBack(action.current, &action.current.Mounts[len(action.current.Mounts)-1])
				failure = true
				continue
			}
			lastActionMessage = "Update unsuccessful, but rollback succeeded - container configuration restored from older instance."
		} else if action.actionType == util.ActionCreate || action.actionType == util.ActionRecreate {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusRollback, action, types.ActionStatusUpdating, action.feedbackAction.Message)
			if err := o.removeContainer(action.desired); err != nil {
				lastActionMessage = err.Error()
				failure = true
				continue
			}
			if action.current != nil {
				if err := o.startContainer(action.current); err != nil {
					lastActionMessage = err.Error()
					failure = true
					continue
				}
				lastActionMessage = "Update unsuccessful, but rollback succeeded - new container instance destroyed, old container instance restored."
			} else {
				lastActionMessage = "Update unsuccessful, but rollback succeeded - new container instance destroyed."
			}
		} else {
			lastAction = nil
		}
	}
}

// ActionRecreate, ActionDestroy: removes the old existing container instance.
func cleanup(o *operation, baselineAction *baselineAction) {
	baseline := baselineAction.baseline
	actions := baselineAction.actions
	if baseline == "*" || baseline == "" {
		for b := range o.baselineActions {
			delete(o.baselineActions, b)
		}
	} else {
		delete(o.baselineActions, baseline)
	}

	log.Debug("cleanup for baseline %s (%s) - starting...", baseline, baselineAction.status)
	result := types.BaselineStatusCleanupSuccess
	if baselineAction.status != types.BaselineStatusActivationSuccess {
		log.Warn("cleanup implemented only for successfully activated baselines, no cleanup for baseline %s (%s)", baseline, baselineAction.status)
		// TODO implement cleanup for failure scenarios, maybe together with rollback
	} else {
		for _, action := range actions {
			if action.actionType == util.ActionRecreate || action.actionType == util.ActionDestroy {
				log.Debug("container %s to be cleanup...", action.feedbackAction.Component.ID)
				err := o.removeContainer(action.current)
				if action.feedbackAction.Status == types.ActionStatusUpdateSuccess && action.actionType == util.ActionDestroy {
					if err != nil {
						action.feedbackAction.Status = types.ActionStatusRemovalFailure
						action.feedbackAction.Message = err.Error()
						result = types.BaselineStatusCleanupFailure
					} else {
						action.feedbackAction.Status = types.ActionStatusRemovalSuccess
						action.feedbackAction.Message = "Old container instance is removed."
					}
				}
			}
		}
	}
	o.Feedback(result, "", baseline)
	log.Debug("cleanup for baseline (%s) %s - done...", baseline, baselineAction.status)

	if len(o.baselineActions) == 0 {
		o.updateManager.operation = nil
		if baselineAction.status == types.BaselineStatusActivationSuccess && result == types.BaselineStatusCleanupSuccess {
			o.Feedback(types.StatusCompleted, "", "")
		} else {
			o.Feedback(types.StatusIncomplete, "", "")
		}
	}
}

func (o *operation) isSystemContainer(containerID string) bool {
	systemContainers := o.desiredState.systemContainers
	if systemContainers == nil {
		systemContainers = o.updateManager.systemContainers
	}
	for _, systemContainerID := range systemContainers {
		if systemContainerID == containerID {
			return true
		}
	}
	return false
}

// Feedback sends desired state feedback responses, baseline parameter is optional
func (o *operation) Feedback(status types.StatusType, message string, baseline string) {
	o.updateManager.eventCallback.HandleDesiredStateFeedbackEvent(o.updateManager.domainName, o.activityID, baseline, status, message, o.toFeedbackActions())
}

func (o *operation) updateBaselineActionStatus(baseline *baselineAction, baselineStatus types.StatusType,
	action *containerAction, actionStatus types.ActionStatusType, message string) {
	if action != nil {
		action.feedbackAction.Status = actionStatus
		action.feedbackAction.Message = message
	}
	baseline.status = baselineStatus
	o.Feedback(baselineStatus, "", baseline.baseline)
}

func (o *operation) toFeedbackActions() []*types.Action {
	if o.allActions == nil {
		return nil
	}
	result := make([]*types.Action, len(o.allActions.actions))
	for i, action := range o.allActions.actions {
		result[i] = action.feedbackAction
	}
	return result
}

func hasStatus(where []types.StatusType, what types.StatusType) bool {
	for _, status := range where {
		if status == what {
			return true
		}
	}
	return false
}

func asStatusString(what []types.StatusType) string {
	var sb strings.Builder
	for _, status := range what {
		if sb.Len() > 0 {
			sb.WriteRune('|')
		}
		sb.WriteString(string(status))
	}
	return sb.String()
}

func (o *operation) createContainer(desired *ctrtypes.Container) error {
	log.Debug("container [%s] does not exist - will create a new one", desired.Name)
	_, err := o.updateManager.mgr.Create(o.ctx, desired)
	if err != nil {
		log.ErrorErr(err, "could not create container [%s]", desired.Name)
		o.configurationFileRollBack(desired, &desired.Mounts[len(desired.Mounts)-1])
		return err
	}
	log.Debug("successfully created container [%s]", desired.Name)
	return nil
}

func (o *operation) startContainer(container *ctrtypes.Container) error {
	if err := o.updateManager.mgr.Start(o.ctx, container.ID); err != nil {
		log.ErrorErr(err, "could not start container [%s]", container.Name)
		return err
	}
	log.Debug("successfully started container [%s]", container.Name)
	return nil
}

func (o *operation) unpauseContainer(container *ctrtypes.Container) error {
	if err := o.updateManager.mgr.Unpause(o.ctx, container.ID); err != nil {
		log.ErrorErr(err, "could not unpause container [%s]", container.Name)
		return err
	}
	log.Debug("successfully unpaused container [%s]", container.Name)
	return nil
}

func (o *operation) updateContainer(current *ctrtypes.Container, desired *ctrtypes.Container) error {
	log.Debug("there is an already existing container [%s] - will be updated with newer configuration", desired.Name)
	updateOpts := &ctrtypes.UpdateOpts{
		RestartPolicy: desired.HostConfig.RestartPolicy,
		Resources:     desired.HostConfig.Resources,
	}
	if err := o.updateManager.mgr.Update(o.ctx, current.ID, updateOpts); err != nil {
		log.ErrorErr(err, "could not update configuration for container [%s]", desired.Name)
		return err
	}
	log.Debug("successfully updated container [%s]", desired.Name)
	return nil
}

func (o *operation) ensureRunningContainer(current *ctrtypes.Container) error {
	container, err := o.updateManager.mgr.Get(o.ctx, current.ID)
	if err != nil {
		log.DebugErr(err, "cannot get current state for container [%s]", current.Name)
		return err
	}
	if container.State.Running {
		log.Debug("container [%s] is RUNNING - nothing to do more", current.Name)
		return nil
	}
	if container.State.Paused {
		log.Debug("container [%s] is PAUSED - will try to unpause it", current.Name)
		return o.unpauseContainer(container)
	}
	log.Debug("container [%s] is not RUNNING - will try to start it", current.Name)
	return o.startContainer(container)
}

func (o *operation) removeContainer(container *ctrtypes.Container) error {
	log.Debug("container [%s] is not desired - will be removed", container.Name)
	if err := o.updateManager.mgr.Remove(o.ctx, container.ID, true, nil); err != nil {
		log.ErrorErr(err, "could not remove undesired container [%s]", container.Name)
		return err
	}
	log.Debug("successfully removed container [%s]", container.Name)
	return nil
}

func (o *operation) stopContainer(container *ctrtypes.Container) error {
	if !util.IsContainerRunningOrPaused(container) {
		log.Debug("container [%s] is not RUNNING, nor PAUSED - nothing to do more", container.Name)
		return nil
	}
	stopOpts := &ctrtypes.StopOpts{
		Force:  true,
		Signal: "SIGTERM",
	}
	log.Debug("container [%s] will be updated - will stop current instance", container.Name)
	if err := o.updateManager.mgr.Stop(o.ctx, container.ID, stopOpts); err != nil {
		log.ErrorErr(err, "could not stop outdated container [%s]", container.Name)
		return err
	}
	log.Debug("successfully stopped outdated container [%s]", container.Name)
	return nil
}

func (o *operation) configurationFileRollBack(container *ctrtypes.Container, mount *ctrtypes.MountPoint) {
	if _, err := os.Stat(mount.Source + "/" + container.Name + "_config.json"); errors.Is(err, os.ErrNotExist) {
		return
	}
	err := util.MakeAtomicCopy(mount.Source+"/"+container.Name+"_config_backup.json", mount.Source+"/"+container.Name+"_config.json")
	if os.IsNotExist(err) {
		log.WarnErr(err, "earlier configuration does not exist, removing any configuration files")
		os.Remove(mount.Source + "/" + container.Name + "_config.json")
	} else {
		log.Debug("configuration file rolled back to previous configuration")
	}
}
