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

package ctr

import (
	"net"
	"net/http"
	"time"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

const (
	registryHostSchemeHTTP          = "http"
	registryHostSchemeHTTPS         = "https"
	registryHostPathV2              = "/v2"
	registryHostCapabilitiesDefault = docker.HostCapabilityPull | docker.HostCapabilityResolve

	registryResolverDialContextTimeout             = 30 * time.Second
	registryResolverDialContextKeepAlive           = 30 * time.Second
	registryResolverTransportMaxIdeConns           = 10
	registryResolverTransportIdleConnTimeout       = 30 * time.Second
	registryResolverTransportTLSHandshakeTimeout   = 10 * time.Second
	registryResolverTransportExpectContinueTimeout = 5 * time.Second
)

type containerImageRegistriesResolver interface {
	ResolveImageRegistry(imageRegistryHost string) remotes.Resolver
}

type ctrImagesResolver struct {
	registryConfigurations map[string]*RegistryConfig
	registryHosts          map[string][]docker.RegistryHost
}

func newContainerImageRegistriesResolver(registryConfigs map[string]*RegistryConfig) containerImageRegistriesResolver {
	resolver := &ctrImagesResolver{
		registryConfigurations: registryConfigs,
		registryHosts:          map[string][]docker.RegistryHost{},
	}
	resolver.processImageRegistries()
	return resolver
}

func (resolver *ctrImagesResolver) ResolveImageRegistry(imageRegistryHost string) remotes.Resolver {
	_, configExists := resolver.registryConfigurations[imageRegistryHost]
	if !configExists {
		log.Warn("no preconfigured image resolver is currently available for image registry host %s", imageRegistryHost)
		return nil
	}
	return docker.NewResolver(docker.ResolverOptions{
		Hosts: resolver.getRegistryHosts,
	})
}

// when we have secured registries support - this method must be called for these registries also
func (resolver *ctrImagesResolver) getRegistryHosts(imageHost string) ([]docker.RegistryHost, error) {
	log.Debug("hook for retrieving the host info for image host %s called", imageHost)
	res := resolver.registryHosts[imageHost]
	if res == nil {
		return nil, log.NewErrorf("no registry hosts found for host [%s]", imageHost)
	}
	log.Debug("found image registry hosts for host %s - resolution info [%s]", imageHost, res)
	return res, nil
}
func (resolver *ctrImagesResolver) getRegistryAuthCreds(registryHost string) (string, string, error) {
	log.Debug("required registry auth credentials for host %s", registryHost)
	cred := resolver.registryConfigurations[registryHost].Credentials
	if cred == nil {
		err := log.NewErrorf("no credentials could be found for registry host %s", registryHost)
		log.WarnErr(err, "error getting auth credentials for registry host %s", registryHost)
		return "", "", err
	}
	log.Debug("found required registry auth credentials for host %s ", registryHost)
	return cred.UserID, cred.Password, nil
}

func (resolver *ctrImagesResolver) processImageRegistries() {
	if resolver.registryConfigurations == nil || len(resolver.registryConfigurations) == 0 {
		log.Debug("no registries configurations provided to generate hosts resolution")
		return
	}
	if resolver.registryHosts == nil {
		resolver.registryHosts = make(map[string][]docker.RegistryHost)
	}
	for host, config := range resolver.registryConfigurations {

		//needed for insecure registries with self-signed certificates
		var httpConfig docker.RegistryHost
		var tr *http.Transport

		tlsConfig := createDefaultTLSConfig(config.IsInsecure)
		if config.Transport != nil && config.IsInsecure {
			log.Warn("a TLS configuration for registry host %s is provided but the registry is marked as insecure - the TLS config will not be applied", host)
		}

		if config.IsInsecure {
			httpConfig = docker.RegistryHost{
				Client:       http.DefaultClient,
				Host:         host,
				Scheme:       registryHostSchemeHTTP,
				Path:         registryHostPathV2,
				Capabilities: registryHostCapabilitiesDefault,
			}

			if config.Credentials != nil {
				httpConfig.Authorizer = docker.NewDockerAuthorizer(docker.WithAuthClient(http.DefaultClient), docker.WithAuthCreds(resolver.getRegistryAuthCreds))
			}

			tr = &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   registryResolverDialContextTimeout,
					KeepAlive: registryResolverDialContextKeepAlive,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          registryResolverTransportMaxIdeConns,
				IdleConnTimeout:       registryResolverTransportIdleConnTimeout,
				TLSHandshakeTimeout:   registryResolverTransportTLSHandshakeTimeout,
				TLSClientConfig:       tlsConfig,
				ExpectContinueTimeout: registryResolverTransportExpectContinueTimeout,
			}
		} else {
			if config.Transport != nil {
				if err := applyLocalTLSConfig(config.Transport, tlsConfig); err != nil {
					log.WarnErr(err, "could not process provided TLS configuration - default will be used for registry host %s", host)
					tlsConfig = createDefaultTLSConfig(config.IsInsecure)
				} else {
					log.Debug("successfully applied TLS configuration for registry host %s", host)
				}
			}
			tr = &http.Transport{
				Proxy: http.ProxyFromEnvironment,
				DialContext: (&net.Dialer{
					Timeout:   registryResolverDialContextTimeout,
					KeepAlive: registryResolverDialContextKeepAlive,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:          registryResolverTransportMaxIdeConns,
				IdleConnTimeout:       registryResolverTransportIdleConnTimeout,
				TLSHandshakeTimeout:   registryResolverTransportTLSHandshakeTimeout,
				TLSClientConfig:       tlsConfig,
				ExpectContinueTimeout: registryResolverTransportExpectContinueTimeout,
			}
		}

		httpsClient := &http.Client{Transport: tr}
		httpsConfig := docker.RegistryHost{
			Client:       httpsClient,
			Host:         host,
			Scheme:       registryHostSchemeHTTPS,
			Path:         registryHostPathV2,
			Capabilities: registryHostCapabilitiesDefault,
		}
		if config.Credentials != nil {
			httpsConfig.Authorizer = docker.NewDockerAuthorizer(docker.WithAuthClient(http.DefaultClient), docker.WithAuthCreds(resolver.getRegistryAuthCreds))
		}

		if config.IsInsecure {
			log.Debug("added image registry host resolution info for image registry [%s]: [%s], [%s] ", host, httpConfig, httpsConfig)
			resolver.registryHosts[host] = []docker.RegistryHost{httpConfig, httpsConfig}
		} else {
			log.Debug("added image registry host resolution info for image registry [%s]: [%s], [%s] ", host, httpsConfig)
			resolver.registryHosts[host] = []docker.RegistryHost{httpsConfig}
		}
	}
}
