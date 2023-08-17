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
	"strings"

	ctrtypes "github.com/eclipse-kanto/container-management/containerm/containers/types"

	"github.com/eclipse-kanto/update-manager/api/types"
	"github.com/pkg/errors"
)

func findContainerVersion(imageName string) string {
	if len(imageName) > 0 {
		name := imageName[strings.LastIndex(imageName, "/")+1:]
		sep := strings.Index(name, "@")
		if sep != -1 && sep != len(name)-1 {
			return name[sep+1:]
		}
		sep = strings.Index(name, ":")
		if sep != -1 && sep != len(name)-1 {
			return name[sep+1:]
		}
	}
	return "n/a"
}

func baselinesWithContainers(prefix string, baselines []*types.Baseline, containers map[string]*ctrtypes.Container) (map[string][]*ctrtypes.Container, error) {
	result := make(map[string][]*ctrtypes.Container)
	for _, baseline := range baselines {
		baselineContainers, err := containersForBaseline(prefix, baseline.Components, containers)
		if err != nil {
			return nil, errors.Wrap(err, "problem with baseline "+baseline.Title)
		}
		result[baseline.Title] = baselineContainers
	}
	for name, container := range containers { // all containers that are not included in a baseline are mapped to single-container baselines
		result[prefix+name] = []*ctrtypes.Container{container}
	}
	return result, nil
}

func containersForBaseline(prefix string, components []string, containers map[string]*ctrtypes.Container) ([]*ctrtypes.Container, error) {
	result := []*ctrtypes.Container{}
	for _, component := range components {
		if strings.HasPrefix(component, prefix) {
			name := component[len(prefix):]
			container, ok := containers[name]
			if !ok {
				return nil, errors.New("cannot find container component " + component)
			}
			result = append(result, container)
			delete(containers, name)
		}
	}
	return result, nil
}
