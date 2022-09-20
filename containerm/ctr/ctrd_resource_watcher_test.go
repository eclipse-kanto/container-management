// Copyright (c) 2022 Contributors to the Eclipse Foundation
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

package ctr

import (
	"context"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestNewResourceWatcher(t *testing.T) {
	ctx := context.Background()

	expectedCtx, expectedCtxCancel := context.WithCancel(ctx)

	testResourceWatcher := newResourcesWatcher(ctx)
	testResourceWatcherInternal, ok := testResourceWatcher.(*resWatcher)
	testutil.AssertTrue(t, ok)
	testutil.AssertNotNil(t, testResourceWatcherInternal.watchCache)
	testutil.AssertEqual(t, 0, len(testResourceWatcherInternal.watchCache))
	testutil.AssertEqual(t, expectedCtx, testResourceWatcherInternal.watcherCtx)
	testutil.AssertEqual(t, reflect.ValueOf(expectedCtxCancel).Pointer(), reflect.ValueOf(testResourceWatcherInternal.watcherCtxCancel).Pointer())
	testutil.AssertNotNil(t, testResourceWatcherInternal.watchCacheWaitGroup)
}

func TestResourceWatcherWatch(t *testing.T) {
	const (
		testResourceID      = "test-res-id"
		testExpiryDuration  = 5 * time.Hour
		testTimeoutDuration = 5 * time.Second
	)

	testCases := map[string]struct {
		watchCache         map[string]watchInfo
		expectedWatchCache map[string]watchInfo
		prepareTestCtx     func() (context.Context, context.CancelFunc)
		expectedError      error
	}{
		"test_ctx_cancelled": {
			watchCache:         map[string]watchInfo{},
			expectedWatchCache: map[string]watchInfo{},
			prepareTestCtx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, cancel
			},
			expectedError: context.Canceled,
		},
		"test_already_watched": {
			watchCache: map[string]watchInfo{
				testResourceID: {
					resourceID: testResourceID,
					timer:      time.NewTimer(testExpiryDuration),
				},
			},
			expectedWatchCache: map[string]watchInfo{
				testResourceID: {
					resourceID: testResourceID,
					timer:      time.NewTimer(testExpiryDuration),
				},
			},
			prepareTestCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			expectedError: errAlreadyWatched,
		},
		"test_added_to_watch_cache_successfully": {
			watchCache: map[string]watchInfo{},
			expectedWatchCache: map[string]watchInfo{
				testResourceID: {
					resourceID: testResourceID,
				},
			},
			prepareTestCtx: func() (context.Context, context.CancelFunc) {
				return context.WithCancel(context.Background())
			},
			expectedError: nil,
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			testResourceWatch := &resWatcher{
				watchCache:          testData.watchCache,
				watchCacheLock:      sync.RWMutex{},
				watchCacheWaitGroup: &sync.WaitGroup{},
			}
			testResourceWatch.watcherCtx, testResourceWatch.watcherCtxCancel = testData.prepareTestCtx()

			defer func() {
				testResourceWatch.watcherCtxCancel()
				testutil.AssertWithTimeout(t, testResourceWatch.watchCacheWaitGroup, testTimeoutDuration)
			}()

			actualErr := testResourceWatch.Watch(testResourceID, testExpiryDuration, nil)

			testutil.AssertError(t, testData.expectedError, actualErr)
			testutil.AssertEqual(t, len(testData.expectedWatchCache), len(testResourceWatch.watchCache))

			if len(testData.expectedWatchCache) > 0 {
				for resID, resInfo := range testData.expectedWatchCache {
					actualWatchInfo, ok := testResourceWatch.watchCache[resID]
					testutil.AssertTrue(t, ok)
					testutil.AssertNotNil(t, actualWatchInfo)
					testutil.AssertEqual(t, resInfo.resourceID, actualWatchInfo.resourceID)
					testutil.AssertNotNil(t, actualWatchInfo.timer)
				}
			}
		})
	}
}

func TestResourceWatcherWatchHandling(t *testing.T) {
	const (
		testResourceID        = "test-res-id"
		testsExecutionTimeout = 10 * time.Second
	)

	testCases := map[string]struct {
		doCancel          bool
		expiryDuration    time.Duration
		testExpiryHandler watchExpired
	}{
		"test_added_to_watch_cache_timer_signal_no_handling_error": {
			doCancel:       false,
			expiryDuration: 250 * time.Millisecond,
			testExpiryHandler: func(ctx context.Context, id string) error {
				testutil.AssertEqual(t, testResourceID, id)
				return nil
			},
		},
		"test_added_to_watch_cache_timer_signal_handling_error": {
			doCancel:       false,
			expiryDuration: 250 * time.Millisecond,
			testExpiryHandler: func(ctx context.Context, id string) error {
				testutil.AssertEqual(t, testResourceID, id)
				return log.NewError("test error")
			},
		},
		"test_added_to_watch_cache_timer_signal_handling_cancelled": {
			doCancel:          true,
			expiryDuration:    24 * time.Hour,
			testExpiryHandler: nil,
		},
	}

	for testName, testData := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)

			testResourceWatch := &resWatcher{
				watchCache:          make(map[string]watchInfo),
				watchCacheLock:      sync.RWMutex{},
				watchCacheWaitGroup: &sync.WaitGroup{},
			}
			testResourceWatch.watcherCtx, testResourceWatch.watcherCtxCancel = context.WithCancel(context.Background())

			defer func() {
				testResourceWatch.watcherCtxCancel()
				testutil.AssertWithTimeout(t, testResourceWatch.watchCacheWaitGroup, testsExecutionTimeout)
			}()

			actualErr := testResourceWatch.Watch(testResourceID, testData.expiryDuration, testData.testExpiryHandler)
			testutil.AssertNil(t, actualErr)
			if testData.doCancel {
				testResourceWatch.watcherCtxCancel()
			}

			testutil.AssertWithTimeout(t, testResourceWatch.watchCacheWaitGroup, testsExecutionTimeout)
			_, ok := testResourceWatch.watchCache[testResourceID]
			testutil.AssertFalse(t, ok)
		})
	}
}

func TestResourceWatcherDispose(t *testing.T) {
	const testsExecutionTimeout = 5 * time.Second

	testResources := []struct {
		resourceID     string
		expiryDuration time.Duration
	}{
		{
			resourceID:     "expired-res-id",
			expiryDuration: 1 * time.Millisecond,
		},
		{
			resourceID:     "active-res-id",
			expiryDuration: 24 * time.Hour,
		},
	}

	testResourceWatch := &resWatcher{
		watchCache:          make(map[string]watchInfo),
		watchCacheLock:      sync.RWMutex{},
		watchCacheWaitGroup: &sync.WaitGroup{},
	}
	testResourceWatch.watcherCtx, testResourceWatch.watcherCtxCancel = context.WithCancel(context.Background())

	for _, testData := range testResources {
		actualErr := testResourceWatch.Watch(testData.resourceID, testData.expiryDuration, nil)
		testutil.AssertNil(t, actualErr)
		_, isAdded := testResourceWatch.watchCache[testData.resourceID]
		testutil.AssertTrue(t, isAdded)
	}
	go testResourceWatch.Dispose()
	testutil.AssertWithTimeout(t, testResourceWatch.watchCacheWaitGroup, testsExecutionTimeout)

	testutil.AssertEqual(t, 0, len(testResourceWatch.watchCache))
}
