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

package util

import (
	"os"
	"sync"
)

var (
	runningSystemd bool
	detectSystemd  sync.Once
)

// Copyright 2015 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package util contains utility functions related to systemd that applications
// can use to check things like whether systemd is running.  Note that some of
// these functions attempt to manually load systemd libraries at runtime rather
// than linking against them.

// IsRunningSystemd checks whether the host was booted with systemd as its init
// system. This functions similarly to systemd's `sd_booted(3)`: internally, it
// checks whether /run/systemd/system/ exists and is a directory.
// http://www.freedesktop.org/software/systemd/man/sd_booted.html
//
// Package name changed also removed not needed logic and added custom code to handle the specific use case, Eclipse Kanto contributors, 2022
func IsRunningSystemd() bool {
	detectSystemd.Do(func() {
		fi, err := os.Lstat("/run/systemd/system")
		if err != nil {
			return
		}
		runningSystemd = fi.IsDir()
	})
	return runningSystemd
}
