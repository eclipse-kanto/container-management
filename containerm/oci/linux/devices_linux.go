// Copyright The PouchContainer Authors.
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

// Package name changed, Bosch.IO GmbH, 2020

package linux

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/runc/libcontainer/configs"
	"github.com/opencontainers/runc/libcontainer/devices"
	"github.com/opencontainers/runtime-spec/specs-go"
)

func iPtr(i int64) *int64        { return &i }
func u32Ptr(i int64) *uint32     { u := uint32(i); return &u }
func fmPtr(i int64) *os.FileMode { fm := os.FileMode(i); return &fm }

// Device transforms a libcontainer configs.Device to a specs.LinuxDevice object.
func Device(d *configs.Device) specs.LinuxDevice {
	return specs.LinuxDevice{
		Type:     string(d.Type),
		Path:     d.Path,
		Major:    d.Major,
		Minor:    d.Minor,
		FileMode: fmPtr(int64(d.FileMode)),
		UID:      u32Ptr(int64(d.Uid)),
		GID:      u32Ptr(int64(d.Gid)),
	}
}

func deviceCgroup(d *configs.Device) specs.LinuxDeviceCgroup {
	t := string(d.Type)
	return specs.LinuxDeviceCgroup{
		Allow:  true,
		Type:   t,
		Major:  &d.Major,
		Minor:  &d.Minor,
		Access: d.Permissions,
	}
}

// DevicesFromPath computes a list of devices and device permissions from paths (pathOnHost and pathInContainer) and cgroup permissions.
func DevicesFromPath(pathOnHost, pathInContainer, cgroupPermissions string) (devs []specs.LinuxDevice, devPermissions []specs.LinuxDeviceCgroup, err error) {
	resolvedPathOnHost := pathOnHost

	// check if it is a symbolic link
	if src, e := os.Lstat(pathOnHost); e == nil && src.Mode()&os.ModeSymlink == os.ModeSymlink {
		if linkedPathOnHost, e := filepath.EvalSymlinks(pathOnHost); e == nil {
			resolvedPathOnHost = linkedPathOnHost
		}
	}

	device, err := devices.DeviceFromPath(resolvedPathOnHost, cgroupPermissions)
	// if there was no error, return the device
	if err == nil {
		device.Path = pathInContainer
		return append(devs, Device(device)), append(devPermissions, deviceCgroup(device)), nil
	}

	// if the device is not a device node
	// try to see if it's a directory holding many devices
	if err == devices.ErrNotADevice {

		// check if it is a directory
		if src, e := os.Stat(resolvedPathOnHost); e == nil && src.IsDir() {

			// mount the internal devices recursively
			filepath.Walk(resolvedPathOnHost, func(dpath string, f os.FileInfo, e error) error {
				childDevice, e := devices.DeviceFromPath(dpath, cgroupPermissions)
				if e != nil {
					// ignore the device
					return nil
				}

				// add the device to userSpecified devices
				childDevice.Path = strings.Replace(dpath, resolvedPathOnHost, pathInContainer, 1)
				devs = append(devs, Device(childDevice))
				devPermissions = append(devPermissions, deviceCgroup(childDevice))

				return nil
			})
		}
	}

	if len(devs) > 0 {
		return devs, devPermissions, nil
	}

	return devs, devPermissions, fmt.Errorf("error gathering device information while adding custom device %q: %s", pathOnHost, err)
}
