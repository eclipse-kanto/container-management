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

package client

import (
	"context"
	"fmt"
	"io"

	pbcontainers "github.com/eclipse-kanto/container-management/containerm/api/services/containers"
	pbsysinfo "github.com/eclipse-kanto/container-management/containerm/api/services/sysinfo"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	sysinfotypes "github.com/eclipse-kanto/container-management/containerm/sysinfo/types"
	"github.com/eclipse-kanto/container-management/containerm/util/protobuf"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type client struct {
	connection           *grpc.ClientConn
	grpcContainersClient pbcontainers.ContainersClient
	grpcSystemInfoClient pbsysinfo.SystemInfoClient
}

// Create a new container.
func (cl *client) Create(ctx context.Context, config *types.Container) (*types.Container, error) {
	pbResponse, err := cl.grpcContainersClient.Create(ctx, &pbcontainers.CreateContainerRequest{Container: protobuf.ToProtoContainer(config)})
	if err != nil {
		return nil, err
	}
	return protobuf.ToInternalContainer(pbResponse.Container), nil
}

// Get the detailed information of container.
func (cl *client) Get(ctx context.Context, id string) (*types.Container, error) {
	pbResponse, err := cl.grpcContainersClient.Get(ctx, &pbcontainers.GetContainerRequest{Id: id})
	if err != nil {
		return nil, err
	}
	return protobuf.ToInternalContainer(pbResponse.Container), nil
}

// List returns the list of containers.
func (cl *client) List(ctx context.Context, filters ...Filter) ([]*types.Container, error) {
	pbResponse, err := cl.grpcContainersClient.List(ctx, &pbcontainers.ListContainersRequest{})
	if err != nil {
		return nil, err
	}
	if pbResponse.Containers == nil {
		return nil, nil
	}
	containers := []*types.Container{}
	for _, ctr := range pbResponse.Containers {
		containers = append(containers, protobuf.ToInternalContainer(ctr))
	}
	if filters == nil || len(filters) == 0 {
		return containers, nil
	}
	var filteredContainers []*types.Container
	for _, ctr := range containers {
		matched := true
		for _, filter := range filters {
			if matched = filter(ctr); !matched {
				break
			}
		}
		if matched {
			filteredContainers = append(filteredContainers, ctr)
		}
	}
	return filteredContainers, nil
}

// Start a container.
func (cl *client) Start(ctx context.Context, id string) error {
	_, err := cl.grpcContainersClient.Start(ctx, &pbcontainers.StartContainerRequest{Id: id})
	return err
}

// Stop a container.
func (cl *client) Stop(ctx context.Context, id string, stopOpts *types.StopOpts) error {
	_, err := cl.grpcContainersClient.Stop(ctx, &pbcontainers.StopContainerRequest{Id: id, StopOptions: protobuf.ToProtoStopOptions(stopOpts)})
	return err
}

// Update a container.
func (cl *client) Update(ctx context.Context, id string, updateOpts *types.UpdateOpts) error {
	_, err := cl.grpcContainersClient.Update(ctx, &pbcontainers.UpdateContainerRequest{Id: id, UpdateOptions: protobuf.ToProtoUpdateOptions(updateOpts)})
	return err
}

// Attach to a container's IO.
func (cl *client) Attach(ctx context.Context, id string, stdin bool) (io.Writer, io.ReadCloser, error) {
	ctrAttachClient, err := cl.grpcContainersClient.Attach(ctx)
	if err != nil {
		return nil, nil, err
	}

	ctrAttachClient.Send(&pbcontainers.AttachContainerRequest{
		Id:          id,
		StdIn:       stdin,
		DataToWrite: nil,
	})

	reader, err := NewReader(ctx, id, stdin, ctrAttachClient)
	if err != nil {
		return nil, nil, err
	}
	writer, err := NewWriter(ctx, id, stdin, ctrAttachClient)
	if err != nil {
		return nil, nil, err
	}

	return writer, reader, nil
}

// Restart restart a running container.
func (cl *client) Restart(ctx context.Context, id string, timeout int64) error {
	_, err := cl.grpcContainersClient.Restart(ctx, &pbcontainers.RestartContainerRequest{Id: id})
	return err
}

// Pause a container.
func (cl *client) Pause(ctx context.Context, id string) error {
	_, err := cl.grpcContainersClient.Pause(ctx, &pbcontainers.PauseContainerRequest{Id: id})
	return err
}

// Resumes a container.
func (cl *client) Resume(ctx context.Context, id string) error {
	_, err := cl.grpcContainersClient.Unpause(ctx, &pbcontainers.UnpauseContainerRequest{Id: id})
	return err
}

// Rename renames a container.
func (cl *client) Rename(ctx context.Context, id string, name string) error {
	_, err := cl.grpcContainersClient.Rename(ctx, &pbcontainers.RenameContainerRequest{Id: id, Name: name})
	return err
}

// Remove removes a container, it may be running or stopped and so on.
func (cl *client) Remove(ctx context.Context, id string, force bool, removecache bool, stopOpts *types.StopOpts) error {
	_, err := cl.grpcContainersClient.Remove(ctx, &pbcontainers.RemoveContainerRequest{Id: id, Force: force, Removecache: removecache, StopOptions: protobuf.ToProtoStopOptions(stopOpts)})
	return err
}

// func (cl *client) RemoveImage(ctx context.Context, image types.Image) error {
// 	_, err := cl.grpcContainersClient.RemoveImage(ctx, &pbcontainers.RemoveContainerImageRequest{Image: image})
// 	return err
// }

func (cl *client) Dispose() error {
	return cl.connection.Close()
}

func (cl *client) ProjectInfo(ctx context.Context) (sysinfotypes.ProjectInfo, error) {
	pbResponse, err := cl.grpcSystemInfoClient.ProjectInfo(ctx, &empty.Empty{})
	if err != nil {
		return sysinfotypes.ProjectInfo{}, err
	}

	return protobuf.ToInternalProjectInfo(pbResponse.ProjectInfo), nil
}

// Logs print the logs of a container.
func (cl *client) Logs(ctx context.Context, id string, tail int32) error {
	stream, err := cl.grpcContainersClient.Logs(ctx, &pbcontainers.GetLogsRequest{Id: id, Tail: tail})
	if err != nil {
		return fmt.Errorf("error while opening stream: %s", err)
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", resp.Log)
	}

	return nil
}
