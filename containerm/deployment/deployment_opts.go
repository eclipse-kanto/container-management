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

// Opt provides deployment manager options
type Opt func(options *opts) error

type opts struct {
	initialDeployPath string
}

func applyOpts(options *opts, opts ...Opt) error {
	for _, o := range opts {
		if err := o(options); err != nil {
			return err
		}
	}
	return nil
}

// WithInitialDeployPath sets container initial deployment path.
func WithInitialDeployPath(initialDeployPath string) Opt {
	return func(dOpts *opts) error {
		dOpts.initialDeployPath = initialDeployPath
		return nil
	}
}
