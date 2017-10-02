package kcp_test

import (
	"context"
	"crypto/rand"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
	"github.com/whatedcgveg/v2ray-core/transport/internet"
	. "github.com/whatedcgveg/v2ray-core/transport/internet/kcp"
)

func TestDialAndListen(t *testing.T) {
	assert := assert.On(t)

	listerner, err := NewListener(internet.ContextWithTransportSettings(context.Background(), &Config{}), net.LocalHostIP, net.Port(0), func(ctx context.Context, conn internet.Connection) bool {
		go func(c internet.Connection) {
			payload := make([]byte, 4096)
			for {
				nBytes, err := c.Read(payload)
				if err != nil {
					break
				}
				for idx, b := range payload[:nBytes] {
					payload[idx] = b ^ 'c'
				}
				c.Write(payload[:nBytes])
			}
			c.Close()
		}(conn)
		return true
	})
	assert.Error(err).IsNil()
	port := net.Port(listerner.Addr().(*net.UDPAddr).Port)

	ctx := internet.ContextWithTransportSettings(context.Background(), &Config{})
	wg := new(sync.WaitGroup)
	for i := 0; i < 10; i++ {
		clientConn, err := DialKCP(ctx, net.UDPDestination(net.LocalHostIP, port))
		assert.Error(err).IsNil()
		wg.Add(1)

		go func() {
			clientSend := make([]byte, 1024*1024)
			rand.Read(clientSend)
			go clientConn.Write(clientSend)

			clientReceived := make([]byte, 1024*1024)
			nBytes, _ := io.ReadFull(clientConn, clientReceived)
			assert.Int(nBytes).Equals(len(clientReceived))
			clientConn.Close()

			clientExpected := make([]byte, 1024*1024)
			for idx, b := range clientSend {
				clientExpected[idx] = b ^ 'c'
			}
			assert.Bytes(clientReceived).Equals(clientExpected)

			wg.Done()
		}()
	}

	wg.Wait()
	for i := 0; i < 60 && listerner.ActiveConnections() > 0; i++ {
		time.Sleep(500 * time.Millisecond)
	}
	assert.Int(listerner.ActiveConnections()).Equals(0)

	listerner.Close()
}
