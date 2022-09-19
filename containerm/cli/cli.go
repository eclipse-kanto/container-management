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
	"fmt"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/client"
	"github.com/eclipse-kanto/container-management/containerm/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	kantocm3D = `


         ~~~~~~~~~~~~~
        ~~,   ~~~~~~ ~~
       ~~:   ~~~~~    ~~
      ~~:   ~:~       :~~      ,,,                            ,
     :~~   :       :::::~~     ,,,                           :,
     ~~         ::::::::::~    ,,,  ,,,  ,,,,,,,  ,,,,,,,,  ,,,,,,  ,,,,,,,
    ::       ::::::::::::::~   ,,, ,,,         ,, ,,,    ,,  ,,    ,,,   ,,,
    ::         ::::::::::::    ,,,,,      ,,,,,,, ,,,    ,,  ,,    ,,     ,,
     ::           ,,,,,,,,     ,,, ,,    ,,    ,, ,,,    ,,  ,,    ,,     ,,
      ::    ,       ,,,,,      ,,,  ,,   ,,   ,,, ,,,    ,,  ,,,   ,,,   ,,,
       ::    ,,,      ,,       ,,,   ,,  ,,,,, ,, ,,,    ,,   ,,,,  ,,,,,,,
        ,,    ,,,,,  ,,
         ,,,,,,,,,,,,,


Eclipse Kanto - Container Management
`
)

var (
	gwcmdLongFullHeader = fmt.Sprintf("%sCLI v%s, API v%s, (build %s %s) \n",
		kantocm3D,
		version.ProjectVersion,
		version.APIVersion,
		version.GitCommit,
		version.BuildTime)
)

type cli struct {
	config      config
	rootCmd     *cobra.Command
	gwManClient client.Client
}

func newCli() *cli {
	return &cli{
		rootCmd: &cobra.Command{
			Use:   "kanto-cm",
			Short: "Eclipse Kanto - Container Management CLI",
			Long:  gwcmdLongFullHeader,
			// disable displaying auto generation tag in cli docs
			DisableAutoGenTag: true,
		},
	}
}

func (c *cli) initLog() {
	if c.config.debug {
		logrus.SetLevel(logrus.DebugLevel)
		logrus.Infof("start client at debug level")
	}

	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}
	logrus.SetFormatter(formatter)
}

func (c *cli) initGwManClient() {
	gwClient, err := client.New(c.config.addressPath)
	if err != nil {
		logrus.Fatal(err)
	}
	c.gwManClient = gwClient
}

// add a subcommand
func (c *cli) addCommand(parent, child command) {
	child.init(c)

	parentCmd := parent.command()
	childCmd := child.command()

	// make command error not return command usage and error
	childCmd.SilenceUsage = true
	childCmd.SilenceErrors = true
	childCmd.DisableFlagsInUseLine = true

	childCmd.PreRun = func(cmd *cobra.Command, args []string) {
		c.initLog()
		c.initGwManClient()
	}

	parentCmd.AddCommand(childCmd)
}

// executes the client program
func (c *cli) run() error {
	return c.rootCmd.Execute()
}
