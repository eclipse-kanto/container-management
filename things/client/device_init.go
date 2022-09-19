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

package client

import (
	"errors"

	"github.com/eclipse-kanto/container-management/things/api/model"
)

// DeviceOpt represents a device options
type DeviceOpt func(devOpt *devOpts) error

func applyOptsDevice(devOpt *devOpts, opts ...DeviceOpt) error {
	for _, o := range opts {
		if err := o(devOpt); err != nil {
			return err
		}
	}
	return validateOptsDevice(devOpt)
}

func validateOptsDevice(opts *devOpts) error {
	if opts.viaGateway.GetName() == "" || opts.viaGateway.GetNamespace() == "" {
		return errors.New("device id must be provided compliant with the format <namespace>:<name>")
	}
	if opts.tenantID == "" {
		return errors.New("device tenant id must be provided")
	}
	if opts.credentials == nil || opts.credentials.authID == "" || opts.credentials.password == "" {
		return errors.New("full device credentials must be provided [authID, password]")
	}
	return nil
}

type devOpts struct {
	id          model.NamespacedID
	viaGateway  model.NamespacedID
	tenantID    string
	credentials *credentials
}

// WithDeviceID sets a device option ID
func WithDeviceID(id string) DeviceOpt {
	return func(devOpt *devOpts) error {
		devOpt.id = NewNamespacedIDFromString(id)
		return nil
	}
}

// WithViaGateway sets a device option via gateway
func WithViaGateway(id string) DeviceOpt {
	return func(devOpt *devOpts) error {
		devOpt.viaGateway = NewNamespacedIDFromString(id)
		return nil
	}
}

// WithDeviceTenantID sets a device option tenant ID
func WithDeviceTenantID(tenantID string) DeviceOpt {
	return func(devOpt *devOpts) error {
		devOpt.tenantID = tenantID
		return nil
	}
}

// WithDeviceCredentials sets a device option credentials
func WithDeviceCredentials(authID string, pass string) DeviceOpt {
	return func(devOpt *devOpts) error {
		devOpt.credentials = &credentials{
			authID:   authID,
			password: pass,
		}
		return nil
	}
}
