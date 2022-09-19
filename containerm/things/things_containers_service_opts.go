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

import "time"

// ContainerThingsManagerOpt represents the available configuration options for the ContainerThingsManager service
type ContainerThingsManagerOpt func(thingsOptions *thingsOpts) error

type thingsOpts struct {
	broker             string
	keepAlive          time.Duration
	disconnectTimeout  time.Duration
	clientUsername     string
	clientPassword     string
	storagePath        string
	featureIds         []string
	connectTimeout     time.Duration
	acknowledgeTimeout time.Duration
	subscribeTimeout   time.Duration
	unsubscribeTimeout time.Duration
}

func applyOptsThings(thingsOpts *thingsOpts, opts ...ContainerThingsManagerOpt) error {
	for _, o := range opts {
		if err := o(thingsOpts); err != nil {
			return err
		}
	}
	return nil
}

// WithMetaPath configures the directory to be used for storage by the service
func WithMetaPath(path string) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.storagePath = path
		return nil
	}
}

// WithFeatures configures the container runtime's Things representation via providing the desired Ditto Features to be created by ID
func WithFeatures(featureIds []string) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.featureIds = featureIds
		return nil
	}
}

// WithConnectionBroker configures the broker, where the connection will be established
func WithConnectionBroker(broker string) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.broker = broker
		return nil
	}
}

// WithConnectionKeepAlive configures the time between between each check for the connection presence
func WithConnectionKeepAlive(keepAlive time.Duration) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.keepAlive = keepAlive
		return nil
	}
}

// WithConnectionDisconnectTimeout configures the duration of inactivity before disconnecting from the broker
func WithConnectionDisconnectTimeout(disconnectTimeout time.Duration) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.disconnectTimeout = disconnectTimeout
		return nil
	}
}

// WithConnectionClientUsername configures the client username used when establishing connection to the broker
func WithConnectionClientUsername(username string) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.clientUsername = username
		return nil
	}
}

// WithConnectionClientPassword configures the client password used when establishing connection to the broker
func WithConnectionClientPassword(password string) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.clientPassword = password
		return nil
	}
}

// WithConnectionConnectTimeout configures the timeout before terminating the connect attempt
func WithConnectionConnectTimeout(connectTimeout time.Duration) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.connectTimeout = connectTimeout
		return nil
	}
}

// WithConnectionAcknowledgeTimeout configures the timeout for the acknowledge receival
func WithConnectionAcknowledgeTimeout(acknowledgeTimeout time.Duration) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.acknowledgeTimeout = acknowledgeTimeout
		return nil
	}
}

// WithConnectionSubscribeTimeout configures the timeout before terminating the subscribe attempt
func WithConnectionSubscribeTimeout(subscribeTimeout time.Duration) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.subscribeTimeout = subscribeTimeout
		return nil
	}
}

// WithConnectionUnsubscribeTimeout configures the timeout before terminating the unsubscribe attempt
func WithConnectionUnsubscribeTimeout(unsubscribeTimeout time.Duration) ContainerThingsManagerOpt {
	return func(thingsOptions *thingsOpts) error {
		thingsOptions.unsubscribeTimeout = unsubscribeTimeout
		return nil
	}
}
