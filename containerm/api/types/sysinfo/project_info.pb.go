// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/types/sysinfo/project_info.proto

package sysinfo

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Represents the atomic version of the Containerm instance
type ProjectInfo struct {
	ProjectVersion       string   `protobuf:"bytes,1,opt,name=project_version,json=projectVersion,proto3" json:"project_version,omitempty"`
	BuildTime            string   `protobuf:"bytes,2,opt,name=build_time,json=buildTime,proto3" json:"build_time,omitempty"`
	ApiVersion           string   `protobuf:"bytes,3,opt,name=api_version,json=apiVersion,proto3" json:"api_version,omitempty"`
	GitCommit            string   `protobuf:"bytes,4,opt,name=git_commit,json=gitCommit,proto3" json:"git_commit,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProjectInfo) Reset()         { *m = ProjectInfo{} }
func (m *ProjectInfo) String() string { return proto.CompactTextString(m) }
func (*ProjectInfo) ProtoMessage()    {}
func (*ProjectInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_ec72001f0307f087, []int{0}
}

func (m *ProjectInfo) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProjectInfo.Unmarshal(m, b)
}
func (m *ProjectInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProjectInfo.Marshal(b, m, deterministic)
}
func (m *ProjectInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProjectInfo.Merge(m, src)
}
func (m *ProjectInfo) XXX_Size() int {
	return xxx_messageInfo_ProjectInfo.Size(m)
}
func (m *ProjectInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_ProjectInfo.DiscardUnknown(m)
}

var xxx_messageInfo_ProjectInfo proto.InternalMessageInfo

func (m *ProjectInfo) GetProjectVersion() string {
	if m != nil {
		return m.ProjectVersion
	}
	return ""
}

func (m *ProjectInfo) GetBuildTime() string {
	if m != nil {
		return m.BuildTime
	}
	return ""
}

func (m *ProjectInfo) GetApiVersion() string {
	if m != nil {
		return m.ApiVersion
	}
	return ""
}

func (m *ProjectInfo) GetGitCommit() string {
	if m != nil {
		return m.GitCommit
	}
	return ""
}

func init() {
	proto.RegisterType((*ProjectInfo)(nil), "github.com.eclipse_kanto.container_management.containerm.api.types.sysinfo.ProjectInfo")
}

func init() {
	proto.RegisterFile("api/types/sysinfo/project_info.proto", fileDescriptor_ec72001f0307f087)
}

var fileDescriptor_ec72001f0307f087 = []byte{
	// 243 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0xb1, 0x4a, 0x04, 0x31,
	0x10, 0x86, 0x59, 0x15, 0xe1, 0x72, 0xa0, 0xb0, 0xd5, 0x36, 0xa2, 0x88, 0xa0, 0xcd, 0x25, 0x85,
	0xa5, 0x9d, 0x56, 0x5a, 0xc9, 0x71, 0x58, 0xd8, 0x84, 0x6c, 0x9c, 0x5b, 0x47, 0x6f, 0x32, 0x61,
	0x33, 0x27, 0xdc, 0x83, 0xf8, 0xbe, 0x62, 0x76, 0xdd, 0x08, 0x57, 0x85, 0x7c, 0xc9, 0xfc, 0xc3,
	0xf7, 0xab, 0x2b, 0x17, 0xd1, 0xc8, 0x2e, 0x42, 0x32, 0x69, 0x97, 0x30, 0xac, 0xd9, 0xc4, 0x9e,
	0x3f, 0xc0, 0x8b, 0xfd, 0xbd, 0xe8, 0xd8, 0xb3, 0x70, 0xfd, 0xd4, 0xa1, 0xbc, 0x6f, 0x5b, 0xed,
	0x99, 0x34, 0xf8, 0x0d, 0xc6, 0x04, 0xf6, 0xd3, 0x05, 0x61, 0xed, 0x39, 0x88, 0xc3, 0x00, 0xbd,
	0x25, 0x17, 0x5c, 0x07, 0x04, 0x41, 0x0a, 0x24, 0xed, 0x22, 0xea, 0x1c, 0xaf, 0xc7, 0xf8, 0xcb,
	0xef, 0x4a, 0xcd, 0x9f, 0x87, 0x15, 0x8f, 0x61, 0xcd, 0xf5, 0xb5, 0x3a, 0xfd, 0xdb, 0xf8, 0x05,
	0x7d, 0x42, 0x0e, 0x4d, 0x75, 0x51, 0xdd, 0xcc, 0x96, 0x27, 0x23, 0x7e, 0x19, 0x68, 0x7d, 0xa6,
	0x54, 0xbb, 0xc5, 0xcd, 0x9b, 0x15, 0x24, 0x68, 0x0e, 0xf2, 0x9f, 0x59, 0x26, 0x2b, 0x24, 0xa8,
	0xcf, 0xd5, 0xdc, 0x45, 0x9c, 0x32, 0x0e, 0xf3, 0xbb, 0x72, 0x11, 0xff, 0xcd, 0x77, 0x28, 0xd6,
	0x33, 0x11, 0x4a, 0x73, 0x34, 0xcc, 0x77, 0x28, 0x0f, 0x19, 0xdc, 0xaf, 0x5e, 0x97, 0xc5, 0xd2,
	0x8c, 0x96, 0x8b, 0x6c, 0x69, 0x26, 0xa1, 0x45, 0xb1, 0x2c, 0x90, 0xcc, 0x5e, 0x89, 0x77, 0xe3,
	0xd9, 0x1e, 0xe7, 0x02, 0x6f, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff, 0xc7, 0x86, 0xb6, 0xa4, 0x68,
	0x01, 0x00, 0x00,
}
