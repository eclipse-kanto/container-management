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

	pbsysinfo "github.com/eclipse-kanto/container-management/containerm/api/services/sysinfo"
	"github.com/eclipse-kanto/container-management/containerm/sysinfo"
	"github.com/eclipse-kanto/container-management/containerm/util/protobuf"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type systemInfo struct {
	sysInfoMgr sysinfo.SystemInfoManager
}

func (server *systemInfo) Register(grpcServer *grpc.Server) error {
	pbsysinfo.RegisterSystemInfoServer(grpcServer, server)
	return nil
}

func (server *systemInfo) ProjectInfo(ctx context.Context, request *empty.Empty) (*pbsysinfo.ProjectInfoResponse, error) {
	projectInfo := server.sysInfoMgr.GetProjectInfo()
	response := &pbsysinfo.ProjectInfoResponse{
		ProjectInfo: protobuf.ToProtoProjectInfo(projectInfo),
	}
	return response, nil
}
