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
// source: api/types/containers/io_config.proto

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

// IO configuration contains the streams to be attached to this container
type IOConfig struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Whether to attach to `stderr`.
	AttachStderr bool `protobuf:"varint,1,opt,name=attach_stderr,json=attachStderr,proto3" json:"attach_stderr,omitempty"`
	// Whether to attach to `stdin`.
	AttachStdin bool `protobuf:"varint,2,opt,name=attach_stdin,json=attachStdin,proto3" json:"attach_stdin,omitempty"`
	// Whether to attach to `stdout`.
	AttachStdout bool `protobuf:"varint,3,opt,name=attach_stdout,json=attachStdout,proto3" json:"attach_stdout,omitempty"`
	// Open `stdin`
	OpenStdin bool `protobuf:"varint,4,opt,name=open_stdin,json=openStdin,proto3" json:"open_stdin,omitempty"`
	// Close `stdin` after one attached client disconnects
	StdinOnce bool `protobuf:"varint,5,opt,name=stdin_once,json=stdinOnce,proto3" json:"stdin_once,omitempty"`
	// Attach standard streams to a TTY, including `stdin` if it is not closed.
	Tty bool `protobuf:"varint,6,opt,name=tty,proto3" json:"tty,omitempty"`
}

func (x *IOConfig) Reset() {
	*x = IOConfig{}
	if protoimpl.UnsafeEnabled {
		mi := &file_api_types_containers_io_config_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IOConfig) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IOConfig) ProtoMessage() {}

func (x *IOConfig) ProtoReflect() protoreflect.Message {
	mi := &file_api_types_containers_io_config_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IOConfig.ProtoReflect.Descriptor instead.
func (*IOConfig) Descriptor() ([]byte, []int) {
	return file_api_types_containers_io_config_proto_rawDescGZIP(), []int{0}
}

func (x *IOConfig) GetAttachStderr() bool {
	if x != nil {
		return x.AttachStderr
	}
	return false
}

func (x *IOConfig) GetAttachStdin() bool {
	if x != nil {
		return x.AttachStdin
	}
	return false
}

func (x *IOConfig) GetAttachStdout() bool {
	if x != nil {
		return x.AttachStdout
	}
	return false
}

func (x *IOConfig) GetOpenStdin() bool {
	if x != nil {
		return x.OpenStdin
	}
	return false
}

func (x *IOConfig) GetStdinOnce() bool {
	if x != nil {
		return x.StdinOnce
	}
	return false
}

func (x *IOConfig) GetTty() bool {
	if x != nil {
		return x.Tty
	}
	return false
}

var File_api_types_containers_io_config_proto protoreflect.FileDescriptor

var file_api_types_containers_io_config_proto_rawDesc = []byte{
	0x0a, 0x24, 0x61, 0x70, 0x69, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x2f, 0x69, 0x6f, 0x5f, 0x63, 0x6f, 0x6e, 0x66, 0x69, 0x67,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x4d, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2e, 0x65, 0x63, 0x6c, 0x69, 0x70, 0x73, 0x65, 0x5f, 0x6b, 0x61, 0x6e, 0x74, 0x6f,
	0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x5f, 0x6d, 0x61, 0x6e, 0x61, 0x67,
	0x65, 0x6d, 0x65, 0x6e, 0x74, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x6d,
	0x2e, 0x61, 0x70, 0x69, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x63, 0x6f, 0x6e, 0x74, 0x61,
	0x69, 0x6e, 0x65, 0x72, 0x73, 0x22, 0xc7, 0x01, 0x0a, 0x08, 0x49, 0x4f, 0x43, 0x6f, 0x6e, 0x66,
	0x69, 0x67, 0x12, 0x23, 0x0a, 0x0d, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x5f, 0x73, 0x74, 0x64,
	0x65, 0x72, 0x72, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0c, 0x61, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x53, 0x74, 0x64, 0x65, 0x72, 0x72, 0x12, 0x21, 0x0a, 0x0c, 0x61, 0x74, 0x74, 0x61, 0x63,
	0x68, 0x5f, 0x73, 0x74, 0x64, 0x69, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x61,
	0x74, 0x74, 0x61, 0x63, 0x68, 0x53, 0x74, 0x64, 0x69, 0x6e, 0x12, 0x23, 0x0a, 0x0d, 0x61, 0x74,
	0x74, 0x61, 0x63, 0x68, 0x5f, 0x73, 0x74, 0x64, 0x6f, 0x75, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x08, 0x52, 0x0c, 0x61, 0x74, 0x74, 0x61, 0x63, 0x68, 0x53, 0x74, 0x64, 0x6f, 0x75, 0x74, 0x12,
	0x1d, 0x0a, 0x0a, 0x6f, 0x70, 0x65, 0x6e, 0x5f, 0x73, 0x74, 0x64, 0x69, 0x6e, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x09, 0x6f, 0x70, 0x65, 0x6e, 0x53, 0x74, 0x64, 0x69, 0x6e, 0x12, 0x1d,
	0x0a, 0x0a, 0x73, 0x74, 0x64, 0x69, 0x6e, 0x5f, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x08, 0x52, 0x09, 0x73, 0x74, 0x64, 0x69, 0x6e, 0x4f, 0x6e, 0x63, 0x65, 0x12, 0x10, 0x0a,
	0x03, 0x74, 0x74, 0x79, 0x18, 0x06, 0x20, 0x01, 0x28, 0x08, 0x52, 0x03, 0x74, 0x74, 0x79, 0x42,
	0x5a, 0x5a, 0x58, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x63,
	0x6c, 0x69, 0x70, 0x73, 0x65, 0x2d, 0x6b, 0x61, 0x6e, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6e, 0x74,
	0x61, 0x69, 0x6e, 0x65, 0x72, 0x2d, 0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x6d, 0x65, 0x6e, 0x74,
	0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x6d, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73,
	0x3b, 0x63, 0x6f, 0x6e, 0x74, 0x61, 0x69, 0x6e, 0x65, 0x72, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x33,
}

var (
	file_api_types_containers_io_config_proto_rawDescOnce sync.Once
	file_api_types_containers_io_config_proto_rawDescData = file_api_types_containers_io_config_proto_rawDesc
)

func file_api_types_containers_io_config_proto_rawDescGZIP() []byte {
	file_api_types_containers_io_config_proto_rawDescOnce.Do(func() {
		file_api_types_containers_io_config_proto_rawDescData = protoimpl.X.CompressGZIP(file_api_types_containers_io_config_proto_rawDescData)
	})
	return file_api_types_containers_io_config_proto_rawDescData
}

var file_api_types_containers_io_config_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_api_types_containers_io_config_proto_goTypes = []interface{}{
	(*IOConfig)(nil), // 0: github.com.eclipse_kanto.container_management.containerm.api.types.containers.IOConfig
}
var file_api_types_containers_io_config_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_api_types_containers_io_config_proto_init() }
func file_api_types_containers_io_config_proto_init() {
	if File_api_types_containers_io_config_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_api_types_containers_io_config_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IOConfig); i {
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
			RawDescriptor: file_api_types_containers_io_config_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_api_types_containers_io_config_proto_goTypes,
		DependencyIndexes: file_api_types_containers_io_config_proto_depIdxs,
		MessageInfos:      file_api_types_containers_io_config_proto_msgTypes,
	}.Build()
	File_api_types_containers_io_config_proto = out.File
	file_api_types_containers_io_config_proto_rawDesc = nil
	file_api_types_containers_io_config_proto_goTypes = nil
	file_api_types_containers_io_config_proto_depIdxs = nil
}
