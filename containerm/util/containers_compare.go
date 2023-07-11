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

package util

import (
	"fmt"
	"reflect"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
)

// ActionType defines a type for an action to achieve desired container
type ActionType int

const (
	// ActionCheck denotes that current container already has the desired configuration, it shall be checked only if the container is running
	ActionCheck ActionType = iota
	// ActionCreate denotes that a new container with desired configurtion shall be created and started
	ActionCreate
	// ActionRecreate denotes that the existing container shall be replaced by a new container with desired configurtion
	ActionRecreate
	// ActionUpdate denotes that the existing container shall be runtime updated with desired configurtion
	ActionUpdate
	// ActionDestroy denotes that the existing container shall be destroyed
	ActionDestroy
)

// DetermineUpdateAction compares the current container with the desired one and determines what action shall be done to achieve desired state
func DetermineUpdateAction(current *types.Container, desired *types.Container) ActionType {
	if current == nil {
		return ActionCreate
	}
	if !isEqualImage(current.Image, desired.Image) {
		return ActionRecreate
	}
	if !compareSliceSet(current.Mounts, desired.Mounts) {
		return ActionRecreate
	}
	if !isEqualContainerConfig(current.Config, desired.Config) {
		return ActionRecreate
	}
	if !isEqualIOConfig(current.IOConfig, desired.IOConfig) {
		return ActionRecreate
	}
	if !isEqualHostConfig0(current.HostConfig, desired.HostConfig) {
		return ActionRecreate
	}
	if !isEqualHostConfig1(current.HostConfig, desired.HostConfig) {
		return ActionUpdate
	}
	return ActionCheck
}

// GetActionMessage returns a text message describing the given action type
func GetActionMessage(actionType ActionType) string {
	switch actionType {
	case ActionCheck:
		return "No changes detected, existing container will be check only if it is running."
	case ActionCreate:
		return "New container will be created and started."
	case ActionRecreate:
		return "Existing container will be destroyed and replaced by a new one."
	case ActionUpdate:
		return "Existing container will be updated with new configuration."
	case ActionDestroy:
		return "Existing container will be destroyed, no longer needed."
	}
	return "Unknown action type: " + fmt.Sprint(actionType)
}

func isEqualImage(currentImage types.Image, newImage types.Image) bool {
	return currentImage.Name == newImage.Name
}

func isEqualContainerConfig(currentContainerCfg *types.ContainerConfiguration, newContainerCfg *types.ContainerConfiguration) bool {
	if currentContainerCfg == nil {
		return newContainerCfg == nil
	}
	if newContainerCfg == nil {
		return false
	}
	// Cmd is order-sensitive, that's why compare Cmd contents with reflect.DeepEqual
	if !(len(currentContainerCfg.Cmd) == 0 && len(newContainerCfg.Cmd) == 0) && !reflect.DeepEqual(currentContainerCfg.Cmd, newContainerCfg.Cmd) {
		return false
	}

	return compareSliceSet(currentContainerCfg.Env, newContainerCfg.Env)
}

func isEqualHostConfig0(currentHostConfig *types.HostConfig, newHostConfig *types.HostConfig) bool {
	if currentHostConfig == nil {
		return newHostConfig == nil
	}
	if newHostConfig == nil {
		return false
	}
	if currentHostConfig.Privileged != newHostConfig.Privileged {
		return false
	}
	if currentHostConfig.NetworkMode != newHostConfig.NetworkMode {
		return false
	}
	if !compareSliceSet(currentHostConfig.Devices, newHostConfig.Devices) {
		return false
	}
	if !compareSliceSet(currentHostConfig.ExtraHosts, newHostConfig.ExtraHosts) {
		return false
	}
	if !compareSliceSet(currentHostConfig.ExtraCaps, newHostConfig.ExtraCaps) {
		return false
	}
	if !compareSliceSet(currentHostConfig.PortMappings, newHostConfig.PortMappings) {
		return false
	}
	if !isEqualLog(currentHostConfig.LogConfig, newHostConfig.LogConfig) {
		return false
	}

	return true
}

func isEqualHostConfig1(currentHostConfig *types.HostConfig, newHostConfig *types.HostConfig) bool {
	if currentHostConfig == nil {
		return newHostConfig == nil
	}
	if !isEqualResources(currentHostConfig.Resources, newHostConfig.Resources) {
		return false
	}
	if !isEqualRestartPolicy(currentHostConfig.RestartPolicy, newHostConfig.RestartPolicy) {
		return false
	}
	return true
}

func isEqualResources(currentResources *types.Resources, newResources *types.Resources) bool {
	if currentResources == nil {
		return newResources == nil
	}
	if newResources == nil {
		return false
	}
	return *currentResources == *newResources
}

func isEqualRestartPolicy(currentRestartPolicy *types.RestartPolicy, newRestartPolicy *types.RestartPolicy) bool {
	if currentRestartPolicy == nil {
		return newRestartPolicy == nil
	}
	if newRestartPolicy == nil {
		return false
	}
	return *currentRestartPolicy == *newRestartPolicy
}

func isEqualLog(currentLogConfig *types.LogConfiguration, newLogConfig *types.LogConfiguration) bool {
	if currentLogConfig == nil {
		return newLogConfig == nil
	}
	if newLogConfig == nil {
		return false
	}
	if currentLogConfig.DriverConfig == nil {
		return newLogConfig.DriverConfig == nil
	}
	if newLogConfig.DriverConfig == nil {
		return false
	}
	if *currentLogConfig.DriverConfig != *newLogConfig.DriverConfig {
		return false
	}
	if currentLogConfig.ModeConfig == nil {
		return newLogConfig.ModeConfig == nil
	}
	if newLogConfig.ModeConfig == nil {
		return false
	}
	if *currentLogConfig.ModeConfig != *newLogConfig.ModeConfig {
		return false
	}
	return true
}

func isEqualIOConfig(currentIOConfig *types.IOConfig, newIOConfig *types.IOConfig) bool {
	if currentIOConfig == nil {
		return newIOConfig == nil
	}

	if newIOConfig == nil {
		return false
	}
	return currentIOConfig.OpenStdin == newIOConfig.OpenStdin && currentIOConfig.Tty == newIOConfig.Tty
}

func compareSliceSet(firstSet interface{}, secondSet interface{}) bool {
	firstValue := reflect.ValueOf(firstSet)
	secondValue := reflect.ValueOf(secondSet)
	if firstValue.Len() != secondValue.Len() {
		return false
	}
	for firstIndex := 0; firstIndex < firstValue.Len(); firstIndex++ {
		if !sliceContains(secondValue, firstValue.Index(firstIndex)) {
			return false
		}
	}
	for secondIndex := 0; secondIndex < secondValue.Len(); secondIndex++ {
		if !sliceContains(firstValue, secondValue.Index(secondIndex)) {
			return false
		}
	}
	return true
}

func sliceContains(slice reflect.Value, element reflect.Value) bool {
	for i := 0; i < slice.Len(); i++ {
		if slice.Index(i).Interface() == element.Interface() {
			return true
		}
	}
	return false
}
