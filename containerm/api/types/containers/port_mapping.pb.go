// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/types/containers/port_mapping.proto

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

// Represents port mapping from the host to a container
type PortMapping struct {
	// Protocol
	Protocol string `protobuf:"bytes,1,opt,name=protocol,proto3" json:"protocol,omitempty"`
	// Host IP
	HostIp string `protobuf:"bytes,2,opt,name=host_ip,json=hostIp,proto3" json:"host_ip,omitempty"`
	// Host port
	HostPort int64 `protobuf:"varint,3,opt,name=host_port,json=hostPort,proto3" json:"host_port,omitempty"`
	// Host port range end
	HostPortEnd int64 `protobuf:"varint,4,opt,name=host_port_end,json=hostPortEnd,proto3" json:"host_port_end,omitempty"`
	// Container port
	ContainerPort        int64    `protobuf:"varint,5,opt,name=container_port,json=containerPort,proto3" json:"container_port,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PortMapping) Reset()         { *m = PortMapping{} }
func (m *PortMapping) String() string { return proto.CompactTextString(m) }
func (*PortMapping) ProtoMessage()    {}
func (*PortMapping) Descriptor() ([]byte, []int) {
	return fileDescriptor_5f76b98b8db741cc, []int{0}
}

func (m *PortMapping) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PortMapping.Unmarshal(m, b)
}
func (m *PortMapping) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PortMapping.Marshal(b, m, deterministic)
}
func (m *PortMapping) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PortMapping.Merge(m, src)
}
func (m *PortMapping) XXX_Size() int {
	return xxx_messageInfo_PortMapping.Size(m)
}
func (m *PortMapping) XXX_DiscardUnknown() {
	xxx_messageInfo_PortMapping.DiscardUnknown(m)
}

var xxx_messageInfo_PortMapping proto.InternalMessageInfo

func (m *PortMapping) GetProtocol() string {
	if m != nil {
		return m.Protocol
	}
	return ""
}

func (m *PortMapping) GetHostIp() string {
	if m != nil {
		return m.HostIp
	}
	return ""
}

func (m *PortMapping) GetHostPort() int64 {
	if m != nil {
		return m.HostPort
	}
	return 0
}

func (m *PortMapping) GetHostPortEnd() int64 {
	if m != nil {
		return m.HostPortEnd
	}
	return 0
}

func (m *PortMapping) GetContainerPort() int64 {
	if m != nil {
		return m.ContainerPort
	}
	return 0
}

func init() {
	proto.RegisterType((*PortMapping)(nil), "github.com.eclipse_kanto.container_management.containerm.api.types.containers.PortMapping")
}

func init() {
	proto.RegisterFile("api/types/containers/port_mapping.proto", fileDescriptor_5f76b98b8db741cc)
}

var fileDescriptor_5f76b98b8db741cc = []byte{
	// 250 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x90, 0x31, 0x4f, 0x03, 0x31,
	0x0c, 0x85, 0x75, 0x14, 0x4a, 0x9b, 0xaa, 0x0c, 0x59, 0x38, 0xc1, 0x52, 0x55, 0x42, 0x74, 0x69,
	0x32, 0x30, 0xb2, 0x21, 0x31, 0x30, 0x54, 0x42, 0x9d, 0x50, 0x97, 0x28, 0xbd, 0x46, 0xd7, 0x88,
	0x26, 0xb6, 0x2e, 0x66, 0xe0, 0x2f, 0xf1, 0x2b, 0x51, 0x0c, 0x24, 0x0b, 0x5b, 0xfc, 0xfc, 0xe2,
	0xe7, 0xcf, 0xe2, 0xde, 0xa2, 0xd7, 0xf4, 0x89, 0x2e, 0xe9, 0x0e, 0x22, 0x59, 0x1f, 0xdd, 0x90,
	0x34, 0xc2, 0x40, 0x26, 0x58, 0x44, 0x1f, 0x7b, 0x85, 0x03, 0x10, 0xc8, 0x4d, 0xef, 0xe9, 0xf8,
	0xb1, 0x57, 0x1d, 0x04, 0xe5, 0xba, 0x93, 0xc7, 0xe4, 0xcc, 0xbb, 0x8d, 0x04, 0xaa, 0xfc, 0x33,
	0xc1, 0x46, 0xdb, 0xbb, 0xe0, 0x22, 0x55, 0x31, 0x28, 0x8b, 0x5e, 0x71, 0x42, 0x15, 0xd3, 0xf2,
	0xab, 0x11, 0xb3, 0x57, 0x18, 0x68, 0xf3, 0x13, 0x22, 0x6f, 0xc4, 0x84, 0x73, 0x3a, 0x38, 0xb5,
	0xcd, 0xa2, 0x59, 0x4d, 0xb7, 0xa5, 0x96, 0xd7, 0xe2, 0xf2, 0x08, 0x89, 0x8c, 0xc7, 0xf6, 0x8c,
	0x5b, 0xe3, 0x5c, 0xbe, 0xa0, 0xbc, 0x15, 0x53, 0x6e, 0xe4, 0x75, 0xdb, 0xd1, 0xa2, 0x59, 0x8d,
	0xb6, 0x93, 0x2c, 0xe4, 0xc1, 0x72, 0x29, 0xe6, 0xa5, 0x69, 0x5c, 0x3c, 0xb4, 0xe7, 0x6c, 0x98,
	0xfd, 0x19, 0x9e, 0xe3, 0x41, 0xde, 0x89, 0xab, 0xba, 0x3d, 0x4f, 0xb9, 0x60, 0xd3, 0xbc, 0xa8,
	0xd9, 0xf9, 0xb4, 0xdb, 0xbd, 0x55, 0x7a, 0xfd, 0x4b, 0xbf, 0x66, 0xfa, 0x7a, 0xb5, 0x75, 0xa5,
	0xaf, 0x62, 0xd0, 0xff, 0xdd, 0xf7, 0xb1, 0x3e, 0xf7, 0x63, 0xc6, 0x7c, 0xf8, 0x0e, 0x00, 0x00,
	0xff, 0xff, 0x36, 0x65, 0xda, 0x64, 0x89, 0x01, 0x00, 0x00,
}
