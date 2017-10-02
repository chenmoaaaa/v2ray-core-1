package core_test

import (
	"testing"

	. "github.com/whatedcgveg/v2ray-core"
	"github.com/whatedcgveg/v2ray-core/app/proxyman"
	"github.com/whatedcgveg/v2ray-core/common/dice"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/common/protocol"
	"github.com/whatedcgveg/v2ray-core/common/serial"
	"github.com/whatedcgveg/v2ray-core/common/uuid"
	_ "github.com/whatedcgveg/v2ray-core/main/distro/all"
	"github.com/whatedcgveg/v2ray-core/proxy/dokodemo"
	"github.com/whatedcgveg/v2ray-core/proxy/vmess"
	"github.com/whatedcgveg/v2ray-core/proxy/vmess/outbound"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestV2RayClose(t *testing.T) {
	assert := assert.On(t)

	port := net.Port(dice.RollUint16())
	config := &Config{
		Inbound: []*proxyman.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(port),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address: net.NewIPOrDomain(net.LocalHostIP),
					Port:    uint32(0),
					NetworkList: &net.NetworkList{
						Network: []net.Network{net.Network_TCP, net.Network_UDP},
					},
				}),
			},
		},
		Outbound: []*proxyman.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&outbound.Config{
					Receiver: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(0),
							User: []*protocol.User{
								{
									Account: serial.ToTypedMessage(&vmess.Account{
										Id: uuid.New().String(),
									}),
								},
							},
						},
					},
				}),
			},
		},
	}

	server, err := New(config)
	assert.Error(err).IsNil()

	server.Close()
}
