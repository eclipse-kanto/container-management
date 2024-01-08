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

package services

import (
	"context"
	"fmt"
	"os"

	pbcontainers "github.com/eclipse-kanto/container-management/containerm/api/services/containers"
	pbcontainerstypes "github.com/eclipse-kanto/container-management/containerm/api/types/containers"
	"github.com/eclipse-kanto/container-management/containerm/containers/types"
	"github.com/eclipse-kanto/container-management/containerm/mgr"
	"github.com/eclipse-kanto/container-management/containerm/streams"
	"github.com/eclipse-kanto/container-management/containerm/util"
	"github.com/eclipse-kanto/container-management/containerm/util/protobuf"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type containers struct {
	mgr mgr.ContainerManager
}

func (server *containers) Register(grpcServer *grpc.Server) error {
	pbcontainers.RegisterContainersServer(grpcServer, server)
	return nil
}

func (server *containers) Create(ctx context.Context, request *pbcontainers.CreateContainerRequest) (*pbcontainers.CreateContainerResponse, error) {
	container, err := server.mgr.Create(ctx, protobuf.ToInternalContainer(request.Container))
	if err != nil {
		return nil, err
	}
	response := &pbcontainers.CreateContainerResponse{
		Container: protobuf.ToProtoContainer(container),
	}
	return response, nil
}

func (server *containers) Get(ctx context.Context, request *pbcontainers.GetContainerRequest) (*pbcontainers.GetContainerResponse, error) {
	container, err := server.mgr.Get(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	pbCtr := protobuf.ToProtoContainer(container)

	response := &pbcontainers.GetContainerResponse{
		Container: pbCtr,
	}
	return response, err
}

func (server *containers) List(ctx context.Context, request *pbcontainers.ListContainersRequest) (*pbcontainers.ListContainersResponse, error) {
	ctrs, err := server.mgr.List(ctx)
	l := len(ctrs)
	pbCtrs := make([]*pbcontainerstypes.Container, l)
	for i, ctr := range ctrs {
		pbCtrs[i] = protobuf.ToProtoContainer(ctr)
	}

	response := &pbcontainers.ListContainersResponse{
		Containers: pbCtrs,
	}
	return response, err
}

func (server *containers) ListStream(*pbcontainers.ListContainersRequest, pbcontainers.Containers_ListStreamServer) error {
	return nil
}

func (server *containers) Start(ctx context.Context, request *pbcontainers.StartContainerRequest) (*empty.Empty, error) {
	err := server.mgr.Start(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (server *containers) Attach(attachServer pbcontainers.Containers_AttachServer) error {
	req, err := attachServer.Recv()
	if err != nil {
		return err
	}
	ctrID := req.Id
	stdIn := req.StdIn

	ctx := context.Background()

	reader, err := NewReader(ctx, ctrID, stdIn, attachServer)
	if err != nil {
		return err
	}

	writer, err := NewWriter(ctx, ctrID, stdIn, attachServer)
	if err != nil {
		return err
	}

	var attach = new(streams.AttachConfig)

	attach.UseStdin = stdIn
	attach.Stdin = reader
	attach.UseStdout = true
	attach.Stdout = writer
	attach.UseStderr = true
	attach.Stderr = writer

	if err := server.mgr.Attach(ctx, ctrID, attach); err != nil {
		writer.Write([]byte(err.Error() + "\r\n"))
		return err
	}
	return nil
}

func (server *containers) Stop(ctx context.Context, request *pbcontainers.StopContainerRequest) (*empty.Empty, error) {
	err := server.mgr.Stop(ctx, request.Id, protobuf.ToInternalStopOptions(request.StopOptions))
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (server *containers) Update(ctx context.Context, request *pbcontainers.UpdateContainerRequest) (*empty.Empty, error) {
	err := server.mgr.Update(ctx, request.Id, protobuf.ToInternalUpdateOptions(request.UpdateOptions))
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (server *containers) Restart(context.Context, *pbcontainers.RestartContainerRequest) (*empty.Empty, error) {
	return nil, nil
}

func (server *containers) Pause(ctx context.Context, request *pbcontainers.PauseContainerRequest) (*empty.Empty, error) {
	err := server.mgr.Pause(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (server *containers) Unpause(ctx context.Context, request *pbcontainers.UnpauseContainerRequest) (*empty.Empty, error) {
	err := server.mgr.Unpause(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (server *containers) Rename(ctx context.Context, request *pbcontainers.RenameContainerRequest) (*empty.Empty, error) {
	err := server.mgr.Rename(ctx, request.Id, request.Name)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (server *containers) Remove(ctx context.Context, request *pbcontainers.RemoveContainerRequest) (*empty.Empty, error) {
	err := server.mgr.Remove(ctx, request.Id, request.Force, protobuf.ToInternalStopOptions(request.StopOptions))
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (server *containers) Logs(request *pbcontainers.GetLogsRequest, srv pbcontainers.Containers_LogsServer) error {
	container, err := server.mgr.Get(context.Background(), request.Id)
	if err != nil {
		return err
	}

	if container.State != nil && container.State.Status == types.Created {
		return fmt.Errorf("there are no logs for container with status \"Created\"")
	}

	logFile, err := getLogFilePath(container)
	if err != nil {
		return err
	}

	isFile, err := util.IsFile(logFile)
	if !isFile {
		return fmt.Errorf("log file not found %s", logFile)
	}
	if err != nil {
		return err
	}

	f, err := os.Open(logFile)
	if err != nil {
		return err
	}

	defer f.Close()

	if request.Tail < 0 {
		return sendAllLogs(f, srv)
	}

	return tailLogs(f, srv, int(request.Tail))
}
