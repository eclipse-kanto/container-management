// Copyright 2013-2018 Docker, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package name and imports changed, Bosch.IO GmbH, 2020

package oci

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/opencontainers/runtime-spec/specs-go"
)

// nolint: gosimple
var deviceCgroupRuleRegex = regexp.MustCompile("^([acb]) ([0-9]+|\\*):([0-9]+|\\*) ([rwm]{1,3})$")

// SetCapabilities sets the provided capabilities on the spec
// All capabilities are added if privileged is true
func SetCapabilities(s *specs.Spec, caplist []string) error {
	s.Process.Capabilities.Effective = caplist
	s.Process.Capabilities.Bounding = caplist
	s.Process.Capabilities.Permitted = caplist
	s.Process.Capabilities.Inheritable = caplist
	// setUser has already been executed here
	// if non root drop capabilities in the way execve does
	if s.Process.User.UID != 0 {
		s.Process.Capabilities.Effective = []string{}
		s.Process.Capabilities.Permitted = []string{}
	}
	return nil
}

// AppendDevicePermissionsFromCgroupRules takes rules for the devices cgroup to append to the default set
func AppendDevicePermissionsFromCgroupRules(devPermissions []specs.LinuxDeviceCgroup, rules []string) ([]specs.LinuxDeviceCgroup, error) {
	for _, deviceCgroupRule := range rules {
		ss := deviceCgroupRuleRegex.FindAllStringSubmatch(deviceCgroupRule, -1)
		if len(ss[0]) != 5 {
			return nil, fmt.Errorf("invalid device cgroup rule format: '%s'", deviceCgroupRule)
		}
		matches := ss[0]

		dPermissions := specs.LinuxDeviceCgroup{
			Allow:  true,
			Type:   matches[1],
			Access: matches[4],
		}
		if matches[2] == "*" {
			major := int64(-1)
			dPermissions.Major = &major
		} else {
			major, err := strconv.ParseInt(matches[2], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid major value in device cgroup rule format: '%s'", deviceCgroupRule)
			}
			dPermissions.Major = &major
		}
		if matches[3] == "*" {
			minor := int64(-1)
			dPermissions.Minor = &minor
		} else {
			minor, err := strconv.ParseInt(matches[3], 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid minor value in device cgroup rule format: '%s'", deviceCgroupRule)
			}
			dPermissions.Minor = &minor
		}
		devPermissions = append(devPermissions, dPermissions)
	}
	return devPermissions, nil
}
