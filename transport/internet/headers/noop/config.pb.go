package noop

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Config struct {
}

func (m *Config) Reset()                    { *m = Config{} }
func (m *Config) String() string            { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()               {}
func (*Config) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type ConnectionConfig struct {
}

func (m *ConnectionConfig) Reset()                    { *m = ConnectionConfig{} }
func (m *ConnectionConfig) String() string            { return proto.CompactTextString(m) }
func (*ConnectionConfig) ProtoMessage()               {}
func (*ConnectionConfig) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func init() {
	proto.RegisterType((*Config)(nil), "v2ray.core.transport.internet.headers.noop.Config")
	proto.RegisterType((*ConnectionConfig)(nil), "v2ray.core.transport.internet.headers.noop.ConnectionConfig")
}

func init() {
	proto.RegisterFile("github.com/whatedcgveg/v2ray-core/transport/internet/headers/noop/config.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 170 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0xce, 0xb1, 0xaa, 0xc2, 0x40,
	0x10, 0x85, 0x61, 0xee, 0x45, 0x82, 0x6c, 0x25, 0x79, 0x84, 0x94, 0x29, 0x66, 0x21, 0x96, 0x76,
	0xa6, 0xd1, 0x46, 0x44, 0xc4, 0xc2, 0x6e, 0x5d, 0x47, 0x4d, 0xe1, 0x9c, 0x65, 0xb2, 0x08, 0x79,
	0x25, 0x9f, 0x52, 0x36, 0x26, 0xa9, 0xad, 0x06, 0x06, 0xbe, 0xc3, 0x6f, 0x56, 0xaf, 0x4a, 0x5d,
	0x47, 0x1e, 0x4f, 0xeb, 0xa1, 0x6c, 0xa3, 0x3a, 0x69, 0x03, 0x34, 0xda, 0x46, 0x22, 0xab, 0x70,
	0xb4, 0x0f, 0x76, 0x57, 0xd6, 0xd6, 0x0a, 0x10, 0xac, 0x87, 0xdc, 0x9a, 0x3b, 0x05, 0x45, 0x44,
	0x5e, 0x8e, 0x58, 0x99, 0x26, 0x48, 0x23, 0xa4, 0x01, 0x52, 0x82, 0xc5, 0xdc, 0x64, 0x75, 0x6f,
	0x8b, 0xdc, 0x2c, 0x6a, 0x88, 0xb0, 0x8f, 0x0d, 0xe4, 0xfb, 0x5b, 0xb3, 0x49, 0x09, 0xf4, 0xfb,
	0xde, 0xfe, 0xef, 0x3c, 0x4b, 0xf7, 0xfd, 0x5f, 0x9e, 0xaa, 0x83, 0xeb, 0xa8, 0x4e, 0xe8, 0x38,
	0xa1, 0xed, 0x88, 0x36, 0x03, 0xda, 0x01, 0xe1, 0x92, 0xf5, 0xdd, 0xcb, 0x4f, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xd0, 0xe6, 0xd7, 0x87, 0xf6, 0x00, 0x00, 0x00,
}
