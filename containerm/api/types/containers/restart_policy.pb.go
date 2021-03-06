// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/types/containers/restart_policy.proto

package containers

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

// Represents the containers restart policy
type RestartPolicy struct {
	// maximum retry count
	MaximumRetryCount int64 `protobuf:"varint,1,opt,name=maximum_retry_count,json=maximumRetryCount,proto3" json:"maximum_retry_count,omitempty"`
	// retry timeout
	RetryTimeout int64 `protobuf:"varint,2,opt,name=retry_timeout,json=retryTimeout,proto3" json:"retry_timeout,omitempty"`
	// type - always, no, on-failure, unless-stopped
	Type                 string   `protobuf:"bytes,3,opt,name=type,proto3" json:"type,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *RestartPolicy) Reset()         { *m = RestartPolicy{} }
func (m *RestartPolicy) String() string { return proto.CompactTextString(m) }
func (*RestartPolicy) ProtoMessage()    {}
func (*RestartPolicy) Descriptor() ([]byte, []int) {
	return fileDescriptor_fc6a2c23b23c6d01, []int{0}
}

func (m *RestartPolicy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RestartPolicy.Unmarshal(m, b)
}
func (m *RestartPolicy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RestartPolicy.Marshal(b, m, deterministic)
}
func (m *RestartPolicy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RestartPolicy.Merge(m, src)
}
func (m *RestartPolicy) XXX_Size() int {
	return xxx_messageInfo_RestartPolicy.Size(m)
}
func (m *RestartPolicy) XXX_DiscardUnknown() {
	xxx_messageInfo_RestartPolicy.DiscardUnknown(m)
}

var xxx_messageInfo_RestartPolicy proto.InternalMessageInfo

func (m *RestartPolicy) GetMaximumRetryCount() int64 {
	if m != nil {
		return m.MaximumRetryCount
	}
	return 0
}

func (m *RestartPolicy) GetRetryTimeout() int64 {
	if m != nil {
		return m.RetryTimeout
	}
	return 0
}

func (m *RestartPolicy) GetType() string {
	if m != nil {
		return m.Type
	}
	return ""
}

func init() {
	proto.RegisterType((*RestartPolicy)(nil), "github.com.eclipse_kanto.container_management.containerm.api.types.containers.RestartPolicy")
}

func init() {
	proto.RegisterFile("api/types/containers/restart_policy.proto", fileDescriptor_fc6a2c23b23c6d01)
}

var fileDescriptor_fc6a2c23b23c6d01 = []byte{
	// 228 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x31, 0x4b, 0x04, 0x31,
	0x10, 0x85, 0x59, 0x4f, 0x04, 0x83, 0x57, 0x18, 0x9b, 0x2d, 0x0f, 0x6d, 0xce, 0xe2, 0x92, 0xc2,
	0xd2, 0x4e, 0x6b, 0x41, 0x16, 0x0b, 0xb9, 0x26, 0xe4, 0xc2, 0x70, 0x06, 0x6f, 0x32, 0x21, 0x99,
	0xc0, 0xed, 0xbf, 0x97, 0xcd, 0x2e, 0xa6, 0xb9, 0x6e, 0xf8, 0xde, 0x1b, 0x1e, 0xef, 0x89, 0x67,
	0x1b, 0xbd, 0xe6, 0x31, 0x42, 0xd6, 0x8e, 0x02, 0x5b, 0x1f, 0x20, 0x65, 0x9d, 0x20, 0xb3, 0x4d,
	0x6c, 0x22, 0x9d, 0xbc, 0x1b, 0x55, 0x4c, 0xc4, 0x24, 0x3f, 0x8e, 0x9e, 0x7f, 0xca, 0x41, 0x39,
	0x42, 0x05, 0xee, 0xe4, 0x63, 0x06, 0xf3, 0x6b, 0x03, 0x93, 0xfa, 0xff, 0x34, 0x68, 0x83, 0x3d,
	0x02, 0x42, 0xe0, 0x06, 0x51, 0xd9, 0xe8, 0x55, 0xcd, 0x68, 0x30, 0x3f, 0x9e, 0xc5, 0x7a, 0x98,
	0x63, 0x3e, 0x6b, 0x8a, 0x54, 0xe2, 0x01, 0xed, 0xd9, 0x63, 0x41, 0x93, 0x80, 0xd3, 0x68, 0x1c,
	0x95, 0xc0, 0x7d, 0xb7, 0xe9, 0xb6, 0xab, 0xe1, 0x7e, 0x91, 0x86, 0x49, 0x79, 0x9f, 0x04, 0xf9,
	0x24, 0xd6, 0xb3, 0x8f, 0x3d, 0x02, 0x15, 0xee, 0xaf, 0xaa, 0xf3, 0xae, 0xc2, 0xaf, 0x99, 0x49,
	0x29, 0xae, 0xa7, 0xe4, 0x7e, 0xb5, 0xe9, 0xb6, 0xb7, 0x43, 0xbd, 0xdf, 0xf6, 0xfb, 0xef, 0x56,
	0x45, 0x2f, 0x55, 0x76, 0xb5, 0x4a, 0x1b, 0x61, 0xd7, 0xaa, 0x34, 0x88, 0xfa, 0xd2, 0x5c, 0xaf,
	0xed, 0x3c, 0xdc, 0xd4, 0xad, 0x5e, 0xfe, 0x02, 0x00, 0x00, 0xff, 0xff, 0x1d, 0xd2, 0xb3, 0x3f,
	0x58, 0x01, 0x00, 0x00,
}
