// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package ctr

import (
	"context"
	"testing"

	"github.com/containerd/containerd/events"
	"github.com/containerd/containerd/leases"
	"github.com/containerd/containerd/namespaces"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	ctrdMocks "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctrd"
	"github.com/golang/mock/gomock"
)

func TestSubscribe(t *testing.T) {
	const testNamespace = "test-namespace"

	testChanEnv := make(<-chan *events.Envelope)
	testChanErr := make(<-chan error, 1)
	testFilters := []string{"namespace==" + testNamespace + ",topic~=tasks/oom.*"}

	// mock
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockCtrdWrapper := ctrdMocks.NewMockcontainerClientWrapper(mockCtrl)
	ctx := context.Background()
	testSpi := &ctrdSpi{client: mockCtrdWrapper, lease: &leases.Lease{ID: containerManagementLeaseID}, namespace: testNamespace}

	expectedContext := namespaces.WithNamespace(ctx, testNamespace)
	expectedContext = leases.WithLease(expectedContext, containerManagementLeaseID)
	mockCtrdWrapper.EXPECT().Subscribe(expectedContext, testFilters).Times(1).Return(testChanEnv, testChanErr)

	// test
	actualChanEnv, actualChanErr := testSpi.Subscribe(ctx, testFilters...)
	testutil.AssertEqual(t, testChanEnv, actualChanEnv)
	testutil.AssertEqual(t, testChanErr, actualChanErr)
}
