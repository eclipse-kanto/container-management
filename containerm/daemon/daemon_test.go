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
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/eclipse-kanto/container-management/containerm/log"
	"github.com/eclipse-kanto/container-management/containerm/pkg/testutil"
	"github.com/eclipse-kanto/container-management/containerm/things"
	"github.com/spf13/cobra"
)

func TestLoadLocalConfig(t *testing.T) {
	cfg := getDefaultInstance()
	t.Run("test_not_existing", func(t *testing.T) {
		err := loadLocalConfig("../pkg/testutil/config/not-existing.json", cfg)
		if err != nil {
			t.Errorf("null error returned expected for non existing file")
		}
	})
	t.Run("test_is_dir", func(t *testing.T) {
		err := loadLocalConfig("../pkg/testutil/config/", cfg)
		testutil.AssertError(t, log.NewErrorf("provided configuration path %s is a directory", "../pkg/testutil/config/"), err)
	})
	t.Run("test_file_empty", func(t *testing.T) {
		err := loadLocalConfig("../pkg/testutil/config/empty.json", cfg)
		if err != nil {
			t.Errorf("no error expected, only warning")
		}
	})
	t.Run("test_json_invalid", func(t *testing.T) {
		err := loadLocalConfig("../pkg/testutil/config/invalid.json", cfg)
		testutil.AssertError(t, log.NewError("unexpected end of JSON input"), err)
	})
}

func TestParseConfigFilePath(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	testPath := "/some/path/file"

	t.Run("test_cfg_file_overridden", func(t *testing.T) {
		actualPath := parseConfigFilePath()
		if actualPath != daemonConfigFileDefault {
			t.Error("config file not set to default")
		}
	})
	t.Run("test_cfg_file_default", func(t *testing.T) {
		os.Args = []string{oldArgs[0], fmt.Sprintf("--%s=%s", daemonConfigFileFlagID, testPath)}
		actualPath := parseConfigFilePath()
		if actualPath != testPath {
			t.Error("config file not overridden by environment variable ")
		}
	})
}

// The following test is intended to serve as a check if any of the default configs has changed within a change.
// The test configuration json must be edited in this case - if the change really must be made.
func TestDefaultConfig(t *testing.T) {
	newDefaultConfig := getDefaultInstance()
	defaultConfig := &config{}
	_ = loadLocalConfig("../pkg/testutil/config/daemon-config.json", defaultConfig)
	if !reflect.DeepEqual(newDefaultConfig, defaultConfig) {
		t.Errorf("default configuration changed: %+v\ngot:%+v", defaultConfig, newDefaultConfig)
	}
}

func TestThingsServiceFeaturesConfig(t *testing.T) {
	local := &config{}
	_ = loadLocalConfig("../pkg/testutil/config/daemon-things-features-config.json", local)
	testutil.AssertEqual(t, []string{things.ContainerFactoryFeatureID}, local.ThingsConfig.Features)
}

func TestThingsTLSConfig(t *testing.T) {
	local := &config{}
	_ = loadLocalConfig("../pkg/testutil/config/daemon-things-tls-config.json", local)
	testutil.AssertEqual(t, &tlsConfig{RootCA: "ca.crt", ClientCert: "client.crt", ClientKey: "client.key"}, local.ThingsConfig.ThingsConnectionConfig.Transport)
}

func TestMgrDefaultCtrsStopTimeoutConfig(t *testing.T) {
	local := &config{}
	_ = loadLocalConfig("../pkg/testutil/config/daemon-mgr-default-ctrs-stop-timeout-config.json", local)
	testutil.AssertEqual(t, local.ManagerConfig.MgrDefaultCtrsStopTimeout, "15")
}

func TestExtractOpts(t *testing.T) {
	t.Run("test_extract_ctr_client_opts", func(t *testing.T) {
		opts := extractCtrClientConfigOptions(cfg)
		if len(opts) == 0 {
			t.Error("no ctr client opts after extraction")
		}
	})
	t.Run("test_extract_net_mgr_opts", func(t *testing.T) {
		opts := extractNetManagerConfigOptions(cfg)
		if len(opts) == 0 {
			t.Error("no net mgr opts after extraction")
		}
	})
	t.Run("test_extract_ctr_mgr_opts", func(t *testing.T) {
		opts := extractContainerManagerOptions(cfg)
		if len(opts) == 0 {
			t.Error("no ctr mgr opts after extraction")
		}
	})
	t.Run("test_extract_ctr_mgr_opts_with_not_suffixed_stop_timeout", func(t *testing.T) {
		config := &config{ManagerConfig: &managerConfig{MgrDefaultCtrsStopTimeout: "10"}}
		opts := extractContainerManagerOptions(config)
		if len(opts) == 0 {
			t.Error("no ctr mgr opts after extraction")
		}
		testutil.AssertEqual(t, "10s", config.ManagerConfig.MgrDefaultCtrsStopTimeout)
	})
	t.Run("test_extract_grpc_opts", func(t *testing.T) {
		opts := extractGrpcOptions(cfg)
		if len(opts) == 0 {
			t.Error("no grpc opts after extraction")
		}
	})
	t.Run("test_extract_things_opts", func(t *testing.T) {
		opts := extractThingsOptions(cfg)
		if len(opts) == 0 {
			t.Error("no things opts after extraction")
		}
	})
}

func TestDumpsNoErrors(t *testing.T) {
	cfg := getDefaultInstance()
	dumpConfiguration(cfg)

	t.Run("test_dump_config_nil", func(t *testing.T) {
		dumpConfiguration(nil)
	})
	t.Run("test_dump_config_log_null", func(t *testing.T) {
		logCfg := cfg.Log
		cfg.Log = nil
		dumpConfiguration(cfg)
		cfg.Log = logCfg
	})
	t.Run("test_dump_config_mgr_null", func(t *testing.T) {
		managerCfg := cfg.ManagerConfig
		cfg.ManagerConfig = nil
		dumpConfiguration(cfg)
		cfg.ManagerConfig = managerCfg
	})
	t.Run("test_dump_config_net_null", func(t *testing.T) {
		networkCfg := cfg.NetworkConfig
		cfg.NetworkConfig = nil
		dumpConfiguration(cfg)
		cfg.NetworkConfig = networkCfg
	})
	t.Run("test_dump_config_grpc_null", func(t *testing.T) {
		grpcCfg := cfg.GrpcServerConfig
		cfg.GrpcServerConfig = nil
		dumpConfiguration(cfg)
		cfg.GrpcServerConfig = grpcCfg
	})
	t.Run("test_dump_config_things_null", func(t *testing.T) {
		thingsCfg := cfg.ThingsConfig
		cfg.ThingsConfig = nil
		dumpConfiguration(cfg)
		cfg.ThingsConfig = thingsCfg
	})
}

func TestSetCommandFlags(t *testing.T) {
	var cmd = &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDaemon(cmd)

		},
	}
	setupCommandFlags(cmd)

	tests := map[string]struct {
		flag         string
		expectedType string
	}{
		"test_flags_log-level": {
			flag:         "log-level",
			expectedType: reflect.String.String(),
		},
		"test_flags_log-file": {
			flag:         "log-file",
			expectedType: reflect.String.String(),
		},
		"test_flags_log-file-size": {
			flag:         "log-file-size",
			expectedType: reflect.Int.String(),
		},
		"test_flags_log-file-count": {
			flag:         "log-file-count",
			expectedType: reflect.Int.String(),
		},
		"test_flags_log-file-max-age": {
			flag:         "log-file-max-age",
			expectedType: reflect.Int.String(),
		},
		"test_flags_log-syslog": {
			flag:         "log-syslog",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_cm-home-dir": {
			flag:         "cm-home-dir",
			expectedType: reflect.String.String(),
		},
		"test_flags_cm-exec-root-dir": {
			flag:         "cm-exec-root-dir",
			expectedType: reflect.String.String(),
		},
		"test_flags_cm-net-sid": {
			flag:         "cm-net-sid",
			expectedType: reflect.String.String(),
		},
		"test_flags_cm-deflt-ctrs-stop-timeout": {
			flag:         "cm-deflt-ctrs-stop-timeout",
			expectedType: reflect.String.String(),
		},
		"test_flags_ccl-default-ns": {
			flag:         "ccl-default-ns",
			expectedType: reflect.String.String(),
		},
		"test_flags_ccl-ap": {
			flag:         "ccl-ap",
			expectedType: reflect.String.String(),
		},
		"test_flags_ccl-insecure-registries": {
			flag:         "ccl-insecure-registries",
			expectedType: "stringSlice",
		},
		"test_flags_ccl-exec-root-dir": {
			flag:         "ccl-exec-root-dir",
			expectedType: reflect.String.String(),
		},
		"test_flags_ccl-home-dir": {
			flag:         "ccl-home-dir",
			expectedType: reflect.String.String(),
		},
		"test_flags_ccl-image-dec-keys": {
			flag:         "ccl-image-dec-keys",
			expectedType: "stringSlice",
		},
		"test_flags_ccl-image-dec-recipients": {
			flag:         "ccl-image-dec-recipients",
			expectedType: "stringSlice",
		},
		"test_flags_ccl-runc-runtime": {
			flag:         "ccl-runc-runtime",
			expectedType: reflect.String.String(),
		},
		"test_flags_ccl-image-expiry": {
			flag:         "ccl-image-expiry",
			expectedType: "duration",
		},
		"test_flags_ccl-image-expiry-disable": {
			flag:         "ccl-image-expiry-disable",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_ccl-lease-id": {
			flag:         "ccl-lease-id",
			expectedType: reflect.String.String(),
		},
		"test_flags_ccl-image-verifier-type": {
			flag:         "ccl-image-verifier-type",
			expectedType: reflect.String.String(),
		},
		"test_flags_ccl-image-verifier-config": {
			flag:         "ccl-image-verifier-config",
			expectedType: "stringSlice",
		},
		"test_flags_net-type": {
			flag:         "net-type",
			expectedType: reflect.String.String(),
		},
		"test_flags_net-home-dir": {
			flag:         "net-home-dir",
			expectedType: reflect.String.String(),
		},
		"test_flags_net-exec-root-dir": {
			flag:         "net-exec-root-dir",
			expectedType: reflect.String.String(),
		},
		"test_flags_net-tbr-disable": {
			flag:         "net-tbr-disable",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_net-br-name": {
			flag:         "net-br-name",
			expectedType: reflect.String.String(),
		},
		"test_flags_net-br-ip4": {
			flag:         "net-br-ip4",
			expectedType: reflect.String.String(),
		},
		"test_flags_net-br-fcidr4": {
			flag:         "net-br-fcidr4",
			expectedType: reflect.String.String(),
		},
		"test_flags_net-br-gwip4": {
			flag:         "net-br-gwip4",
			expectedType: reflect.String.String(),
		},
		"net-br-enable-ip6": {
			flag:         "net-br-enable-ip6",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_net-br-mtu": {
			flag:         "net-br-mtu",
			expectedType: reflect.Int.String(),
		},
		"test_flags_net-br-icc": {
			flag:         "net-br-icc",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_net-br-ipt": {
			flag:         "net-br-ipt",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_net-br-ipfw": {
			flag:         "net-br-ipfw",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_net-br-ipmasq": {
			flag:         "net-br-ipmasq",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_net-br-ulp": {
			flag:         "net-br-ulp",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_grpc-serv-netp": {
			flag:         "grpc-serv-netp",
			expectedType: reflect.String.String(),
		},
		"test_flags_grpc-serv-ap": {
			flag:         "grpc-serv-ap",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-enable": {
			flag:         "things-enable",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_things-home-dir": {
			flag:         "things-home-dir",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-features": {
			flag:         "things-features",
			expectedType: "stringSlice",
		},
		"test_flags_things-conn-broker": {
			flag:         "things-conn-broker",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-conn-keep-alive": {
			flag:         "things-conn-keep-alive",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-disconnect-timeout": {
			flag:         "things-conn-disconnect-timeout",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-client-username": {
			flag:         "things-conn-client-username",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-conn-client-password": {
			flag:         "things-conn-client-password",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-conn-connect-timeout": {
			flag:         "things-conn-connect-timeout",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-ack-timeout": {
			flag:         "things-conn-ack-timeout",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-sub-timeout": {
			flag:         "things-conn-sub-timeout",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-unsub-timeout": {
			flag:         "things-conn-unsub-timeout",
			expectedType: reflect.Int64.String(),
		},
		"test_flags_things-conn-root-ca": {
			flag:         "things-conn-root-ca",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-conn-client-cert": {
			flag:         "things-conn-client-cert",
			expectedType: reflect.String.String(),
		},
		"test_flags_things-conn-client-key": {
			flag:         "things-conn-client-key",
			expectedType: reflect.String.String(),
		},
		"test_flags_deployment-enable": {
			flag:         "deployment-enable",
			expectedType: reflect.Bool.String(),
		},
		"test_flags_deployment-mode": {
			flag:         "deployment-mode",
			expectedType: reflect.String.String(),
		},
		"test_flags_deployment-home-dir": {
			flag:         "deployment-home-dir",
			expectedType: reflect.String.String(),
		},
		"test_flags_deployment-ctr-dir": {
			flag:         "deployment-ctr-dir",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-broker": {
			flag:         "conn-broker-url",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-keep-alive": {
			flag:         "conn-keep-alive",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-disconnect-timeout": {
			flag:         "conn-disconnect-timeout",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-client-username": {
			flag:         "conn-client-username",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-client-password": {
			flag:         "conn-client-password",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-connect-timeout": {
			flag:         "conn-connect-timeout",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-ack-timeout": {
			flag:         "conn-ack-timeout",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-sub-timeout": {
			flag:         "conn-sub-timeout",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-unsub-timeout": {
			flag:         "conn-unsub-timeout",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-root-ca": {
			flag:         "conn-root-ca",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-client-cert": {
			flag:         "conn-client-cert",
			expectedType: reflect.String.String(),
		},
		"test_flags-conn-client-key": {
			flag:         "conn-client-key",
			expectedType: reflect.String.String(),
		},
		"test_flags-ua-enable": {
			flag:         "ua-enable",
			expectedType: reflect.Bool.String(),
		},
		"test_flags-ua-domain": {
			flag:         "ua-domain",
			expectedType: reflect.String.String(),
		},
		"test_flags-ua-system-containers": {
			flag:         "ua-system-containers",
			expectedType: "stringSlice",
		},
		"test_flags-ua-verbose-inventory-report": {
			flag:         "ua-verbose-inventory-report",
			expectedType: reflect.Bool.String(),
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			flag := cmd.Flag(testCase.flag)
			if flag.Value.Type() != testCase.expectedType {
				t.Errorf("incorrect type: %s for flag %s, expecting: %s", flag.Value.Type(), flag.Name, testCase.expectedType)
			}
			if flag == nil {
				t.Errorf("flag %s, not found", testCase.flag)
			}
		})
	}
}

func TestParseRegistryConfigs(t *testing.T) {
	const (
		registriesDaemonConfig = "../pkg/testutil/config/daemon-config-insecure-registries.json"
	)
	t.Run("test_parse_registry_configs", func(t *testing.T) {
		cfg := getDefaultInstance()
		err := loadLocalConfig(registriesDaemonConfig, cfg)
		if err != nil {
			t.Errorf("error while loading local configs with insecure registries")
		}
		registryConfigs := parseRegistryConfigs(cfg.ContainerClientConfig.CtrRegistryConfigs, cfg.ContainerClientConfig.CtrInsecureRegistries)
		if registryConfigs == nil || len(registryConfigs) == 0 {
			t.Errorf("error while parsing registry configs")
		}
	})
	t.Run("test_parse_registry_configs_null_insecure", func(t *testing.T) {
		cfg := getDefaultInstance()
		_ = loadLocalConfig(registriesDaemonConfig, cfg)
		cfg.ContainerClientConfig.CtrInsecureRegistries = nil
		registryConfigs := parseRegistryConfigs(cfg.ContainerClientConfig.CtrRegistryConfigs, cfg.ContainerClientConfig.CtrInsecureRegistries)
		if registryConfigs == nil || len(registryConfigs) == 0 {
			t.Errorf("error while parsing registry configs, with nil insecure registries")
		}
	})
	t.Run("test_parse_registry_configs_null_registry_config", func(t *testing.T) {
		cfg := getDefaultInstance()
		_ = loadLocalConfig(registriesDaemonConfig, cfg)
		cfg.ContainerClientConfig.CtrRegistryConfigs = nil
		registryConfigs := parseRegistryConfigs(cfg.ContainerClientConfig.CtrRegistryConfigs, cfg.ContainerClientConfig.CtrInsecureRegistries)
		if registryConfigs == nil || len(registryConfigs) == 0 {
			t.Errorf("error while parsing registry configs, with nil insecure registries")
		}
	})

	cfg := getDefaultInstance()
	_ = loadLocalConfig(registriesDaemonConfig, cfg)
	registryConfigs := parseRegistryConfigs(cfg.ContainerClientConfig.CtrRegistryConfigs, cfg.ContainerClientConfig.CtrInsecureRegistries)
	t.Run("test_parse_registry_configs_basic", func(t *testing.T) {
		basicAuthCfg := registryConfigs["my-basic-auth-host.acme"]
		if basicAuthCfg == nil {
			t.Errorf("basic auth config missing after parse")
		}
		if basicAuthCfg.IsInsecure {
			t.Errorf("basic auth config isInsecure not parsed correctly")
		}
		if basicAuthCfg.Credentials.UserID != "my-username" {
			t.Errorf("basic auth config username not parsed correctly")
		}
		if basicAuthCfg.Credentials.Password != "my-plaintext-password" {
			t.Errorf("basic auth config password not parsed correctly")
		}
		if basicAuthCfg.Transport != nil {
			t.Errorf("basic auth config transport must be nil after parsing")
		}
	})
	t.Run("test_parse_registry_configs_tls", func(t *testing.T) {
		tlsCfg := registryConfigs["my-tls-host.acme"]
		if tlsCfg == nil {
			t.Errorf("tls config missing after parse")
		}
		if tlsCfg.IsInsecure {
			t.Errorf("tls config isInsecure not parsed correctly")
		}
		if tlsCfg.Transport.ClientCert != "/my/secure/path/client.cert" {
			t.Errorf("tls config client cert not parsed correctly")
		}
		if tlsCfg.Transport.ClientKey != "/my/secure/path/client.key" {
			t.Errorf("tls config client key not parsed correctly")
		}
		if tlsCfg.Transport.RootCA != "/my/secure/path/ca.crt" {
			t.Errorf("tls config client root ca not parsed correctly")
		}
		if tlsCfg.Credentials != nil {
			t.Errorf("tls config credentials must be nil after parsing")
		}
	})
	t.Run("test_parse_registry_configs_tls_auth", func(t *testing.T) {
		basicAuthTLS := registryConfigs["my-tls-with-basic-auth-host.acme"]
		if basicAuthTLS == nil {
			t.Errorf("basic auth config missing after parse")
		}
		if basicAuthTLS.IsInsecure {
			t.Errorf("basic auth with tls isInsecure not parsed correctly")
		}
		if basicAuthTLS.Transport.ClientCert != "/my/secure/path/client.cert" {
			t.Errorf("basic auth with tls config client cert not parsed correctly")
		}
		if basicAuthTLS.Transport.ClientKey != "/my/secure/path/client.key" {
			t.Errorf("basic auth with tls config client key not parsed correctly")
		}
		if basicAuthTLS.Transport.RootCA != "/my/secure/path/ca.crt" {
			t.Errorf("basic auth with tls config client root ca not parsed correctly")
		}
		if basicAuthTLS.Credentials.UserID != "my-username" {
			t.Errorf("basic auth with tls config username not parsed correctly")
		}
		if basicAuthTLS.Credentials.Password != "my-plaintext-password" {
			t.Errorf("basic auth with tls config password not parsed correctly")
		}
	})
	t.Run("test_parse_registry_configs_insecure_no_port", func(t *testing.T) {
		insecureNoPort := registryConfigs["my-insecure-host.acme"]
		if insecureNoPort == nil {
			t.Errorf("insecure registry without port config missing after parse")
		}
		if insecureNoPort.Transport != nil {
			t.Errorf("insecure registry config transport must be nil after parsing")
		}
		if insecureNoPort.Credentials != nil {
			t.Errorf("insecure registry config credentials must be nil after parsing")
		}

	})
	t.Run("test_parse_registry_configs_insecure_with_port", func(t *testing.T) {
		insecureWithPort := registryConfigs["192.101.1.101:500"]
		if insecureWithPort == nil {
			t.Errorf("insecure registry with port config missing after parse")
		}
		if insecureWithPort.Transport != nil {
			t.Errorf("insecure registry config transport must be nil after parsing")
		}
		if insecureWithPort.Credentials != nil {
			t.Errorf("insecure registry config credentials must be nil after parsing")
		}
	})
	t.Run("test_parse_registry_configs_size", func(t *testing.T) {
		if len(registryConfigs) != 5 {
			t.Errorf("registry config length %d not correct after parse, expected %d", len(registryConfigs), 5)
		}
	})
}

func TestRunLock(t *testing.T) {
	t.Run("test_lock_create", func(t *testing.T) {
		lock, lockErr := newRunLock(lockFileName)
		if lockErr != nil {
			t.Error("error while creating run lock")
		}
		if lock == nil {
			t.Error("couldn't create lock")
		}
	})

	t.Run("test_lock_try_lock_new_goroutine", func(t *testing.T) {
		lock, _ := newRunLock(lockFileName)
		lockErr := lock.TryLock()
		if lockErr == nil {
			defer func() {
				lock.Unlock()
				_ = os.Remove(lockFileName)
			}()
			go func() {
				secondLock, err := newRunLock(lockFileName)
				if err != nil {
					t.Error("could create second lock")
				}
				lockErrorAlreadyLocked := secondLock.TryLock()
				if lockErrorAlreadyLocked == nil {
					t.Error("run lock locked twice")
				}
			}()
			time.Sleep(1 * time.Second)
		} else {
			t.Error("couldn't create lock")
		}
	})
}

func TestImageVerifierConfig(t *testing.T) {
	local := &config{}
	_ = loadLocalConfig("../pkg/testutil/config/daemon-config-image-verifier.json", local)
	testutil.AssertEqual(t, "notation", local.ContainerClientConfig.CtrImageVerifierType)
	expected := map[string]string{"configDir": "/path/notation/config", "libexecDir": "/path/notation/libexec"}
	assertStringMap(t, expected, local.ContainerClientConfig.CtrImageVerifierConfig)
}

func TestImageVerifierFlag(t *testing.T) {
	const (
		testSinglePair    = "key=value"
		testMultiplePairs = "key0=value0,key1=value1,key2=value2"
	)

	tests := map[string]struct {
		value               string
		expectedValue       map[string]string
		expectedStringPairs []string
		expectedErr         error
	}{
		"test_empty_error": {
			expectedErr: log.NewError("the image verifier config could not be empty"),
		},
		"test_parse_error": {
			value:       "invalid",
			expectedErr: log.NewError("could not parse image verification config, invalid key-value pair - invalid"),
		},
		"test_single_config_no_error": {
			value:               testSinglePair,
			expectedStringPairs: []string{testSinglePair},
			expectedValue:       map[string]string{"key": "value"},
		},
		"test_multiple_configs_no_error": {
			value:               testMultiplePairs,
			expectedStringPairs: strings.Split(testMultiplePairs, ","),
			expectedValue:       map[string]string{"key0": "value0", "key1": "value1", "key2": "value2"},
		},
	}
	for testName, test := range tests {
		t.Run(testName, func(t *testing.T) {
			verifierConfig := &verifierConfig{}
			testutil.AssertError(t, test.expectedErr, verifierConfig.Set(test.value))
			for _, expectedPair := range test.expectedStringPairs {
				testutil.AssertTrue(t, strings.Contains(verifierConfig.String(), expectedPair))
			}
			testutil.AssertEqual(t, len(test.expectedValue), len(*verifierConfig))
			assertStringMap(t, test.expectedValue, *verifierConfig)

		})

	}
}

func assertStringMap(t *testing.T, expected, actual map[string]string) {
	testutil.AssertEqual(t, len(expected), len(actual))
	for key, expectedValue := range expected {
		value, ok := actual[key]
		testutil.AssertTrue(t, ok)
		testutil.AssertEqual(t, expectedValue, value)
	}
}

// TODO test the behavior of the daemon towards its services (start, stop), with mocked instanced of GRPC service etc.
