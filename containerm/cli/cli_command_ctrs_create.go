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

package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/spf13/cobra"
)

type createCmd struct {
	baseCommand
	config createConfig
}

type restartPolicy struct {
	kind          string
	timeout       int64
	maxRetryCount int
}

type resources struct {
	memory            string
	memoryReservation string
	memorySwap        string
}

type createConfig struct {
	name        string
	terminal    bool
	interactive bool
	privileged  bool
	network     string
	extraHosts  []string
	devices     []string
	mountPoints []string
	ports       []string
	env         []string
	// log configs
	logDriver        string
	logMaxFiles      int
	logMaxSize       string
	logRootDirPath   string
	logMode          string
	logMaxBufferSize string
	decKeys          []string
	decRecipients    []string
	restartPolicy
	resources
}

func (cc *createCmd) init(cli *cli) {
	cc.cli = cli
	cc.cmd = &cobra.Command{
		Use:   "create [option]... container-image-id [command] [command-arg]...",
		Short: "Create a container.",
		Long:  "Create a container.",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: "create container-image-id",
	}
	cc.cmd.Flags().SetInterspersed(false)
	cc.setupFlags()
}

func (cc *createCmd) run(args []string) error {
	// parse parameters
	imageID := args[0]
	var command []string
	if len(args) > 1 {
		command = args[1:]
	}

	if cc.config.privileged && cc.config.devices != nil {
		return log.NewError("cannot create the container as privileged and with specified devices at the same time - choose one of the options")
	}

	ctrToCreate := &types.Container{
		Name: cc.config.name,
		Image: types.Image{
			Name: imageID,
		},
		HostConfig: &types.HostConfig{
			Privileged:  cc.config.privileged,
			ExtraHosts:  cc.config.extraHosts,
			NetworkMode: types.NetworkMode(cc.config.network),
		},
		IOConfig: &types.IOConfig{
			Tty:       cc.config.terminal,
			OpenStdin: cc.config.interactive,
		},
	}

	if cc.config.env != nil || command != nil {
		ctrToCreate.Config = &types.ContainerConfiguration{
			Env: cc.config.env,
			Cmd: command,
		}
	}

	if cc.config.devices != nil {
		devs, err := parseDevices(cc.config.devices)
		if err != nil {
			return err
		}
		ctrToCreate.HostConfig.Devices = devs
	}

	if cc.config.mountPoints != nil {
		mounts, err := parseMountPoints(cc.config.mountPoints)
		if err != nil {
			return err
		} else if mounts != nil {
			ctrToCreate.Mounts = mounts
		}
	}
	if cc.config.ports != nil {
		mappings, err := parsePortMappings(cc.config.ports)
		if err != nil {
			return err
		}
		ctrToCreate.HostConfig.PortMappings = mappings
	}

	switch cc.config.restartPolicy.kind {
	case string(types.Always):
		ctrToCreate.HostConfig.RestartPolicy = &types.RestartPolicy{
			Type: types.Always,
		}
	case string(types.No):
		ctrToCreate.HostConfig.RestartPolicy = &types.RestartPolicy{
			Type: types.No,
		}
	case string(types.UnlessStopped):

		ctrToCreate.HostConfig.RestartPolicy = &types.RestartPolicy{
			Type: types.UnlessStopped,
		}
	case string(types.OnFailure):
		ctrToCreate.HostConfig.RestartPolicy = &types.RestartPolicy{
			Type:              types.OnFailure,
			MaximumRetryCount: cc.config.restartPolicy.maxRetryCount,
			RetryTimeout:      time.Duration(cc.config.restartPolicy.timeout) * time.Second,
		}
	default:
		ctrToCreate.HostConfig.RestartPolicy = nil
	}

	ctrToCreate.HostConfig.LogConfig = &types.LogConfiguration{}
	switch cc.config.logDriver {
	case string(types.LogConfigDriverJSONFile):
		ctrToCreate.HostConfig.LogConfig.DriverConfig = &types.LogDriverConfiguration{
			Type:     types.LogConfigDriverJSONFile,
			MaxFiles: cc.config.logMaxFiles,
			MaxSize:  cc.config.logMaxSize,
			RootDir:  cc.config.logRootDirPath,
		}
	case string(types.LogConfigDriverNone):
		ctrToCreate.HostConfig.LogConfig.DriverConfig = &types.LogDriverConfiguration{
			Type: types.LogConfigDriverNone,
		}
	default:
		ctrToCreate.HostConfig.LogConfig.DriverConfig = nil
	}

	switch cc.config.logMode {
	case string(types.LogModeBlocking):
		ctrToCreate.HostConfig.LogConfig.ModeConfig = &types.LogModeConfiguration{
			Mode: types.LogModeBlocking,
		}
	case string(types.LogModeNonBlocking):
		ctrToCreate.HostConfig.LogConfig.ModeConfig = &types.LogModeConfiguration{
			Mode:          types.LogModeNonBlocking,
			MaxBufferSize: cc.config.logMaxBufferSize,
		}
	default:
		ctrToCreate.HostConfig.LogConfig.ModeConfig = nil
	}

	ctrToCreate.HostConfig.Resources = getResourceLimits(cc.config.resources)
	ctrToCreate.Image.DecryptConfig = getDecryptConfig(cc.config)

	if err := util.ValidateContainer(ctrToCreate); err != nil {
		return err
	}

	ctr, err := cc.cli.gwManClient.Create(context.Background(), ctrToCreate)
	if ctr != nil {
		fmt.Println(ctr.ID)
	}
	return err
}

func getDecryptConfig(config createConfig) *types.DecryptConfig {
	if len(config.decKeys) == 0 && len(config.decRecipients) == 0 {
		return nil
	}
	return &types.DecryptConfig{
		Keys:       config.decKeys,
		Recipients: config.decRecipients,
	}
}

func getResourceLimits(r resources) *types.Resources {
	if r.memory != "" || r.memoryReservation != "" || r.memorySwap != "" {
		return &types.Resources{
			Memory:            r.memory,
			MemoryReservation: r.memoryReservation,
			MemorySwap:        r.memorySwap,
		}
	}
	return nil
}

func (cc *createCmd) setupFlags() {
	flagSet := cc.cmd.Flags()
	// init name flags
	flagSet.StringVarP(&cc.config.name, "name", "n", "", "Create a container with a specific name. A valid name must start with an uppercase or a lowercase letter, a digit or an underscore and not exceed 32 symbols. It can also contain a dot and a hyphen.")
	// init terminal flags
	flagSet.BoolVar(&cc.config.terminal, "t", false, "Enable terminal for the current container")
	// init interactive flags
	flagSet.BoolVar(&cc.config.interactive, "i", false, "Enable interaction with the current container")
	// init interactive flags
	flagSet.BoolVar(&cc.config.privileged, "privileged", false, "Create the container as privileged")
	// init restart policy flags
	flagSet.StringVar(&cc.config.restartPolicy.kind, "rp", "",
		"Sets the restart policy for the container.Supported restart policies are - no, always, unless-stopped (the default), always. \n"+
			"no - no attempts to restart the container for any reason will be made \n"+
			"always - an attempt to restart the container will be me made each time the container exits regardless of the exit code \n"+
			"unless-stopped - restart attempts will be made only if the container has not been stopped by the user \n"+
			"on-failure - restart attempts will be made if the container exits with an exit code != 0; \n"+
			"the additional flags (--rp-cnt and --rp-to) apply only for this policy; if max retry count if not provided - the system will retry until it succeeds endlessly \n")
	// init  restart policy max retry count flags
	flagSet.IntVar(&cc.config.restartPolicy.maxRetryCount, "rp-cnt", 1, "Sets the number of retries that will be made to restart the container on exit if the policy is set to Always")
	// init  restart policy max retry count flags
	flagSet.Int64Var(&cc.config.restartPolicy.timeout, "rp-to", 30, "Sets the time out period in seconds for each retry that will be made to restart the container on exit if the policy is set to Always")
	// init devices
	flagSet.StringSliceVar(&cc.config.devices, "devices", nil, "Devices to be made available in the current container and optional cgroups permissions configuration. Both path on host and in container must be set. Possible cgroup permissions options are \"r\" (read), \"w\" (write), \"m\" (mknod) and all combinations of the three are possible. If not set, \"rwm\" is default device configuration. Example: \n"+
		"--devices=/dev/ttyACM0:/dev/ttyUSB0[:rwm]")
	// init ports
	flagSet.StringSliceVar(&cc.config.ports, "ports", nil, "Ports to be mapped from the host to the container instance. Template: \n"+
		"--ports=[<host-ip>:]<host-port>:<container-port>[-<range>][/<proto>] \n"+
		"Most common use-case: \n"+
		"--ports=80:80\n"+
		"Mapping the container's 80 port to a host port in the 5000-6000 range: \n"+
		"--ports=5000-6000:80/udp\n"+
		"Specifying port protocol (default is tcp): \n"+
		"--ports=80:80/udp\n"+
		"By default the port mappings will set on all network interfaces, but this is also manageable. Example with two mappings including an optional host port range and udp: \n"+
		"--ports=0.0.0.0:80-100:80/udp",
	)
	// init network mode
	flagSet.StringVar(&cc.config.network, "network", string(types.NetworkModeBridge),
		"Sets the networking mode for the container. Possible options are:\n"+
			"bridge - the container is connected to the default bridge network interface of th engine and is assigned an IP (this is the default)\n"+
			"host - the container shares the network stack of the host (use with caution as this breaks the network's isolation!)")
	// init extra hosts
	flagSet.StringSliceVar(&cc.config.extraHosts, "hosts", nil, "Extra hosts to be added in the current container's /etc/hosts file. Example: \n"+
		"--hosts=\"hostname1:<IP1>, hostname2:<IP2>..\" \n"+
		"If the IP of the host machine is to be added to the container's hosts file the reserved host_ip[_<network-interface>] must be provided. Example:\n"+
		"--hosts=\"local.host.machine.ip.custom.if:host_ip_myNetIf0\" \n"+
		"this will automatically resolve the host's IP on the myNetIf0 network interface and add it to the container's hosts file \n"+
		"--hosts=\"local.host.machine.ip.default.bridge:host_ip\" \n"+
		"this will automatically resolve the host's IP on the default bridge network interface for containerm (the default configuration is kanto-cm0) and add it to the container's hosts file if the container is configured to use it\n"+
		"If the IP of a container in the same bridge network is to be added to the contains hosts file the reserved container[_<container-host_name>] must be provided. Example:\n"+
		"--hosts=\"service:container_service-host\"")
	flagSet.StringSliceVar(&cc.config.mountPoints, "mp", nil, "Sets mount points so a source directory on the host can be accessed via a destination directory in the container. Example:\n"+
		"--mp=\"source1:destination1:propagation_mode, source2:destination2\" \n"+
		"If the propagation mode parameter is omitted, 'rprivate' will be set by default.  \n"+
		"Available propagation modes are: rprivate, private, rshared, shared, rslave, slave")
	flagSet.StringArrayVar(&cc.config.env, "e", nil, "Sets the provided environment variables in the root container's process environment. Example:\n"+
		"--e=VAR1=2 --e=VAR2=\"a bc\"\n"+
		"If --e=VAR1= is used, the environment variable would be set to empty.\n"+
		"If --e=VAR1 is used, the environment variable would be removed from the container environment inherited from the image.")
	flagSet.StringVar(&cc.config.logDriver, "log-driver", string(types.LogConfigDriverJSONFile), "Sets the type of the log driver to be used for the container - json-file (default), none")
	flagSet.IntVar(&cc.config.logMaxFiles, "log-max-files", 2, "Sets the max number of log files to be rotated - applicable for json-file log driver only")
	flagSet.StringVar(&cc.config.logMaxSize, "log-max-size", "100M", "Sets the max size of the logs files for rotation in the form of 1, 1.2m,1g, etc. - applicable for json-file log driver only")
	flagSet.StringVar(&cc.config.logRootDirPath, "log-path", "", "Sets the path to the directory where the log files will be stored - applicable for json-file log driver only")
	flagSet.StringVar(&cc.config.logMode, "log-mode", string(types.LogModeBlocking), "Sets the mode of the logger - blocking (default), non-blocking")
	flagSet.StringVar(&cc.config.logMaxBufferSize, "log-max-buffer-size", "1M", "Sets the max size of the logger buffer in the form of 1, 1.2m - applicable for non-blocking mode only")
	flagSet.StringVarP(&cc.config.resources.memory, "memory", "m", "", "Sets the max amount of memory the container can use in the form of 200m, 1.2g. The minimum allowed value is 3m\n"+
		"By default, a container has no memory constraints.")
	flagSet.StringVar(&cc.config.resources.memoryReservation, "memory-reservation", "", "Sets a soft memory limitation in the form of 200m, 1.2g. Must be smaller than --memory.\n"+
		"When the system detects memory contention or low memory, control groups are pushed back to their soft limits.\n"+
		"There is no guarantee that the container memory usage will not exceed the soft limit.")
	flagSet.StringVar(&cc.config.resources.memorySwap, "memory-swap", "", "Sets the total amount of memory + swap that the container can use in the form of 200m, 1.2g.\n"+
		"If set must not be smaller than --memory. If equal to --memory, than the container will not have access to swap.\n"+
		"If not set and --memory is set, than the container can use as much swap as the --memory setting.\n"+
		"If set to -1, the container can use unlimited swap, up to the amount available on the host.")
	flagSet.StringSliceVar(&cc.config.decKeys, "dec-keys", nil, "Sets a list of private keys filenames (GPG private key ring, JWE and PKCS7 private key). Each entry can include an optional password separated by a colon after the filename.")
	flagSet.StringSliceVar(&cc.config.decRecipients, "dec-recipients", nil, "Sets a recipients certificates list of the image (used only for PKCS7 and must be an x509)")
}

func parseDevices(devices []string) ([]types.DeviceMapping, error) {
	var devs []types.DeviceMapping
	for _, devPair := range devices {
		pair := strings.Split(strings.TrimSpace(devPair), ":")
		if len(pair) == 2 {
			devs = append(devs, types.DeviceMapping{
				PathOnHost:        pair[0],
				PathInContainer:   pair[1],
				CgroupPermissions: "rwm",
			})
		} else if len(pair) == 3 {
			if len(pair[2]) == 0 || len(pair[2]) > 3 {
				return nil, log.NewError("incorrect device cgroup permissions format")
			}
			for i := 0; i < len(pair[2]); i++ {
				if (pair[2])[i] != "w"[0] && (pair[2])[i] != "r"[0] && (pair[2])[i] != "m"[0] {
					return nil, log.NewError("incorrect device cgroup permissions format")
				}
			}

			devs = append(devs, types.DeviceMapping{
				PathOnHost:        pair[0],
				PathInContainer:   pair[1],
				CgroupPermissions: pair[2],
			})
		} else {
			return nil, log.NewError("incorrect device configuration format")
		}
	}
	return devs, nil
}

func parseMountPoints(mps []string) ([]types.MountPoint, error) {
	var mountPoints []types.MountPoint
	var mountPoint types.MountPoint
	for _, mp := range mps {
		mount := strings.Split(strings.TrimSpace(mp), ":")
		// if propagation mode is omitted, "rprivate" is set as default
		if len(mount) < 2 || len(mount) > 3 {
			return nil, log.NewError("Incorrect number of parameters of the mount point")
		}
		mountPoint = types.MountPoint{
			Destination: mount[1],
			Source:      mount[0],
		}
		if len(mount) == 2 {
			log.Debug("propagation mode ommited - setting default to rprivate")
			mountPoint.PropagationMode = types.RPrivatePropagationMode
		} else {
			mountPoint.PropagationMode = mount[2]
		}
		mountPoints = append(mountPoints, mountPoint)
	}
	return mountPoints, nil
}

func parsePortMappings(mappings []string) ([]types.PortMapping, error) {
	var (
		portMappings  []types.PortMapping
		err           error
		protocol      string
		containerPort int64
		hostIP        string
		hostPort      int64
		hostPortEnd   int64
	)

	for _, mapping := range mappings {
		mappingWithProto := strings.Split(strings.TrimSpace(mapping), "/")
		mapping = mappingWithProto[0]
		if len(mappingWithProto) == 2 {
			// port is specified, e.g.80:80/tcp
			protocol = mappingWithProto[1]
		}
		addressAndPorts := strings.Split(strings.TrimSpace(mapping), ":")
		hostPortIdx := 0 // if host ip not set
		if len(addressAndPorts) == 2 {
			// host address not specified, e.g. 80:80
		} else if len(addressAndPorts) == 3 {
			hostPortIdx = 1
			hostIP = addressAndPorts[0]
			validIP := net.ParseIP(hostIP)
			if validIP == nil {
				return nil, log.NewError("Incorrect host ip port mapping configuration")
			}
			hostPort, err = strconv.ParseInt(addressAndPorts[1], 10, 32)
			containerPort, err = strconv.ParseInt(addressAndPorts[2], 10, 32)

		} else {
			return nil, log.NewError("Incorrect port mapping configuration")
		}

		hostPortWithRange := strings.Split(strings.TrimSpace(addressAndPorts[hostPortIdx]), "-")
		if len(hostPortWithRange) == 2 {
			hostPortEnd, err = strconv.ParseInt(hostPortWithRange[1], 10, 32)
			if err != nil {
				return nil, log.NewError("Incorrect host range port mapping configuration")
			}
			hostPort, err = strconv.ParseInt(hostPortWithRange[0], 10, 32)
		} else {
			hostPort, err = strconv.ParseInt(addressAndPorts[hostPortIdx], 10, 32)
		}
		containerPort, err = strconv.ParseInt(addressAndPorts[hostPortIdx+1], 10, 32)
		if err != nil {
			return nil, log.NewError("Incorrect port mapping configuration, parsing error")
		}

		portMappings = append(portMappings, types.PortMapping{
			Proto:         protocol,
			ContainerPort: uint16(containerPort),
			HostIP:        hostIP,
			HostPort:      uint16(hostPort),
			HostPortEnd:   uint16(hostPortEnd),
		})
	}
	return portMappings, nil

}
