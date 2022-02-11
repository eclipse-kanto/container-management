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

package client

import (
	"fmt"
	"strings"

	"github.com/eclipse-kanto/container-management/things/api/model"
)

type definitionID struct {
	namespace string
	name      string
	version   string
}

// NewDefinitionIDFromString creates a new DefinitionId instance from the provided string
func NewDefinitionIDFromString(full string) model.DefinitionID {
	elements := strings.Split(full, ":")
	return definitionID{namespace: elements[0], name: elements[1], version: elements[2]}
}

// NewDefinitionID creates a new DefinitionId instance with namespace, name and version provided
func NewDefinitionID(namespace string, name string, version string) model.DefinitionID {
	return definitionID{namespace: namespace, name: name, version: version}
}

// GetNamespace returns a definition ID namespace
func (definitionId definitionID) GetNamespace() string {
	return definitionId.namespace
}

// GetName returns a definition ID name
func (definitionId definitionID) GetName() string {
	return definitionId.name
}

// GetVersion returns a definition ID version
func (definitionId definitionID) GetVersion() string {
	return definitionId.version
}

// String provides the string representation of definition ID namespace
func (definitionId definitionID) String() string {
	return fmt.Sprintf("%s:%s:%s", definitionId.namespace, definitionId.name, definitionId.version)
}
