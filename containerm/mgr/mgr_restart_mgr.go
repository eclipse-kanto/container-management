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

package mgr

import (
	"sync"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
)

const (
	backoffMultiplier        = 2
	defaultMaxRetryTimeout   = 100 * time.Millisecond
	defaultMaxRestartTimeout = 1 * time.Minute
)

var restartManagerCanceled = log.NewError("container's restart manager is canceled")

type restartManager struct {
	sync.Mutex
	sync.Once
	restartPolicy     *types.RestartPolicy
	restartsPerformed int
	timeout           time.Duration
	maxRestartTimeout time.Duration
	isActive          bool
	cancelChan        chan struct{}
	isCanceled        bool
}

// New returns a new restartManager based on a policy.
func newRestartManager(policy *types.RestartPolicy, restartCount int) *restartManager {
	mrt := defaultMaxRestartTimeout
	if policy.RetryTimeout != 0 {
		mrt = policy.RetryTimeout
	}
	return &restartManager{restartPolicy: policy, restartsPerformed: restartCount, maxRestartTimeout: mrt, cancelChan: make(chan struct{})}
}

func (rm *restartManager) shouldRestart(exitCode uint32, hasBeenManuallyStopped bool, executionDuration time.Duration) (bool, chan error, error) {
	if util.IsRestartPolicyNone(rm.restartPolicy) {
		return false, nil, nil
	}
	rm.Lock()
	unlockOnExit := true
	defer func() {
		if unlockOnExit {
			rm.Unlock()
		}
	}()

	if rm.isCanceled {
		return false, nil, restartManagerCanceled
	}

	if rm.isActive {
		return false, nil, log.NewErrorf("invalid call on an active restart manager")
	}
	// if the container ran for more than 10s, regardless of status and policy
	//reset the the timeout back to the default
	if executionDuration.Seconds() >= 10 {
		rm.timeout = 0
	}
	switch {
	case rm.timeout == 0:
		rm.timeout = defaultMaxRetryTimeout
	case rm.timeout < rm.maxRestartTimeout:
		rm.timeout *= backoffMultiplier
	default:
		log.Debug("no restart manager timeout adjustments needed")
	}
	if rm.timeout > rm.maxRestartTimeout {
		rm.timeout = rm.maxRestartTimeout
	}

	var restart bool
	switch {
	case util.IsRestartPolicyAlways(rm.restartPolicy):
		restart = true
	case util.IsRestartPolicyUnlessStopped(rm.restartPolicy) && !hasBeenManuallyStopped:
		restart = true
	case util.IsRestartPolicyOnFailure(rm.restartPolicy):
		// the default value of 0 for MaximumRetryCount means that we will not enforce a maximum count
		log.Debug("restart manager retry count is %d, policy's max retry count is %d", rm.restartsPerformed, rm.restartPolicy.MaximumRetryCount)
		if max := rm.restartPolicy.MaximumRetryCount; max == 0 || rm.restartsPerformed < max {
			restart = exitCode != 0
			log.Debug("checking exit code for policy onFailure (exitCode != 0): %v", restart)
		}
	default:
		log.Debug("no restart attempts are required and will be made for restart policy %s , hasBeenManuallyStopped = %v", rm.restartPolicy.Type, hasBeenManuallyStopped)
	}

	if !restart {
		rm.isActive = false
		return false, nil, nil
	}

	rm.restartsPerformed++
	log.Debug("incremented restart manager retry count to %d", rm.restartsPerformed)

	unlockOnExit = false
	rm.isActive = true
	rm.Unlock()

	ch := make(chan error)
	go func() {
		select {
		case <-rm.cancelChan:
			ch <- restartManagerCanceled
			close(ch)
		case <-time.After(rm.timeout):
			rm.Lock()
			close(ch)
			rm.isActive = false
			rm.Unlock()
		}
	}()

	return true, ch, nil
}

func (rm *restartManager) cancel() error {
	rm.Do(func() {
		rm.Lock()
		rm.isCanceled = true
		close(rm.cancelChan)
		rm.Unlock()
	})
	return nil
}
