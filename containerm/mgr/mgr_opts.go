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

package mgr

import (
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
)

// ContainerManagerOpt provides container manager options
type ContainerManagerOpt func(mgrOptions *mgrOpts) error

type mgrOpts struct {
	metaPath                 string
	rootExec                 string
	containerClientServiceID string
	networkManagerServiceID  string
	defaultCtrsStopTimeout   time.Duration
}

func applyOptsMgr(mgrOpts *mgrOpts, opts ...ContainerManagerOpt) error {
	for _, o := range opts {
		if err := o(mgrOpts); err != nil {
			return err
		}
	}
	return nil
}

// WithMgrMetaPath sets container manager meta path.
func WithMgrMetaPath(metaPath string) ContainerManagerOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.metaPath = metaPath
		return nil
	}
}

// WithMgrRootExec sets container manager root exec.
func WithMgrRootExec(rootExec string) ContainerManagerOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.rootExec = rootExec
		return nil
	}
}

// WithMgrContainerClientServiceID sets container manager client service ID.
func WithMgrContainerClientServiceID(containerClientServiceID string) ContainerManagerOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.containerClientServiceID = containerClientServiceID
		return nil
	}
}

// WithMgrNetworkManagerServiceID sets container manager network service ID.
func WithMgrNetworkManagerServiceID(networkManagerServiceID string) ContainerManagerOpt {
	return func(mgrOptions *mgrOpts) error {
		mgrOptions.networkManagerServiceID = networkManagerServiceID
		return nil
	}
}

// WithMgrDefaultContainerStopTimeout sets default container stop timeout.
func WithMgrDefaultContainerStopTimeout(managerCtrsStopTimeout interface{}) ContainerManagerOpt {
	return func(mgrOptions *mgrOpts) error {
		switch v := managerCtrsStopTimeout.(type) {
		case int64:
			mgrOptions.defaultCtrsStopTimeout = time.Duration(managerCtrsStopTimeout.(int64)) * time.Second
		case time.Duration:
			mgrOptions.defaultCtrsStopTimeout = managerCtrsStopTimeout.(time.Duration)
		default:
			return log.NewErrorf("unexpected stop timeout type: %v", v)
		}
		return nil
	}
}
