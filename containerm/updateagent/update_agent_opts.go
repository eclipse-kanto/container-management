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
	"time"
)

// ContainersUpdateAgentOpt represents the available configuration options for the Containers UpdateAgent service
type ContainersUpdateAgentOpt func(updateAgentOptions *updateAgentOpts) error

type updateAgentOpts struct {
	domainName             string
	systemContainers       []string
	verboseInventoryReport bool
	broker                 string
	keepAlive              time.Duration
	disconnectTimeout      time.Duration
	clientUsername         string
	clientPassword         string
	connectTimeout         time.Duration
	acknowledgeTimeout     time.Duration
	subscribeTimeout       time.Duration
	unsubscribeTimeout     time.Duration
	tlsConfig              *tlsConfig
}

// tls-secured communication config
type tlsConfig struct {
	RootCA     string
	ClientCert string
	ClientKey  string
}

func applyOptsUpdateAgent(updateAgentOpts *updateAgentOpts, opts ...ContainersUpdateAgentOpt) error {
	for _, o := range opts {
		if err := o(updateAgentOpts); err != nil {
			return err
		}
	}
	return nil
}

// WithDomainName configures the domain name for the containers update agent
func WithDomainName(domain string) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.domainName = domain
		return nil
	}
}

// WithSystemContainers configures the list of system containers (names) that will not be processed by the containers update agent
func WithSystemContainers(systemContainers []string) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.systemContainers = systemContainers
		return nil
	}
}

// WithVerboseInventoryReport enables / disables verbose inventory reporting of current containers
func WithVerboseInventoryReport(verboseInventoryReport bool) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.verboseInventoryReport = verboseInventoryReport
		return nil
	}
}

// WithConnectionBroker configures the broker, where the connection will be established
func WithConnectionBroker(broker string) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.broker = broker
		return nil
	}
}

// WithConnectionKeepAlive configures the time between between each check for the connection presence
func WithConnectionKeepAlive(keepAlive time.Duration) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.keepAlive = keepAlive
		return nil
	}
}

// WithConnectionDisconnectTimeout configures the duration of inactivity before disconnecting from the broker
func WithConnectionDisconnectTimeout(disconnectTimeout time.Duration) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.disconnectTimeout = disconnectTimeout
		return nil
	}
}

// WithConnectionClientUsername configures the client username used when establishing connection to the broker
func WithConnectionClientUsername(username string) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.clientUsername = username
		return nil
	}
}

// WithConnectionClientPassword configures the client password used when establishing connection to the broker
func WithConnectionClientPassword(password string) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.clientPassword = password
		return nil
	}
}

// WithConnectionConnectTimeout configures the timeout before terminating the connect attempt
func WithConnectionConnectTimeout(connectTimeout time.Duration) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.connectTimeout = connectTimeout
		return nil
	}
}

// WithConnectionAcknowledgeTimeout configures the timeout for the acknowledge receival
func WithConnectionAcknowledgeTimeout(acknowledgeTimeout time.Duration) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.acknowledgeTimeout = acknowledgeTimeout
		return nil
	}
}

// WithConnectionSubscribeTimeout configures the timeout before terminating the subscribe attempt
func WithConnectionSubscribeTimeout(subscribeTimeout time.Duration) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.subscribeTimeout = subscribeTimeout
		return nil
	}
}

// WithConnectionUnsubscribeTimeout configures the timeout before terminating the unsubscribe attempt
func WithConnectionUnsubscribeTimeout(unsubscribeTimeout time.Duration) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.unsubscribeTimeout = unsubscribeTimeout
		return nil
	}
}

// WithTLSConfig configures the CA certificate for TLS communication
func WithTLSConfig(rootCA, clientCert, clientKey string) ContainersUpdateAgentOpt {
	return func(updateAgentOptions *updateAgentOpts) error {
		updateAgentOptions.tlsConfig = &tlsConfig{
			RootCA:     rootCA,
			ClientCert: clientCert,
			ClientKey:  clientKey,
		}
		return nil
	}
}
