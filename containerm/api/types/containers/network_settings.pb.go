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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.22.0
// source: api/types/containers/network_settings.proto

package containers

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Represents the network settings of a container
type NetworkSettings struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// A map - network id to endpoint settings for all the joined networks
	Networks map[string]*EndpointSettings `protobuf:"bytes,1,rep,name=networks,proto3" json:"networks,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// The container's network sandbox id
	SandboxId string `protobuf:"bytes,2,opt,name=sandbox_id,json=sandboxId,proto3" json:"sandbox_id,omitempty"`
	// The container's network sandbox key
	SandboxKey string `protobuf:"bytes,3,opt,name=sandbox_key,json=sandboxKey,proto3" json:"sandbox_key,omitempty"`
	// The container's network controller id
	NetworkControllerId string `protobuf:"bytes,4,opt,name=network_controller_id,json=networkControllerId,proto3" json:"network_controller_id,omitempty"`
}

func (x *NetworkSettings) Reset() {
	*x = NetworkSettings{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_types_containers_network_settings_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NetworkSettings) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkSettings) ProtoMessage() {}

func (x *NetworkSettings) ProtoReflect() protoreflect.Message {
	mi := &file_api_types_containers_network_settings_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NetworkSettings.ProtoReflect.Descriptor instead.
func (*NetworkSettings) Descriptor() ([]byte, []int) {
	return file_api_types_containers_network_settings_proto_rawDescGZIP(), []int{0}
}

func (x *NetworkSettings) GetNetworks() map[string]*EndpointSettings {
	if x != nil {
		return x.Networks
	}
	return nil
}

func (x *NetworkSettings) GetSandboxId() string {
	if x != nil {
		return x.SandboxId
	}
	return ""
}

func (x *NetworkSettings) GetSandboxKey() string {
	if x != nil {
		return x.SandboxKey
	}
	return ""
}

func (x *NetworkSettings) GetNetworkControllerId() string {
	if x != nil {
		return x.NetworkControllerId
	}
	return ""
}

var File_api_types_containers_network_settings_proto protoreflect.FileDescriptor

var file_api_types_containers_network_settings_proto_rawDesc = []byte{
	0x0a, 0x2b, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x2f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x73,
	0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x4d, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x65, 0x63, 0x6c, 0x69, 0x70, 0x73,
	0x65, 0x5f, 0x6b, 0x61, 0x6e, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65,
	0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x63, 0x6f, 0x6e,
	0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x6d, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x74, 0x79, 0x70, 0x65,
	0x73, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x1a, 0x2c, 0x61, 0x70,
	0x69, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65,
	0x72, 0x73, 0x2f, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x5f, 0x73, 0x65, 0x74, 0x74,
	0x69, 0x6e, 0x67, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xaf, 0x03, 0x0a, 0x0f, 0x4e,
	0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x12, 0x88,
	0x01, 0x0a, 0x08, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x6c, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x65,
	0x63, 0x6c, 0x69, 0x70, 0x73, 0x65, 0x5f, 0x6b, 0x61, 0x6e, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e,
	0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x6d, 0x65, 0x6e,
	0x74, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x6d, 0x2e, 0x61, 0x70, 0x69,
	0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72,
	0x73, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67,
	0x73, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52,
	0x08, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x12, 0x1d, 0x0a, 0x0a, 0x73, 0x61, 0x6e,
	0x64, 0x62, 0x6f, 0x78, 0x5f, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x09, 0x73,
	0x61, 0x6e, 0x64, 0x62, 0x6f, 0x78, 0x49, 0x64, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x61, 0x6e, 0x64,
	0x62, 0x6f, 0x78, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x73,
	0x61, 0x6e, 0x64, 0x62, 0x6f, 0x78, 0x4b, 0x65, 0x79, 0x12, 0x32, 0x0a, 0x15, 0x6e, 0x65, 0x74,
	0x77, 0x6f, 0x72, 0x6b, 0x5f, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x5f,
	0x69, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x13, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72,
	0x6b, 0x43, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c, 0x6c, 0x65, 0x72, 0x49, 0x64, 0x1a, 0x9c, 0x01,
	0x0a, 0x0d, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x73, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x75, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x5f, 0x2e, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2e, 0x65, 0x63,
	0x6c, 0x69, 0x70, 0x73, 0x65, 0x5f, 0x6b, 0x61, 0x6e, 0x74, 0x6f, 0x2e, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x6d, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73,
	0x2e, 0x45, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x53, 0x65, 0x74, 0x74, 0x69, 0x6e, 0x67,
	0x73, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x5a, 0x5a, 0x58,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x63, 0x6c, 0x69, 0x70,
	0x73, 0x65, 0x2d, 0x6b, 0x61, 0x6e, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e,
	0x65, 0x72, 0x2d, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x6d, 0x65, 0x6e, 0x74, 0x2f, 0x63, 0x6f,
	0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x79, 0x70,
	0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x3b, 0x63, 0x6f,
	0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_api_types_containers_network_settings_proto_rawDescOnce sync.Once
	file_api_types_containers_network_settings_proto_rawDescData = file_api_types_containers_network_settings_proto_rawDesc
)

func file_api_types_containers_network_settings_proto_rawDescGZIP() []byte {
	file_api_types_containers_network_settings_proto_rawDescOnce.Do(func() {
		file_api_types_containers_network_settings_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_types_containers_network_settings_proto_rawDescData)
	})
	return file_api_types_containers_network_settings_proto_rawDescData
}

var file_api_types_containers_network_settings_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_api_types_containers_network_settings_proto_goTypes = []interface{}{
	(*NetworkSettings)(nil),  // 0: github.com.eclipse_kanto.container_management.containerm.api.types.containers.NetworkSettings
	nil,                      // 1: github.com.eclipse_kanto.container_management.containerm.api.types.containers.NetworkSettings.NetworksEntry
	(*EndpointSettings)(nil), // 2: github.com.eclipse_kanto.container_management.containerm.api.types.containers.EndpointSettings
}
var file_api_types_containers_network_settings_proto_depIdxs = []int32{
	1, // 0: github.com.eclipse_kanto.container_management.containerm.api.types.containers.NetworkSettings.networks:type_name -> github.com.eclipse_kanto.container_management.containerm.api.types.containers.NetworkSettings.NetworksEntry
	2, // 1: github.com.eclipse_kanto.container_management.containerm.api.types.containers.NetworkSettings.NetworksEntry.value:type_name -> github.com.eclipse_kanto.container_management.containerm.api.types.containers.EndpointSettings
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_api_types_containers_network_settings_proto_init() }
func file_api_types_containers_network_settings_proto_init() {
	if File_api_types_containers_network_settings_proto != nil {
		return
	}
	file_api_types_containers_endpoint_settings_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_api_types_containers_network_settings_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NetworkSettings); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_api_types_containers_network_settings_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_types_containers_network_settings_proto_goTypes,
		DependencyIndexes: file_api_types_containers_network_settings_proto_depIdxs,
		MessageInfos:      file_api_types_containers_network_settings_proto_msgTypes,
	}.Build()
	File_api_types_containers_network_settings_proto = out.File
	file_api_types_containers_network_settings_proto_rawDesc = nil
	file_api_types_containers_network_settings_proto_goTypes = nil
	file_api_types_containers_network_settings_proto_depIdxs = nil
}
