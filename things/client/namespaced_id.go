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

type namespacedID struct {
	namespace string
	name      string
}

// NewNamespacedID creates a new NamespacedID instance using the provided namespace and name
func NewNamespacedID(namespace string, name string) model.NamespacedID {
	return namespacedID{namespace: namespace, name: name}
}

// NewNamespacedIDFromString creates a new NamespacedID using the provided string
func NewNamespacedIDFromString(full string) model.NamespacedID {
	elements := strings.Split(full, ":")
	return namespacedID{namespace: elements[0], name: strings.Join(elements[1:], ":")}
}

// GetNamespace returns a namespace
func (namespacedId namespacedID) GetNamespace() string {
	return namespacedId.namespace
}

// GetName returns a name
func (namespacedId namespacedID) GetName() string {
	return namespacedId.name
}

// String provides the string representation of the NamespaceID
func (namespacedId namespacedID) String() string {
	return fmt.Sprintf("%s:%s", namespacedId.namespace, namespacedId.name)
}
