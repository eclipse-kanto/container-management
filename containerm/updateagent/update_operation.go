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
	"strings"

	"github.com/eclipse-kanto/container-management/containerm/log"

	"github.com/eclipse-kanto/update-manager/api/types"
)

type containerAction struct {
	// TODO add current / desired container + actionType
	feedbackAction *types.Action
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
	desiredState  *types.DesiredState

	allActions      *baselineAction
	baselineActions map[string]*baselineAction
}

// UpdateOperation defines an interface for an update operation process
type UpdateOperation interface {
	GetActivityID() string
	Identify() error
	Execute(command types.CommandType, baseline string)
	Feedback(status types.StatusType, message string, baseline string)
}

type createUpdateOperation func(*containersUpdateManager, string, *types.DesiredState) UpdateOperation

func newOperation(updMgr *containersUpdateManager, activityID string, desiredState *types.DesiredState) UpdateOperation {
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
func (o *operation) Identify() error {
	if o.ctx == nil {
		o.ctx = context.Background()
	}
	// TODO compare current vs. desired containers and identify actions
	return errors.New("Not implemented yet")
}

// Execute executes each COMMAND (download, update, activate, etc) phase, triggered per baseline or for all the identified actions
func (o *operation) Execute(command types.CommandType, baseline string) {
	switch command {
	case types.CommandDownload:
		o.download(baseline)
	case types.CommandUpdate:
		o.update(baseline)
	case types.CommandActivate:
		o.activate(baseline)
	case types.CommandRollback:
		o.rollback(baseline)
	case types.CommandCleanup:
		o.cleanup(baseline)
		if len(o.baselineActions) == 0 {
			o.updateManager.operation = nil
			o.Feedback(types.StatusCompleted, "", "")
		}
	default:
		log.Warn("Ignoring unknown command %", command)
	}
}

var baselineValidation = map[types.CommandType]struct {
	expectedBaselineStatus []types.StatusType
	baselineFailureStatus  types.StatusType
}{
	types.CommandDownload: {
		expectedBaselineStatus: []types.StatusType{types.StatusIdentified},
		baselineFailureStatus:  types.BaselineStatusDownloadFailure,
	},
	types.CommandUpdate: {
		expectedBaselineStatus: []types.StatusType{types.BaselineStatusDownloadSuccess},
		baselineFailureStatus:  types.BaselineStatusUpdateFailure,
	},
	types.CommandActivate: {
		expectedBaselineStatus: []types.StatusType{types.BaselineStatusUpdateSuccess},
		baselineFailureStatus:  types.BaselineStatusActivationFailure,
	},
	types.CommandRollback: {
		expectedBaselineStatus: []types.StatusType{types.BaselineStatusActivationFailure, types.BaselineStatusActivationSuccess},
		baselineFailureStatus:  types.BaselineStatusRollbackFailure,
	},
	types.CommandCleanup: {
		baselineFailureStatus: types.BaselineStatusCleanup,
	},
}

func (o *operation) getBaselineActionForCommand(baseline string, command types.CommandType) *baselineAction {
	var baselineAction *baselineAction
	if baseline == "*" || baseline == "" {
		o.allActions.baseline = baseline
		baselineAction = o.allActions
	} else {
		baselineAction = o.baselineActions[baseline]
	}
	validation := baselineValidation[command]
	if baselineAction == nil {
		o.Feedback(validation.baselineFailureStatus, "Unknown baseline "+baseline, baseline)
		return nil
	}
	if len(validation.expectedBaselineStatus) > 0 && !hasStatus(validation.expectedBaselineStatus, baselineAction.status) {
		o.Feedback(validation.baselineFailureStatus, fmt.Sprintf("%s is possible only after status %s is reported", command, asStatusString(validation.expectedBaselineStatus)), baseline)
		return nil
	}
	return baselineAction
}

// ActionCreate and ActionRecreate: create new container instance, this will download the container image.
func (o *operation) download(baseline string) {
	baselineAction := o.getBaselineActionForCommand(baseline, types.CommandDownload)
	if baselineAction == nil {
		return
	}

	var lastAction *containerAction
	var lastActionErr error
	lastActionMessage := ""

	log.Debug("downloading for baseline %s - starting...", baseline)
	defer func() {
		if lastActionErr == nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusDownloadSuccess, lastAction, types.ActionStatusDownloadSuccess, lastActionMessage)
		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusDownloadFailure, lastAction, types.ActionStatusDownloadFailure, lastActionErr.Error())
		}
		log.Debug("downloading for baseline %s - done", baseline)
	}()

	// TODO implement download
}

// ActionRecreate, ActionDestroy: stops the current container instance.
// ActionUpdate: update the running container configuration.
func (o *operation) update(baseline string) {
	baselineAction := o.getBaselineActionForCommand(baseline, types.CommandUpdate)
	if baselineAction == nil {
		return
	}

	var lastAction *containerAction
	var lastActionErr error
	lastActionMessage := ""

	log.Debug("updating for baseline %s - starting...", baseline)
	defer func() {
		if lastActionErr == nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusUpdateSuccess, lastAction, types.ActionStatusUpdateSuccess, lastActionMessage)
		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusUpdateFailure, lastAction, types.ActionStatusUpdateFailure, lastActionErr.Error())
		}
		log.Debug("updating for baseline %s - done.", baseline)
	}()

	// TODO implement update
}

// ActionCreate, ActionRecreate: starts the newly created container instance (from DOWNLOAD phase).
// ActionUpdate, ActionCheck: ensure the existing container is running (call start/unpause container).
func (o *operation) activate(baseline string) {
	baselineAction := o.getBaselineActionForCommand(baseline, types.CommandActivate)
	if baselineAction == nil {
		return
	}

	var lastAction *containerAction
	var lastActionErr error
	lastActionMessage := ""

	log.Debug("activating for baseline %s - starting...", baseline)
	defer func() {
		if lastActionErr == nil {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusActivationSuccess, lastAction, types.ActionStatusActivationSuccess, lastActionMessage)
		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusActivationFailure, lastAction, types.ActionStatusActivationFailure, lastActionErr.Error())
		}
		log.Debug("activating for baseline %s - done...", baseline)
	}()

	// TODO implement activate
}

// ActionCreate: removes the newly created container instance (from DOWNLOAD phase)
// ActionRecreate: removes the newly created container instance (from DOWNLOAD phase) and restarts the old existing container instance.
// ActionUpdate: restores the old configuration to the existing container and ensures it is started.
func (o *operation) rollback(baseline string) {
	baselineAction := o.getBaselineActionForCommand(baseline, types.CommandRollback)
	if baselineAction == nil {
		return
	}

	var failure bool
	var lastAction *containerAction
	var lastActionMessage string

	log.Debug("rollback for baseline %s - starting...", baseline)
	defer func() {
		if !failure {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusRollbackSuccess, lastAction, types.ActionStatusUpdateFailure, lastActionMessage)
		} else {
			o.updateBaselineActionStatus(baselineAction, types.BaselineStatusRollbackFailure, lastAction, types.ActionStatusUpdateFailure, lastActionMessage)
		}
		log.Debug("rollback for baseline %s - done.", baseline)
	}()

	// TODO implement rollback
}

// ActionRecreate, ActionDestroy: removes the old existing container instance.
func (o *operation) cleanup(baseline string) {
	baselineAction := o.getBaselineActionForCommand(baseline, types.CommandUpdate)
	if baselineAction == nil {
		return
	}

	if baseline == "*" || baseline == "" {
		for b := range o.baselineActions {
			delete(o.baselineActions, b)
		}
	} else {
		delete(o.baselineActions, baseline)
	}
	log.Debug("cleanup for baseline %s - starting...", baseline)

	// TODO implement cleanup

	o.Feedback(types.BaselineStatusCleanupSuccess, "", baseline)
	log.Debug("cleanup for baseline %s - done...", baseline)
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
