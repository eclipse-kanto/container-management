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
	"fmt"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

func TestApplyOptsUpdateAgent(t *testing.T) {
	uaOpts := &updateAgentOpts{}
	options := []ContainersUpdateAgentOpt{
		WithDomainName("CONTAINERS"),
		WithSystemContainers([]string{"corelib", "systemlib"}),
		WithVerboseInventoryReport(true),
		WithConnectionBroker("127.0.0.1:18883"),
		WithConnectionClientUsername("client-username"),
		WithConnectionClientPassword("client-secret"),
		WithConnectionKeepAlive(10 * time.Second),
		WithConnectionDisconnectTimeout(20 * time.Second),
		WithConnectionConnectTimeout(30 * time.Second),
		WithConnectionAcknowledgeTimeout(40 * time.Second),
		WithConnectionSubscribeTimeout(50 * time.Second),
		WithConnectionUnsubscribeTimeout(60 * time.Second),
		WithTLSConfig("./certs/ca.cer", "./certs/client.cer", "./certs/client.key"),
	}
	testutil.AssertNil(t, applyOptsUpdateAgent(uaOpts, options...))

	testutil.AssertEqual(t, "CONTAINERS", uaOpts.domainName)
	testutil.AssertEqual(t, []string{"corelib", "systemlib"}, uaOpts.systemContainers)
	testutil.AssertTrue(t, uaOpts.verboseInventoryReport)
	testutil.AssertEqual(t, "127.0.0.1:18883", uaOpts.broker)
	testutil.AssertEqual(t, "client-username", uaOpts.clientUsername)
	testutil.AssertEqual(t, "client-secret", uaOpts.clientPassword)
	testutil.AssertEqual(t, 10*time.Second, uaOpts.keepAlive)
	testutil.AssertEqual(t, 20*time.Second, uaOpts.disconnectTimeout)
	testutil.AssertEqual(t, 30*time.Second, uaOpts.connectTimeout)
	testutil.AssertEqual(t, 40*time.Second, uaOpts.acknowledgeTimeout)
	testutil.AssertEqual(t, 50*time.Second, uaOpts.subscribeTimeout)
	testutil.AssertEqual(t, 60*time.Second, uaOpts.unsubscribeTimeout)
	testutil.AssertEqual(t, "./certs/ca.cer", uaOpts.tlsConfig.RootCA)
	testutil.AssertEqual(t, "./certs/client.cer", uaOpts.tlsConfig.ClientCert)
	testutil.AssertEqual(t, "./certs/client.key", uaOpts.tlsConfig.ClientKey)
}

func TestApplyOptsUpdateAgentWithError(t *testing.T) {
	uaOpts := &updateAgentOpts{}
	options := []ContainersUpdateAgentOpt{
		WithDomainName("ctrs"),
		func(updateAgentOptions *updateAgentOpts) error {
			return fmt.Errorf("wrong option")
		},
	}
	testutil.AssertNotNil(t, applyOptsUpdateAgent(uaOpts, options...))
}
