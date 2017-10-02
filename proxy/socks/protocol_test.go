package socks_test

import (
	"testing"

	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/common/protocol"
	. "github.com/whatedcgveg/v2ray-core/proxy/socks"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestUDPEncoding(t *testing.T) {
	assert := assert.On(t)

	b := buf.New()

	request := &protocol.RequestHeader{
		Address: net.IPAddress([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}),
		Port:    1024,
	}
	writer := buf.NewSequentialWriter(NewUDPWriter(request, b))

	content := []byte{'a'}
	payload := buf.New()
	payload.Append(content)
	assert.Error(writer.Write(buf.NewMultiBufferValue(payload))).IsNil()

	reader := NewUDPReader(b)

	decodedPayload, err := reader.Read()
	assert.Error(err).IsNil()
	assert.Bytes(decodedPayload[0].Bytes()).Equals(content)
}
