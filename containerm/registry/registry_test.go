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

package registry

import (
	"context"
	"testing"

	"github.com/containerd/containerd/errdefs"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/pkg/errors"
)

const (
	DummyManagerService Type = "container-management.service.dummy.manager.v1"
	MissingService      Type = "container-management.service.missing.v1"
)

type DummyManager interface{}

type dummyManager struct {
}

var (
	testCtx                 context.Context = context.Background()
	testServiceInfoSet                      = NewServiceInfoSet()
	testID                  string          = "dummy-manager-service-id"
	testID2                 string          = "dummy-manager-service-id2"
	testService             DummyManager    = dummyManager{}
	testService2            DummyManager    = dummyManager{}
	testServiceRegistration                 = &Registration{
		ID:   testID,
		Type: DummyManagerService,
		InitFunc: func(registryCtx *ServiceRegistryContext) (interface{}, error) {
			return testService, nil
		},
	}
	testServiceRegistration2 = &Registration{
		ID:   testID2,
		Type: DummyManagerService,
	}
	testServiceInfo = &ServiceInfo{
		Registration: testServiceRegistration,
		instance:     testService,
	}
	testServiceInfo2 = &ServiceInfo{
		Registration: testServiceRegistration2,
		instance:     testService2,
	}
	testServiceInfos = []*ServiceInfo{testServiceInfo, testServiceInfo2}
	testRegistryCtx  = NewContext(testCtx, nil, testServiceRegistration, testServiceInfoSet)
)

func TestAdd(t *testing.T) {
	t.Run("test_add_service", func(t *testing.T) {
		service, err := testRegistryCtx.Get(DummyManagerService)
		if service != nil {
			t.Errorf("nil expected for non existing service")
		}
		testutil.AssertError(t, log.NewErrorf("no services registered for %s", DummyManagerService), err)
		err = testServiceInfoSet.Add(testServiceInfo)
		if err != nil {
			t.Errorf("no error is expected when testServiceInfo is added")
		}
	})
	t.Run("test_add_service_same_type_different_id", func(t *testing.T) {
		err := testServiceInfoSet.Add(testServiceInfo2)
		if err != nil {
			t.Errorf("no error is expected when testServiceInfo2 is added")
		}
	})
	t.Run("test_add_service_same_type_same_id", func(t *testing.T) {
		err := testServiceInfoSet.Add(testServiceInfo)
		testutil.AssertError(t, log.NewErrorf("service %v already initialized", testID), err)
	})
}

func TestGet(t *testing.T) {
	t.Run("test_get_not_existing_service", func(t *testing.T) {
		service, err := testRegistryCtx.Get(MissingService)
		if service != nil {
			t.Errorf("nil expected for non existing service")
		}
		testutil.AssertError(t, log.NewErrorf("no services registered for %s", MissingService), err)
	})
	t.Run("test_get", func(t *testing.T) {
		service, err := testRegistryCtx.Get(DummyManagerService)
		testutil.AssertEqual(t, testService, service)
		if err != nil {
			t.Errorf("no error is expected when there is existing service")
		}
	})
}

func TestGetByType(t *testing.T) {
	t.Run("test_get_by_type_not_existing_service", func(t *testing.T) {
		serviceInfo, errInfo := testRegistryCtx.GetByType(MissingService)
		if serviceInfo != nil {
			t.Errorf("nil expected for non existing service")
		}
		testutil.AssertError(t, errors.Wrapf(errdefs.ErrNotFound, "no services registered for %s", MissingService), errInfo)
	})
	t.Run("test_get_by_type", func(t *testing.T) {
		serviceInfo, errInfo := testRegistryCtx.GetByType(DummyManagerService)
		if errInfo != nil {
			t.Errorf("no error is expected when there is existing service info")
		}
		service, err := serviceInfo[testID].Instance()
		testutil.AssertEqual(t, testService, service)
		if err != nil {
			t.Errorf("no error is expected when there is existing testService instance")
		}
		service, err = serviceInfo[testID2].Instance()
		testutil.AssertEqual(t, testService2, service)
		if err != nil {
			t.Errorf("no error is expected when there is existing testService2 instance")
		}
	})
}

func TestGetAll(t *testing.T) {
	t.Run("test_get_all", func(t *testing.T) {
		testutil.AssertEqual(t, testServiceInfos, testRegistryCtx.GetAll())
	})
	t.Run("test_get_all_with_type", func(t *testing.T) {
		testutil.AssertEqual(t, []*ServiceInfo{}, testServiceInfoSet.GetAll(MissingService))
		testutil.AssertEqual(t, testServiceInfos, testServiceInfoSet.GetAll(DummyManagerService))
	})
}

func TestRegistry(t *testing.T) {
	t.Run("test_init", func(t *testing.T) {
		serviceInfo := testServiceRegistration.Init(testRegistryCtx)
		testutil.AssertEqual(t, testServiceInfo, serviceInfo)
	})
	t.Run("test_register", func(t *testing.T) {
		Register(testServiceRegistration)
		testutil.AssertEqual(t, []*Registration{testServiceRegistration}, RegistrationsMap()[DummyManagerService])
	})
}
