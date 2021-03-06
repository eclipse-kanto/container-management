// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/types/containers/update_options.proto

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

// UpdateOptions represent options for updating a container.
type UpdateOptions struct {
	// The container's restart policy
	RestartPolicy *RestartPolicy `protobuf:"bytes,1,opt,name=restart_policy,json=restartPolicy,proto3" json:"restart_policy,omitempty"`
	// The container's resource config
	Resources            *Resources `protobuf:"bytes,2,opt,name=resources,proto3" json:"resources,omitempty"`
	XXX_NoUnkeyedLiteral struct{}   `json:"-"`
	XXX_unrecognized     []byte     `json:"-"`
	XXX_sizecache        int32      `json:"-"`
}

func (m *UpdateOptions) Reset()         { *m = UpdateOptions{} }
func (m *UpdateOptions) String() string { return proto.CompactTextString(m) }
func (*UpdateOptions) ProtoMessage()    {}
func (*UpdateOptions) Descriptor() ([]byte, []int) {
	return fileDescriptor_63b4a2996851a2b5, []int{0}
}

func (m *UpdateOptions) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_UpdateOptions.Unmarshal(m, b)
}
func (m *UpdateOptions) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_UpdateOptions.Marshal(b, m, deterministic)
}
func (m *UpdateOptions) XXX_Merge(src proto.Message) {
	xxx_messageInfo_UpdateOptions.Merge(m, src)
}
func (m *UpdateOptions) XXX_Size() int {
	return xxx_messageInfo_UpdateOptions.Size(m)
}
func (m *UpdateOptions) XXX_DiscardUnknown() {
	xxx_messageInfo_UpdateOptions.DiscardUnknown(m)
}

var xxx_messageInfo_UpdateOptions proto.InternalMessageInfo

func (m *UpdateOptions) GetRestartPolicy() *RestartPolicy {
	if m != nil {
		return m.RestartPolicy
	}
	return nil
}

func (m *UpdateOptions) GetResources() *Resources {
	if m != nil {
		return m.Resources
	}
	return nil
}

func init() {
	proto.RegisterType((*UpdateOptions)(nil), "github.com.eclipse_kanto.container_management.containerm.api.types.containers.UpdateOptions")
}

func init() {
	proto.RegisterFile("api/types/containers/update_options.proto", fileDescriptor_63b4a2996851a2b5)
}

var fileDescriptor_63b4a2996851a2b5 = []byte{
	// 242 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0xd0, 0xb1, 0x4b, 0x03, 0x31,
	0x14, 0x06, 0x70, 0xae, 0x83, 0xe0, 0x49, 0x1d, 0x6e, 0x2a, 0x9d, 0x44, 0x1c, 0x74, 0x68, 0x02,
	0x3a, 0xba, 0xb9, 0x8b, 0x72, 0x20, 0x94, 0x22, 0x1c, 0xaf, 0xf1, 0x51, 0x83, 0x4d, 0xde, 0x23,
	0x79, 0x27, 0x74, 0x76, 0xf6, 0x7f, 0x16, 0x72, 0xa5, 0x4f, 0xa1, 0x6e, 0xb7, 0xdd, 0x3d, 0x92,
	0xf7, 0xfb, 0xf2, 0xd5, 0x37, 0xc0, 0xde, 0xca, 0x8e, 0x31, 0x5b, 0x47, 0x51, 0xc0, 0x47, 0x4c,
	0xd9, 0xf6, 0xfc, 0x06, 0x82, 0x1d, 0xb1, 0x78, 0x8a, 0xd9, 0x70, 0x22, 0xa1, 0xe6, 0x71, 0xe3,
	0xe5, 0xbd, 0x5f, 0x1b, 0x47, 0xc1, 0xa0, 0xdb, 0x7a, 0xce, 0xd8, 0x7d, 0x40, 0x14, 0x32, 0x87,
	0x9b, 0x5d, 0x80, 0x08, 0x1b, 0x0c, 0x18, 0x45, 0x87, 0xc1, 0x00, 0x7b, 0x53, 0x0c, 0x1d, 0xe6,
	0xf9, 0x71, 0x39, 0x61, 0x16, 0x48, 0xd2, 0x31, 0x6d, 0xbd, 0xdb, 0x0d, 0xf2, 0xfc, 0xea, 0xbf,
	0xa3, 0xd4, 0x27, 0x87, 0xfb, 0x7c, 0x97, 0xdf, 0x93, 0x7a, 0xfa, 0x52, 0x82, 0x3f, 0x0d, 0xb9,
	0x9b, 0xaf, 0xaa, 0x3e, 0xff, 0xbb, 0x70, 0x56, 0x5d, 0x54, 0xd7, 0x67, 0xb7, 0xaf, 0x66, 0xd4,
	0xb7, 0x98, 0x76, 0x40, 0x9e, 0x8b, 0xd1, 0x4e, 0xd3, 0xef, 0xdf, 0xe6, 0xb3, 0x3e, 0x3d, 0x44,
	0x9d, 0x4d, 0x8a, 0xbf, 0x1c, 0xdf, 0x1f, 0xf6, 0xb7, 0x4a, 0x3d, 0xac, 0x56, 0x4b, 0x55, 0xec,
	0x5e, 0x59, 0x14, 0x45, 0x6b, 0x5c, 0xa8, 0xa2, 0xc3, 0x60, 0x8f, 0x15, 0x7e, 0xaf, 0x9f, 0xeb,
	0x93, 0x52, 0xf9, 0xdd, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x69, 0x53, 0xe8, 0x02, 0x3f, 0x02,
	0x00, 0x00,
}
