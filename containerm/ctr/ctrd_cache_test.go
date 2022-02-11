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
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/containerd/containerd"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	mocksCtrdpb "github.com/eclipse-kanto/container-management/containerm/pkg/testutil/mocks/containerd"
	"github.com/golang/mock/gomock"
)

const (
	testCtrID1 = "test-container-id1"
	testCtrID2 = "test-container-id2"
	testCtrID3 = "test-container-id3"
)

var (
	testCtr1 = &types.Container{
		ID: testCtrID1,
	}
	testCtr2 = &types.Container{
		ID: testCtrID2,
	}
	testCtr3 = &types.Container{
		ID: testCtrID3,
	}
	ctrInfo1 = &containerInfo{
		c:             testCtr1,
		skipExitHooks: false,
	}
	ctrInfo2 = &containerInfo{
		c:             testCtr2,
		skipExitHooks: true,
	}
	ctrInfo3 = &containerInfo{
		c:             testCtr3,
		skipExitHooks: true,
	}
)

func TestNewContainerInfoCache(t *testing.T) {
	t.Run("test_new_container_info_cache", func(t *testing.T) {
		ctrInfoCache := newContainerInfoCache()
		testutil.AssertNotNil(t, ctrInfoCache)
	})
}

func TestSetExitHooks(t *testing.T) {
	t.Run("test_set_exit_hooks", func(t *testing.T) {
		ctrInfoCache := newContainerInfoCache()
		ctrInfoCache.setExitHooks(testExitHook)

		testutil.AssertEqual(t, reflect.ValueOf(testExitHook).Pointer(), reflect.ValueOf(ctrInfoCache.containerExitHooks[0]).Pointer())
	})
}

func testExitHook(container *types.Container, code int64, err error, oomKilled bool, cleanup func() error) error {
	return nil
}

func TestRemove(t *testing.T) {
	t.Run("test_remove", func(t *testing.T) {
		ctrInfoCache := newContainerInfoCache()
		ctrInfoCache.cache[ctrInfo1.c.ID] = ctrInfo1
		ctrInfoCache.cache[ctrInfo2.c.ID] = ctrInfo2

		expectedCtrInfoCache := newContainerInfoCache()
		expectedCtrInfoCache.cache[ctrInfo1.c.ID] = ctrInfo1

		ctrInfoCache.remove(ctrInfo2.c.ID)

		testutil.AssertEqual(t, expectedCtrInfoCache, ctrInfoCache)

		// try to remove already removed info
		removedInfo := ctrInfoCache.remove(ctrInfo2.c.ID)
		testutil.AssertNil(t, removedInfo)
	})
}

func TestGet(t *testing.T) {
	t.Run("test_get", func(t *testing.T) {
		ctrInfoCache := newContainerInfoCache()
		ctrInfoCache.cache[ctrInfo1.c.ID] = ctrInfo1
		ctrInfoCache.cache[ctrInfo2.c.ID] = ctrInfo2

		testInfo := ctrInfoCache.get(testCtrID1)
		testutil.AssertEqual(t, testInfo, ctrInfo1)

		// try to get not cached info
		testInfo = ctrInfoCache.get(testCtrID3)
		testutil.AssertNil(t, testInfo)
	})
}

func TestGetAll(t *testing.T) {
	t.Run("test_get", func(t *testing.T) {
		ctrInfoCache := newContainerInfoCache()
		ctrInfoCache.cache[ctrInfo1.c.ID] = ctrInfo1
		ctrInfoCache.cache[ctrInfo2.c.ID] = ctrInfo2

		testInfo := ctrInfoCache.getAll()
		testutil.AssertEqual(t, 2, len(testInfo))

		found := 0
		for _, info := range testInfo {
			if info.c.ID == ctrInfo1.c.ID {
				testutil.AssertEqual(t, ctrInfo1, info)
				found++
			} else if info.c.ID == ctrInfo2.c.ID {
				testutil.AssertEqual(t, ctrInfo2, info)
				found++
			}
		}
		testutil.AssertEqual(t, 2, found)
	})
}

func TestSetContainerdDead(t *testing.T) {
	t.Run("test_is_containerd_dead", func(t *testing.T) {
		ctrInfoCache := newContainerInfoCache()
		ctrInfoCache.cache[ctrInfo1.c.ID] = ctrInfo1
		ctrInfoCache.setContainerdDead(true)

		testutil.AssertTrue(t, ctrInfoCache.containerdStopped)
	})
}

func TestIsContainerdDead(t *testing.T) {
	t.Run("test_is_containerd_dead", func(t *testing.T) {
		ctrInfoCache := newContainerInfoCache()
		ctrInfoCache.cache[ctrInfo1.c.ID] = ctrInfo1
		ctrInfoCache.containerdStopped = true

		testutil.AssertTrue(t, ctrInfoCache.isContainerdDead())
	})
}

func TestIsContainerdConnectable(t *testing.T) {
	t.Run("test_is_containerd_connectable", func(t *testing.T) {
		ctrInfoCache := newContainerInfoCache()
		ctrInfoCache.cache[ctrInfo1.c.ID] = ctrInfo1
		ctrdExitStatus := containerd.NewExitStatus(0, time.Now(), errors.New("rpc error"))
		testutil.AssertFalse(t, isContainerdConnectable(*ctrdExitStatus))

		//test no error
		ctrdExitStatus = containerd.NewExitStatus(0, time.Now(), nil)
		testutil.AssertFalse(t, isContainerdConnectable(*ctrdExitStatus))

	})
}

func TestSetAndGetTask(t *testing.T) {
	t.Run("test_set_and_get_task", func(t *testing.T) {
		mockCtrl := gomock.NewController(t)
		defer mockCtrl.Finish()
		mockCtrdTask := mocksCtrdpb.NewMockTask(mockCtrl)
		ctrInfo1.setTask(mockCtrdTask)
		testutil.AssertEqual(t, mockCtrdTask, ctrInfo1.task)
		testutil.AssertEqual(t, mockCtrdTask, ctrInfo1.getTask())
	})
}

func TestOOMKilled(t *testing.T) {
	t.Run("test_oom_killed", func(t *testing.T) {
		testutil.AssertFalse(t, ctrInfo1.isOOmKilled())
		ctrInfo1.setOOMKilled(true)
		testutil.AssertTrue(t, ctrInfo1.isOOmKilled())
	})
}
