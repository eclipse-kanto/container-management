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
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	name              string
	terminal          bool
	interactive       bool
	privileged        bool
	network           string
	containerFile     string
	extraHosts        []string
	extraCapabilities []string
	devices           []string
	mountPoints       []string
	ports             []string
	env               []string
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
		Use:   "create [option]... [container-image-id] [command] [command-arg]...",
		Short: "Create a container.",
		Long:  "Create a container.",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return cc.run(args)
		},
		Example: "create container-image-id",
	}
	cc.cmd.Flags().SetInterspersed(false)
	cc.setupFlags()
}

func initContainer(config createConfig, imageName string) *types.Container {
	return &types.Container{
		Name: config.name,
		Image: types.Image{
			Name: imageName,
		},
		HostConfig: &types.HostConfig{
			Privileged:        config.privileged,
			ExtraHosts:        config.extraHosts,
			ExtraCapabilities: config.extraCapabilities,
			NetworkMode:       types.NetworkMode(config.network),
		},
		IOConfig: &types.IOConfig{
			Tty:       config.terminal,
			OpenStdin: config.interactive,
		},
	}
}

func (cc *createCmd) containerFromFile() (*types.Container, error) {
	var err error
	cc.cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Changed && flag.Name != "file" {
			err = log.NewError("no other flags are expected when creating a container from file")
		}
	})
	if err != nil {
		return nil, err
	}
	byteValue, err := os.ReadFile(cc.config.containerFile)
	if err != nil {
		return nil, err
	}
	ctrToCreate := initContainer(cc.config, "")
	if err = json.Unmarshal(byteValue, ctrToCreate); err != nil {
		return nil, err
	}
	return ctrToCreate, nil
}

func (cc *createCmd) containerFromFlags(args []string) (*types.Container, error) {
	ctrToCreate := initContainer(cc.config, args[0])

	var command []string
	if len(args) > 1 {
		command = args[1:]
	}

	if cc.config.privileged && cc.config.devices != nil {
		return nil, log.NewError("cannot create the container as privileged and with specified devices at the same time - choose one of the options")
	}

	if cc.config.privileged && cc.config.extraCapabilities != nil {
		return nil, log.NewError("cannot create the container as privileged and with extra capabilities at the same time - choose one of the options")
	}

	if cc.config.env != nil || command != nil {
		ctrToCreate.Config = &types.ContainerConfiguration{
			Env: cc.config.env,
			Cmd: command,
		}
	}

	if cc.config.devices != nil {
		devs, err := util.ParseDeviceMappings(cc.config.devices)
		if err != nil {
			return nil, err
		}
		ctrToCreate.HostConfig.Devices = devs
	}

	if cc.config.mountPoints != nil {
		mounts, err := util.ParseMountPoints(cc.config.mountPoints)
		if err != nil {
			return nil, err
		} else if mounts != nil {
			ctrToCreate.Mounts = mounts
		}
	}
	if cc.config.ports != nil {
		mappings, err := util.ParsePortMappings(cc.config.ports)
		if err != nil {
			return nil, err
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

	return ctrToCreate, nil
}

func (cc *createCmd) run(args []string) error {
	var (
		ctrToCreate *types.Container
		err         error
	)

	if len(cc.config.containerFile) > 0 {
		if len(args) > 0 {
			return log.NewError("no arguments are expected when creating a container from file")
		}
		if ctrToCreate, err = cc.containerFromFile(); err != nil {
			return err
		}
	} else if len(args) != 0 {
		if ctrToCreate, err = cc.containerFromFlags(args); err != nil {
			return err
		}
	} else {
		return log.NewError("container image argument is expected")
	}

	if err = util.ValidateContainer(ctrToCreate); err != nil {
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
			"bridge - the container is connected to the default bridge network interface of the engine and is assigned an IP (this is the default)\n"+
			"host - the container shares the network stack of the host (use with caution as this breaks the network's isolation!)")
	// init extra hosts
	flagSet.StringSliceVar(&cc.config.extraHosts, "hosts", nil, "Extra hosts to be added in the current container's /etc/hosts file. Example: \n"+
		"--hosts=\"hostname1:<IP1>, hostname2:<IP2>..\" \n"+
		"If the IP of the host machine is to be added to the container's hosts file the reserved host_ip[_<network-interface>] must be provided. Example:\n"+
		"--hosts=\"local.host.machine.ip.custom.if:host_ip_myNetIf0\" \n"+
		"this will automatically resolve the host's IP on the myNetIf0 network interface and add it to the container's hosts file \n"+
		"--hosts=\"local.host.machine.ip.default.bridge:host_ip\" \n"+
		"this will automatically resolve the host's IP on the default bridge network interface for containerm (the default configuration is kanto-cm0) and add it to the container's hosts file if the container is configured to use it\n"+
		"If the IP of a container in the same bridge network is to be added to the hosts file the reserved container_<container-host_name> must be provided. Example:\n"+
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
	//init extra capabilities
	flagSet.StringSliceVar(&cc.config.extraCapabilities, "cap-add", nil, "Add Linux capabilities to the container")
	flagSet.StringVarP(&cc.config.containerFile, "file", "f", "", "Creates a container with a predefined config given by the user.")
}
