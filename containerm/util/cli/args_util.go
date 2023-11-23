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

package cli

import (
	"context"

	"github.com/eclipse-kanto/container-management/containerm/client"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/log"
)

// ValidateContainerByNameArgsSingle validates the parameters and returns a container by provided name
func ValidateContainerByNameArgsSingle(ctx context.Context, args []string, providedContainerName string, kantoCMClient client.Client) (*types.Container, error) {
	if len(args) == 0 && providedContainerName == "" {
		return nil, log.NewError("You must provide either an ID or a name for the container via --name (-n) ")
	} else if len(args) == 1 && providedContainerName != "" {
		return nil, log.NewError("Container ID and --name (-n) cannot be provided at the same time - use only one of them")
	}
	var (
		container *types.Container
		err       error
	)
	// parse parameters
	if len(args) >= 1 {
		if container, err = kantoCMClient.Get(ctx, args[0]); err != nil {
			return nil, err
		} else if container == nil {
			return nil, log.NewErrorf("The requested container with ID = %s was not found.", args[0])
		}
	} else {
		ctrs, listErr := kantoCMClient.List(ctx, client.WithName(providedContainerName))
		if listErr != nil {
			return nil, listErr
		} else if len(ctrs) == 0 {
			return nil, log.NewErrorf("The requested container with name = %s was not found. Try using an ID instead.", providedContainerName)
		} else if len(ctrs) != 1 {
			return nil, log.NewErrorf("There are more than one containers with name = %s. Try using an ID instead.", providedContainerName)
		}
		container = ctrs[0]
	}
	return container, nil
}
