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

package main

import (
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/docker/docker/pkg/reexec"
	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/spf13/cobra"
)

var (
	cfg = getDefaultInstance()
)

var rootCmd = &cobra.Command{
	Use:               "container-management",
	Short:             "The Eclipse Kanto - Container Management",
	Args:              cobra.NoArgs,
	SilenceUsage:      true,
	DisableAutoGenTag: true, // disable displaying auto generation tag in cli docs
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDaemon(cmd)

	},
}

func main() {
	if reexec.Init() {
		return
	}

	cfgFilePath := parseConfigFilePath()
	if cfgFilePath != "" {
		if err := loadLocalConfig(cfgFilePath, cfg); err != nil {
			log.ErrorErr(err, "failed to load local configuration provided - will exit")
			os.Exit(1)
		}
	} else {
		log.Debug("no external configuration is set for the daemon")
	}

	setupCommandFlags(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.ErrorErr(err, "failed to execute root command - will exit")
		os.Exit(1)
	}
}

func runDaemon(cmd *cobra.Command) error {
	initLogger(cfg)
	dumpConfiguration(cfg)

	gwDaemon, err := newDaemon(cfg)
	if err != nil {
		log.ErrorErr(err, "failed to create Kanto CM daemon instance")
		return err
	}

	gwDaemon.init()
	sockDir, _ := path.Split(gwDaemon.config.GrpcServerConfig.GrpcServerAddressPath)
	runLockFile := path.Join(sockDir, string(os.PathSeparator), lockFileName)
	l, lockErr := newRunLock(runLockFile)
	if lockErr == nil {
		err = l.TryLock()
		if err == nil {
			defer l.Unlock()

			os.MkdirAll(sockDir, 0755)
			os.Remove(gwDaemon.config.GrpcServerConfig.GrpcServerAddressPath)

			err := gwDaemon.start()
			if err != nil {
				log.ErrorErr(err, "failed to start Kanto CM daemon instance")
				return err
			}
			log.Debug("successfully started Kanto CM daemon instance")

			var signalChan = make(chan os.Signal, 1)
			signal.Notify(signalChan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL)
			select {
			case sig := <-signalChan:
				log.Debug("Received OS SIGNAL >> %d ! Will exit!", sig)
				gwDaemon.stop()
			}
		} else {
			log.ErrorErr(err, "another instance of container-management is already running at %s", gwDaemon.config.GrpcServerConfig.GrpcServerAddressPath)
			return err
		}
	} else {
		log.ErrorErr(lockErr, "unable to create lock file at %s", runLockFile)
		return lockErr
	}
	return nil
}
