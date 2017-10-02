package scenarios

import (
	"fmt"
	"testing"

	xproxy "golang.org/x/net/proxy"
	"github.com/whatedcgveg/v2ray-core"
	"github.com/whatedcgveg/v2ray-core/app/dns"
	"github.com/whatedcgveg/v2ray-core/app/proxyman"
	"github.com/whatedcgveg/v2ray-core/app/router"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/common/serial"
	"github.com/whatedcgveg/v2ray-core/proxy/blackhole"
	"github.com/whatedcgveg/v2ray-core/proxy/freedom"
	"github.com/whatedcgveg/v2ray-core/proxy/socks"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
	"github.com/whatedcgveg/v2ray-core/testing/servers/tcp"
)

func TestResolveIP(t *testing.T) {
	assert := assert.On(t)

	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	assert.Error(err).IsNil()
	defer tcpServer.Close()

	serverPort := pickPort()
	serverConfig := &core.Config{
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&dns.Config{
				Hosts: map[string]*net.IPOrDomain{
					"google.com": net.NewIPOrDomain(dest.Address),
				},
			}),
			serial.ToTypedMessage(&router.Config{
				DomainStrategy: router.Config_IpIfNonMatch,
				Rule: []*router.RoutingRule{
					{
						Cidr: []*router.CIDR{
							{
								Ip:     []byte{127, 0, 0, 0},
								Prefix: 8,
							},
						},
						Tag: "direct",
					},
				},
			}),
		},
		Inbound: []*proxyman.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&socks.ServerConfig{
					AuthType: socks.AuthType_NO_AUTH,
					Accounts: map[string]string{
						"Test Account": "Test Password",
					},
					Address:    net.NewIPOrDomain(net.LocalHostIP),
					UdpEnabled: false,
				}),
			},
		},
		Outbound: []*proxyman.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&blackhole.Config{}),
			},
			{
				Tag: "direct",
				ProxySettings: serial.ToTypedMessage(&freedom.Config{
					DomainStrategy: freedom.Config_USE_IP,
				}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	assert.Error(err).IsNil()

	{
		noAuthDialer, err := xproxy.SOCKS5("tcp", net.TCPDestination(net.LocalHostIP, serverPort).NetAddr(), nil, xproxy.Direct)
		assert.Error(err).IsNil()
		conn, err := noAuthDialer.Dial("tcp", fmt.Sprintf("google.com:%d", dest.Port))
		assert.Error(err).IsNil()

		payload := "test payload"
		nBytes, err := conn.Write([]byte(payload))
		assert.Error(err).IsNil()
		assert.Int(nBytes).Equals(len(payload))

		response := make([]byte, 1024)
		nBytes, err = conn.Read(response)
		assert.Error(err).IsNil()
		assert.Bytes(response[:nBytes]).Equals(xor([]byte(payload)))
		assert.Error(conn.Close()).IsNil()
	}

	CloseAllServers(servers)
}
