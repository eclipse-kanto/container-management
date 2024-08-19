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

	"github.com/eclipse-kanto/container-management/containerm/events"
	"github.com/eclipse-kanto/container-management/containerm/mgr"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	errorUtil "github.com/eclipse-kanto/container-management/containerm/util/error"
	"github.com/eclipse-kanto/container-management/rollouts/api/datatypes"
	"github.com/eclipse-kanto/container-management/rollouts/api/features"
	"github.com/eclipse-kanto/container-management/things/api/model"
	"github.com/eclipse-kanto/container-management/things/client"
)

// SoftwareUpdatable feature information
const (
	SoftwareUpdatableFeatureID           = "SoftwareUpdatable"
	SoftwareUpdatableDefinitionNamespace = "org.eclipse.hawkbit.swupdatable"
	SoftwareUpdatableDefinitionName      = "SoftwareUpdatable"
	SoftwareUpdatableDefinitionVersion   = "2.0.0"
	softwareUpdatablePropertyNameStatus  = "status"

	softwareUpdatablePropertySoftwareModuleType    = softwareUpdatablePropertyNameStatus + "/softwareModuleType"
	softwareUpdatablePropertyLastOperation         = softwareUpdatablePropertyNameStatus + "/lastOperation"
	softwareUpdatablePropertyLastFailedOperation   = softwareUpdatablePropertyNameStatus + "/lastFailedOperation"
	softwareUpdatablePropertyInstalledDependencies = softwareUpdatablePropertyNameStatus + "/installedDependencies"
	softwareUpdatablePropertyContextDependencies   = softwareUpdatablePropertyNameStatus + "/contextDependencies"

	featureSoftwareUpdatableInstalledDependenciesKeyTemplate = "%s.%s:%s"

	softwareUpdatableOperationDownload     = "download"
	softwareUpdatableOperationInstall      = "install"
	softwareUpdatableOperationCancel       = "cancel"
	softwareUpdatableOperationRemove       = "remove"
	softwareUpdatableOperationCancelRemove = "cancelRemove"

	containersSoftwareUpdatableAgentType   = "oci:container"
	installedDependenciesKeysSlashEncoding = "%2F"
	installedDependenciesKeysSlash         = "/"
)

func (su *softwareUpdatable) createFeature() model.Feature {
	feature := client.NewFeature(SoftwareUpdatableFeatureID,
		client.WithFeatureDefinition(client.NewDefinitionID(SoftwareUpdatableDefinitionNamespace, SoftwareUpdatableDefinitionName, SoftwareUpdatableDefinitionVersion)),
		client.WithFeatureProperty(softwareUpdatablePropertyNameStatus, su.status),
		client.WithFeatureOperationsHandler(su.operationsHandler))
	return feature
}

func newSoftwareUpdatable(rootThing model.Thing, mgr mgr.ContainerManager, eventsMgr events.ContainerEventsManager) managedFeature {
	supStatus := &features.SoftwareUpdatableStatus{
		SoftwareModuleType: containersSoftwareUpdatableAgentType,
	}
	return &softwareUpdatable{
		rootThing: rootThing,
		status:    supStatus,
		mgr:       mgr,
		eventsMgr: eventsMgr,
	}
}

func (su *softwareUpdatable) register(ctx context.Context) error {
	log.Debug("initializing SoftwareUpdatable feature")

	ctrs, err := su.mgr.List(ctx)
	if err != nil {
		log.ErrorErr(err, "could not list containers for initialization of the SoftwareUpdatable feature's inventory")
	} else {
		su.processContainers(ctrs)
	}
	if su.cancelEventsHandler == nil {
		su.handleContainerEvents(ctx)
		log.Debug("subscribed for container events")
	}
	return su.rootThing.SetFeature(SoftwareUpdatableFeatureID, su.createFeature())
}

func (su *softwareUpdatable) dispose() {
	log.Debug("disposing SoftwareUpdatable feature")
	if su.cancelEventsHandler != nil {
		log.Debug("unsubscribing from container events")
		su.cancelEventsHandler()
		su.cancelEventsHandler = nil
	}
}

func (su *softwareUpdatable) processContainers(ctrs []*types.Container) {
	if ctrs != nil && len(ctrs) > 0 {
		su.status.InstalledDependencies = map[string]*datatypes.DependencyDescription{}
		for _, ctr := range ctrs {
			depDescr := dependencyDescription(ctr)
			su.status.InstalledDependencies[generateDependencyDescriptionKey(depDescr)] = depDescr
		}
	}
}

func (su *softwareUpdatable) operationsHandler(operationName string, args interface{}) (interface{}, error) {
	log.Debug("containers agent operation initiated - [operation = %s]", operationName)
	switch operationName {
	case softwareUpdatableOperationInstall:
		ua, err := convertToUpdateAction(args)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		return nil, su.Install(ua)
	case softwareUpdatableOperationRemove:
		ra, err := convertToRemoveAction(args)
		if err != nil {
			return nil, client.NewMessagesParameterInvalidError(err.Error())
		}
		return nil, su.Remove(ra)
	default:
		err := log.NewErrorf("unsupported operation called [operationId = %s]", operationName)
		log.ErrorErr(err, "unsupported operation")
		return nil, client.NewMessagesSubjectNotFound(err.Error())
	}
}

func (su *softwareUpdatable) updateLastOperation(operationStatus *datatypes.OperationStatus) {
	err := su.rootThing.SetFeatureProperty(SoftwareUpdatableFeatureID, softwareUpdatablePropertyLastOperation, operationStatus)
	if err != nil {
		log.ErrorErr(err, "error while updating lastOperation property")
	}
}

func (su *softwareUpdatable) updateLastFailedOperation(operationStatus *datatypes.OperationStatus) {
	err := su.rootThing.SetFeatureProperty(SoftwareUpdatableFeatureID, softwareUpdatablePropertyLastFailedOperation, operationStatus)
	if err != nil {
		log.ErrorErr(err, "error while updating lastFailedOperation property")
	}
}

func (su *softwareUpdatable) addInstalledDependency(dep *datatypes.DependencyDescription) {
	su.depsLock.Lock()
	defer su.depsLock.Unlock()
	if dep == nil {
		return
	}
	if su.status.InstalledDependencies == nil {
		su.status.InstalledDependencies = map[string]*datatypes.DependencyDescription{}
	}

	key := generateDependencyDescriptionKey(dep)
	su.status.InstalledDependencies[key] = dep
	log.Debug("updated local dependency for SoftwareUpdatable Id = %s, new entry = [%s:%s]", SoftwareUpdatableFeatureID, key, dep)

	if err := su.rootThing.SetFeatureProperty(SoftwareUpdatableFeatureID, softwareUpdatablePropertyInstalledDependencies, su.status.InstalledDependencies); err != nil {
		log.ErrorErr(err, "failed to update the installedDependencies property with newly installed dependencies")
	}
}
func (su *softwareUpdatable) removeInstalledDependency(dep *datatypes.DependencyDescription) {
	su.depsLock.Lock()
	defer su.depsLock.Unlock()
	if dep == nil {
		return
	}
	if su.status.InstalledDependencies != nil && len(su.status.InstalledDependencies) > 0 {
		key := generateDependencyDescriptionKey(dep)

		delete(su.status.InstalledDependencies, key)
		log.Debug("removed local dependency for SoftwareUpdatable Id = %s, new entry = [%s:%s]", SoftwareUpdatableFeatureID, key, dep)

		if len(su.status.InstalledDependencies) == 0 {
			su.status.InstalledDependencies = nil
		}

		if err := su.rootThing.SetFeatureProperty(SoftwareUpdatableFeatureID, softwareUpdatablePropertyInstalledDependencies, su.status.InstalledDependencies); err != nil {
			log.ErrorErr(err, "failed to update the installedDependencies property and remove all uninstalled dependencies")
		}
	}
}

func (su *softwareUpdatable) processRemoveAction(dsAction datatypes.RemoveAction) {
	su.processOperationsLock.Lock()
	defer su.processOperationsLock.Unlock()

	operationStatus := &datatypes.OperationStatus{
		CorrelationID: dsAction.CorrelationID,
	}
	cmpdError := &errorUtil.CompoundError{}

	var (
		rejectedRemove   []*datatypes.DependencyDescription
		errorRemove      []*datatypes.DependencyDescription
		successfulRemove []*datatypes.DependencyDescription
	)

	for _, toRemove := range dsAction.Software {
		// reject all others after the first fail when not forced
		if !dsAction.Forced && len(errorRemove) > 0 {
			rejectedRemove = append(rejectedRemove, toRemove)
		} else if err := su.removeDependency(toRemove, operationStatus); err != nil {
			cmpdError.Append(err)
			errorRemove = append(errorRemove, toRemove)
		} else {
			successfulRemove = append(successfulRemove, toRemove)
		}
	}

	if len(rejectedRemove) > 0 {
		operationStatus.Software = rejectedRemove
		operationStatus.Status = datatypes.FinishedRejected
		su.updateLastFailedOperation(operationStatus)
		su.updateLastOperation(operationStatus)
	}
	if len(errorRemove) > 0 {
		operationStatus.Software = errorRemove
		operationStatus.Status = datatypes.FinishedError
		operationStatus.Message = cmpdError.Error()
		su.updateLastFailedOperation(operationStatus)
		su.updateLastOperation(operationStatus)
	}
	if len(successfulRemove) > 0 {
		operationStatus.Software = successfulRemove
		operationStatus.Status = datatypes.FinishedSuccess
		operationStatus.Message = ""
		su.updateLastOperation(operationStatus)
	}
}

func (su *softwareUpdatable) removeDependency(toRemove *datatypes.DependencyDescription, operationStatus *datatypes.OperationStatus) (err error) {

	defer func() {
		// in case of panic change err at the very last moment
		if e := recover(); e != nil {
			log.Error("failed to remove Dependency [Name] = [%s] %v", toRemove.Name, e)
			err = log.NewError("internal runtime error")
		}
		if err == nil {
			// Removed
			operationStatus.Status = datatypes.Removed
			su.updateLastOperation(operationStatus)
		}
	}()

	// Removing
	operationStatus.Status = datatypes.Removing
	operationStatus.Software = []*datatypes.DependencyDescription{toRemove}
	su.updateLastOperation(operationStatus)

	ctx := context.Background()
	if ctr, _ := su.mgr.Get(ctx, toRemove.Name); ctr == nil {
		log.Warn("container with ID = %s does not exist", toRemove.Name)
		err = log.NewErrorf("container with ID = %s does not exist", toRemove.Name)
	} else {
		err = su.mgr.Remove(ctx, toRemove.Name, true, false, nil) // TODO currently matching only on Name - the container id
	}
	return err
}

func (su *softwareUpdatable) processUpdateAction(updateAction datatypes.UpdateAction) {
	su.processOperationsLock.Lock()
	defer su.processOperationsLock.Unlock()

	for _, softMod := range updateAction.SoftwareModules {
		su.installModule(softMod, updateAction.CorrelationID)
	}
}

func (su *softwareUpdatable) installModule(softMod *datatypes.SoftwareModuleAction, correlationID string) {
	log.Debug("will perform installation of SoftwareModule [Name.version] = [%s.%s]", softMod.SoftwareModule.Name, softMod.SoftwareModule.Version)
	operationStatus := &datatypes.OperationStatus{
		Status:         datatypes.Started,
		CorrelationID:  correlationID,
		SoftwareModule: softMod.SoftwareModule,
	}

	var (
		installError error
		rejected     bool
	)

	defer func() {
		// in case of panic report FinishedError
		if err := recover(); err != nil {
			log.Error("failed to install SoftwareModule [Name.version] = [%s.%s]  %v", softMod.SoftwareModule.Name, softMod.SoftwareModule.Version, err)
			operationStatus.Message = "internal runtime error"
			operationStatus.Status = datatypes.FinishedError
			su.updateLastFailedOperation(operationStatus)
		} else if installError != nil {
			operationStatus.Message = installError.Error()
			if rejected {
				operationStatus.Status = datatypes.FinishedRejected
			} else {
				operationStatus.Status = datatypes.FinishedError
			}
			su.updateLastFailedOperation(operationStatus)
		} else {
			operationStatus.Status = datatypes.FinishedSuccess
		}
		su.updateLastOperation(operationStatus)
	}()

	su.updateLastOperation(operationStatus)

	// Downloading
	operationStatus.Status = datatypes.Downloading
	su.updateLastOperation(operationStatus)

	containers := make([]*types.Container, len(softMod.Artifacts))
	for i, saa := range softMod.Artifacts {
		if containers[i], rejected, installError = createContainer(saa); installError != nil {
			log.ErrorErr(installError, "failed to create container from the provided SoftwareArtifact [FileName] = [%s]", saa.FileName)
			return
		}
	}

	// Downloaded
	operationStatus.Status = datatypes.Downloaded
	su.updateLastOperation(operationStatus)

	// Installing
	operationStatus.Status = datatypes.Installing
	su.updateLastOperation(operationStatus)

	ctx := context.Background()
	for i := range containers {
		if _, installError = su.mgr.Create(ctx, containers[i]); installError != nil {
			log.ErrorErr(installError, "failed to create container ID = %s", containers[i].ID)
			return
		}
	}
	// Installed
	operationStatus.Status = datatypes.Installed
	su.updateLastOperation(operationStatus)
	for i := range containers {
		if installError = su.mgr.Start(ctx, containers[i].ID); installError != nil {
			log.ErrorErr(installError, "failed to start container ID = %s", containers[i].ID)
			return
		}
	}
}
