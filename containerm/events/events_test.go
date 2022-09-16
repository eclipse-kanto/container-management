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

package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
)

type eventsManagerPrep func() ContainerEventsManager

func TestPublishErr(t *testing.T) {
	var ctx context.Context
	pubErrTests := map[string]struct {
		eventsMgrInit eventsManagerPrep
		ctr           *types.Container
		err           error
	}{
		"test_nil_ctr_status": {
			ctr: &types.Container{
				ID:    "test-ctr",
				State: nil,
			},
			err: log.NewErrorf("container info missing - cannot publish event"),
		},
		"test_sink_closed": {
			eventsMgrInit: func() ContainerEventsManager {
				broadcaster := newEventSinksDispatcher()
				broadcaster.close()
				return &eventsMgr{broadcaster: broadcaster}
			},
			ctr: &types.Container{
				ID:      "test-ctr",
				Created: time.Now().UTC().Format(time.RFC3339),
				State: &types.State{
					Status: types.Created,
				},
			},
			err: errEventsSinkClosed,
		},
	}

	for testName, tc := range pubErrTests {
		t.Run(testName, func(t *testing.T) {
			t.Log(testName)
			var evMgr ContainerEventsManager
			if tc.eventsMgrInit != nil {
				evMgr = tc.eventsMgrInit()
			} else {
				evMgr = newEventsManager()
			}
			err := evMgr.Publish(ctx, types.EventTypeContainers, types.EventActionContainersCreated, tc.ctr)
			testutil.AssertError(t, tc.err, err)
		})
	}
}

func TestSubscribe(t *testing.T) {
	evMgr := newEventsManager()

	ctx := context.Background()

	subscribeCtx, subscribeCtxCancelFunc := context.WithCancel(ctx)
	t.Cleanup(subscribeCtxCancelFunc)
	eventsChan, eventsErrChan := evMgr.Subscribe(subscribeCtx)

	tcs := []struct {
		action    types.EventAction
		container *types.Container
	}{{
		action: types.EventActionContainersCreated,
		container: &types.Container{

			ID:      "test-ctr",
			Created: time.Now().UTC().Format(time.RFC3339),
			State: &types.State{
				Status: types.Created,
			},
		},
	},
		{
			action: types.EventActionContainersRunning,
			container: &types.Container{
				ID:      "test-ctr",
				Created: time.Now().UTC().Format(time.RFC3339),
				State: &types.State{
					Status: types.Running,
				},
			},
		},
	}

	t.Log("publish test events ")
	var wg sync.WaitGroup
	wg.Add(1)
	errChan := make(chan error)
	go func() {
		defer wg.Done()
		defer close(errChan)
		for _, tc := range tcs {
			if err := evMgr.Publish(ctx, types.EventTypeContainers, tc.action, tc.container); err != nil {
				errChan <- err
				return
			}
		}
		t.Log("finished publishing test events")
	}()

	t.Log("waiting to publish all test events")
	wg.Wait()
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
	var received []*types.Event
assertSubscribe:
	for {
		select {
		case msg := <-eventsChan:
			received = append(received, msg)
		case err := <-eventsErrChan:
			if err != nil {
				t.Errorf("unexpected error received: %v", err)
				t.Fatal(err)
			}
			break assertSubscribe
		}
		if len(received) == len(tcs) {
			subscribeCtxCancelFunc()
			for i, received := range received {
				testutil.AssertEqual(t, types.EventTypeContainers, received.Type)
				testutil.AssertEqual(t, tcs[i].action, received.Action)
				testutil.AssertEqual(t, tcs[i].container, &received.Source)
			}
		}
	}
}

func TestSubscribeContextError(t *testing.T) {
	evMgr := newEventsManager()
	subscribeCtx, subscribeCtxCancelFunc := context.WithDeadline(context.Background(), time.Now().UTC())
	t.Cleanup(subscribeCtxCancelFunc)
	eventsChan, eventsErrChan := evMgr.Subscribe(subscribeCtx)

assertSubscribe:
	for {
		select {
		case msg := <-eventsChan:
			if msg != nil {
				t.Errorf("unexpected message received: %v", msg)
				t.Fatal(msg)
			}
			break assertSubscribe
		case err := <-eventsErrChan:
			testutil.AssertEqual(t, context.DeadlineExceeded, err)
			break assertSubscribe
		}
	}
}

func TestSubscribeInvalidMessageError(t *testing.T) {
	broadcaster := newEventSinksDispatcher()
	evMgr := &eventsMgr{
		broadcaster: broadcaster,
	}

	subscribeCtx, subscribeCtxCancelFunc := context.WithCancel(context.Background())
	eventsChan, eventsErrChan := evMgr.Subscribe(subscribeCtx)

	expectedErr := log.NewError("test expected error")

	t.Log("publish test events ")
	var wg sync.WaitGroup
	wg.Add(1)
	errChan := make(chan error)
	go func() {
		defer wg.Done()
		defer close(errChan)
		if err := broadcaster.write(expectedErr); err != nil {
			errChan <- err
			return
		}
		t.Log("finished publishing test events")
	}()

	t.Log("waiting to publish all test events")
	wg.Wait()
	if err := <-errChan; err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(1 * time.Second)
		subscribeCtxCancelFunc()
	}()

assertSubscribe:
	for {
		select {
		case msg := <-eventsChan:
			if msg != nil {
				t.Errorf("unexpected message received: %v", msg)
				t.Fatal(msg)
			}
			break assertSubscribe
		case err := <-eventsErrChan:
			testutil.AssertError(t, log.NewErrorf("invalid message received: %#v", expectedErr), err)
			break assertSubscribe
		}
	}
}
