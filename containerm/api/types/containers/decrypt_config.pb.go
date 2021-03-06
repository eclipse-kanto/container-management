// Code generated by protoc-gen-go. DO NOT EDIT.
// source: api/types/containers/decrypt_config.proto

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

// Represents container image's decryption configuration
type DecryptConfig struct {
	// Private key filepath with an optional password separated by a colon
	Keys []string `protobuf:"bytes,1,rep,name=keys,proto3" json:"keys,omitempty"`
	// Recipient protocol(pkcs7) and filepath to a x509 certificate separated by a colon
	Recipients           []string `protobuf:"bytes,2,rep,name=recipients,proto3" json:"recipients,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *DecryptConfig) Reset()         { *m = DecryptConfig{} }
func (m *DecryptConfig) String() string { return proto.CompactTextString(m) }
func (*DecryptConfig) ProtoMessage()    {}
func (*DecryptConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_96fd24169c4faeb2, []int{0}
}

func (m *DecryptConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_DecryptConfig.Unmarshal(m, b)
}
func (m *DecryptConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_DecryptConfig.Marshal(b, m, deterministic)
}
func (m *DecryptConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_DecryptConfig.Merge(m, src)
}
func (m *DecryptConfig) XXX_Size() int {
	return xxx_messageInfo_DecryptConfig.Size(m)
}
func (m *DecryptConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_DecryptConfig.DiscardUnknown(m)
}

var xxx_messageInfo_DecryptConfig proto.InternalMessageInfo

func (m *DecryptConfig) GetKeys() []string {
	if m != nil {
		return m.Keys
	}
	return nil
}

func (m *DecryptConfig) GetRecipients() []string {
	if m != nil {
		return m.Recipients
	}
	return nil
}

func init() {
	proto.RegisterType((*DecryptConfig)(nil), "github.com.eclipse_kanto.container_management.containerm.api.types.containers.DecryptConfig")
}

func init() {
	proto.RegisterFile("api/types/containers/decrypt_config.proto", fileDescriptor_96fd24169c4faeb2)
}

var fileDescriptor_96fd24169c4faeb2 = []byte{
	// 192 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x6c, 0x8f, 0xbd, 0x0b, 0xc2, 0x30,
	0x10, 0xc5, 0xf1, 0x03, 0xc1, 0x80, 0x4b, 0xa6, 0x4e, 0x22, 0x4e, 0x3a, 0x34, 0x19, 0x1c, 0xdd,
	0xac, 0xab, 0x8b, 0x93, 0x74, 0x29, 0x69, 0x3c, 0xeb, 0x51, 0xf3, 0x41, 0x72, 0x0e, 0xfd, 0xef,
	0x85, 0x28, 0xc6, 0xc1, 0xed, 0xf1, 0x3b, 0x8e, 0xf7, 0x7b, 0x6c, 0xab, 0x3c, 0x4a, 0x1a, 0x3c,
	0x44, 0xa9, 0x9d, 0x25, 0x85, 0x16, 0x42, 0x94, 0x57, 0xd0, 0x61, 0xf0, 0xd4, 0x68, 0x67, 0x6f,
	0xd8, 0x09, 0x1f, 0x1c, 0x39, 0x7e, 0xea, 0x90, 0xee, 0xcf, 0x56, 0x68, 0x67, 0x04, 0xe8, 0x07,
	0xfa, 0x08, 0x4d, 0xaf, 0x2c, 0x39, 0xf1, 0xfd, 0x6c, 0x8c, 0xb2, 0xaa, 0x03, 0x03, 0x96, 0x32,
	0x34, 0x42, 0x79, 0x14, 0xa9, 0x23, 0xc3, 0xb8, 0xae, 0xd8, 0xe2, 0xf8, 0xae, 0xa9, 0x52, 0x0b,
	0xe7, 0x6c, 0xda, 0xc3, 0x10, 0x8b, 0xd1, 0x6a, 0xb2, 0x99, 0x9f, 0x53, 0xe6, 0x4b, 0xc6, 0x02,
	0x68, 0xf4, 0x08, 0x96, 0x62, 0x31, 0x4e, 0x97, 0x1f, 0x72, 0xa8, 0xeb, 0x4b, 0xb6, 0x92, 0x1f,
	0xab, 0x32, 0x59, 0xe5, 0x3d, 0x65, 0xb6, 0xca, 0xd0, 0xc8, 0x7f, 0xcb, 0xf7, 0x39, 0xb6, 0xb3,
	0x34, 0x7b, 0xf7, 0x0a, 0x00, 0x00, 0xff, 0xff, 0x3e, 0x4e, 0xcb, 0x89, 0x23, 0x01, 0x00, 0x00,
}
