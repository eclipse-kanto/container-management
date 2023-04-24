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

package deployment

import "github.com/eclipse-kanto/container-management/containerm/log"

// Opt provides deployment manager options
type Opt func(options *opts) error

type opts struct {
	mode     Mode
	metaPath string
	ctrPath  string
}

func applyOpts(options *opts, opts ...Opt) error {
	for _, o := range opts {
		if err := o(options); err != nil {
			return err
		}
	}
	return nil
}

// WithMetaPath configures the directory to be used for storage by the service
func WithMetaPath(metaPath string) Opt {
	return func(dOpts *opts) error {
		dOpts.metaPath = metaPath
		return nil
	}
}

// WithCtrPath sets the path to container descriptors
func WithCtrPath(ctrPath string) Opt {
	return func(dOpts *opts) error {
		dOpts.ctrPath = ctrPath
		return nil
	}
}

// WithMode sets the mode of deployment service
func WithMode(mode string) Opt {
	return func(dOpts *opts) error {
		dOpts.mode = toDeploymentMode(mode)
		return nil
	}
}

func toDeploymentMode(mode string) Mode {
	switch mode {
	case string(InitialDeployMode):
		return InitialDeployMode
	case string(UpdateMode):
		return UpdateMode
	}
	defValue := UpdateMode
	log.Warn("Invalid value '%s' for deployment mode option, switching to default mode %s", mode, defValue)
	return defValue
}
