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

package ctr

import (
	"context"
	"testing"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/ctrd"
	"github.com/golang/mock/gomock"
)

func TestCtrdSpiDispose(t *testing.T) {
	testCases := map[string]struct {
		mapExec func(mockCtrdWrapper *ctrd.MockcontainerClientWrapper) error
	}{
		"test_no_err": {
			mapExec: func(mockCtrdWrapper *ctrd.MockcontainerClientWrapper) error {
				mockCtrdWrapper.EXPECT().Close().Return(nil)
				return nil
			},
		},
		"test_err": {
			mapExec: func(mockCtrdWrapper *ctrd.MockcontainerClientWrapper) error {
				err := log.NewError("test error")
				mockCtrdWrapper.EXPECT().Close().Return(err)
				return err
			},
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			// init mock ctrl
			mockCtrl := gomock.NewController(t)
			defer mockCtrl.Finish()
			// init mocks
			mockCtrdWrapper := ctrd.NewMockcontainerClientWrapper(mockCtrl)
			// mock exec
			expectedErr := testData.mapExec(mockCtrdWrapper)
			// init spi under test
			testSpi := &ctrdSpi{
				client: mockCtrdWrapper,
			}
			// test
			actualErr := testSpi.Dispose(context.Background())
			testutil.AssertError(t, expectedErr, actualErr)
		})
	}
}
