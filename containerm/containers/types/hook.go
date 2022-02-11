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

package types

// HookType represents a hook type
type HookType int

// constants for the supported hook types
const (
	HookTypePrestart HookType = iota
	HookTypePoststart
	HookTypePoststop
	HookTypeUnknown
)

// String the string representation of the hook type
func (hookType HookType) String() string {
	return [...]string{"Prestart", "Poststart", "Poststop", "Unknown"}[hookType]
}

// Hook enables injection of actions to be performed throughout different stages of the container's OCI lifecycle
type Hook struct {
	// Path to the executable logic relevant for this hook's execution
	Path string `json:"path"`
	// Args is the hook arguments
	Args []string `json:"args"`
	// Env is the environmental variables needed for the hook's execution
	Env []string `json:"env"`
	// Timeout is the timeout for the hook's execution
	Timeout int `json:"timeout"`
	// Type is the type of the hook
	Type HookType `json:"type"` //prestart, poststart, poststop
}
