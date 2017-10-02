package internet_test

import (
	"context"
	"testing"

	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
	"github.com/whatedcgveg/v2ray-core/testing/servers/tcp"
	. "github.com/whatedcgveg/v2ray-core/transport/internet"
)

func TestDialWithLocalAddr(t *testing.T) {
	assert := assert.On(t)

	server := &tcp.Server{}
	dest, err := server.Start()
	assert.Error(err).IsNil()
	defer server.Close()

	conn, err := DialSystem(context.Background(), net.LocalHostIP, net.TCPDestination(net.LocalHostIP, dest.Port))
	assert.Error(err).IsNil()
	assert.String(conn.RemoteAddr().String()).Equals("127.0.0.1:" + dest.Port.String())
	conn.Close()
}
