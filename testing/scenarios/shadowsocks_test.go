package scenarios

import (
	"crypto/rand"
	"sync"
	"testing"
	"time"

	"github.com/whatedcgveg/v2ray-core"
	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/app/proxyman"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/common/protocol"
	"github.com/whatedcgveg/v2ray-core/common/serial"
	"github.com/whatedcgveg/v2ray-core/proxy/dokodemo"
	"github.com/whatedcgveg/v2ray-core/proxy/freedom"
	"github.com/whatedcgveg/v2ray-core/proxy/shadowsocks"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
	"github.com/whatedcgveg/v2ray-core/testing/servers/tcp"
	"github.com/whatedcgveg/v2ray-core/testing/servers/udp"
)

func TestShadowsocksAES256TCP(t *testing.T) {
	assert := assert.On(t)

	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	assert.Error(err).IsNil()
	defer tcpServer.Close()

	account := serial.ToTypedMessage(&shadowsocks.Account{
		Password:   "shadowsocks-password",
		CipherType: shadowsocks.CipherType_AES_256_CFB,
		Ota:        shadowsocks.Account_Enabled,
	})

	serverPort := pickPort()
	serverConfig := &core.Config{
		Inbound: []*proxyman.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&shadowsocks.ServerConfig{
					User: &protocol.User{
						Account: account,
						Level:   1,
					},
				}),
			},
		},
		Outbound: []*proxyman.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&log.Config{
				ErrorLogLevel: log.LogLevel_Debug,
				ErrorLogType:  log.LogType_Console,
			}),
		},
	}

	clientPort := pickPort()
	clientConfig := &core.Config{
		Inbound: []*proxyman.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(clientPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address: net.NewIPOrDomain(dest.Address),
					Port:    uint32(dest.Port),
					NetworkList: &net.NetworkList{
						Network: []net.Network{net.Network_TCP},
					},
				}),
			},
		},
		Outbound: []*proxyman.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&shadowsocks.ClientConfig{
					Server: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(serverPort),
							User: []*protocol.User{
								{
									Account: account,
								},
							},
						},
					},
				}),
			},
		},
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&log.Config{
				ErrorLogLevel: log.LogLevel_Debug,
				ErrorLogType:  log.LogType_Console,
			}),
		},
	}

	servers, err := InitializeServerConfigs(serverConfig, clientConfig)
	assert.Error(err).IsNil()

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: int(clientPort),
			})
			assert.Error(err).IsNil()

			payload := make([]byte, 10240*1024)
			rand.Read(payload)

			nBytes, err := conn.Write([]byte(payload))
			assert.Error(err).IsNil()
			assert.Int(nBytes).Equals(len(payload))

			response := readFrom(conn, time.Second*20, 10240*1024)
			assert.Bytes(response).Equals(xor([]byte(payload)))
			assert.Error(conn.Close()).IsNil()
			wg.Done()
		}()
	}
	wg.Wait()

	CloseAllServers(servers)
}

func TestShadowsocksAES128UDP(t *testing.T) {
	assert := assert.On(t)

	udpServer := udp.Server{
		MsgProcessor: xor,
	}
	dest, err := udpServer.Start()
	assert.Error(err).IsNil()
	defer udpServer.Close()

	account := serial.ToTypedMessage(&shadowsocks.Account{
		Password:   "shadowsocks-password",
		CipherType: shadowsocks.CipherType_AES_128_CFB,
		Ota:        shadowsocks.Account_Enabled,
	})

	serverPort := pickPort()
	serverConfig := &core.Config{
		Inbound: []*proxyman.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&shadowsocks.ServerConfig{
					UdpEnabled: true,
					User: &protocol.User{
						Account: account,
						Level:   1,
					},
				}),
			},
		},
		Outbound: []*proxyman.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&log.Config{
				ErrorLogLevel: log.LogLevel_Debug,
				ErrorLogType:  log.LogType_Console,
			}),
		},
	}

	clientPort := pickPort()
	clientConfig := &core.Config{
		Inbound: []*proxyman.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(clientPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address: net.NewIPOrDomain(dest.Address),
					Port:    uint32(dest.Port),
					NetworkList: &net.NetworkList{
						Network: []net.Network{net.Network_UDP},
					},
				}),
			},
		},
		Outbound: []*proxyman.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&shadowsocks.ClientConfig{
					Server: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(serverPort),
							User: []*protocol.User{
								{
									Account: account,
								},
							},
						},
					},
				}),
			},
		},
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&log.Config{
				ErrorLogLevel: log.LogLevel_Debug,
				ErrorLogType:  log.LogType_Console,
			}),
		},
	}

	servers, err := InitializeServerConfigs(serverConfig, clientConfig)
	assert.Error(err).IsNil()

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			conn, err := net.DialUDP("udp", nil, &net.UDPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: int(clientPort),
			})
			assert.Error(err).IsNil()

			payload := make([]byte, 1024)
			rand.Read(payload)

			nBytes, err := conn.Write([]byte(payload))
			assert.Error(err).IsNil()
			assert.Int(nBytes).Equals(len(payload))

			response := readFrom(conn, time.Second*5, 1024)
			assert.Bytes(response).Equals(xor([]byte(payload)))
			assert.Error(conn.Close()).IsNil()
			wg.Done()
		}()
	}
	wg.Wait()

	CloseAllServers(servers)
}

func TestShadowsocksChacha20TCP(t *testing.T) {
	assert := assert.On(t)

	tcpServer := tcp.Server{
		MsgProcessor: xor,
	}
	dest, err := tcpServer.Start()
	assert.Error(err).IsNil()
	defer tcpServer.Close()

	account := serial.ToTypedMessage(&shadowsocks.Account{
		Password:   "shadowsocks-password",
		CipherType: shadowsocks.CipherType_CHACHA20_IETF,
		Ota:        shadowsocks.Account_Enabled,
	})

	serverPort := pickPort()
	serverConfig := &core.Config{
		Inbound: []*proxyman.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&shadowsocks.ServerConfig{
					User: &protocol.User{
						Account: account,
						Level:   1,
					},
				}),
			},
		},
		Outbound: []*proxyman.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&log.Config{
				ErrorLogLevel: log.LogLevel_Debug,
				ErrorLogType:  log.LogType_Console,
			}),
		},
	}

	clientPort := pickPort()
	clientConfig := &core.Config{
		Inbound: []*proxyman.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(clientPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&dokodemo.Config{
					Address: net.NewIPOrDomain(dest.Address),
					Port:    uint32(dest.Port),
					NetworkList: &net.NetworkList{
						Network: []net.Network{net.Network_TCP},
					},
				}),
			},
		},
		Outbound: []*proxyman.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&shadowsocks.ClientConfig{
					Server: []*protocol.ServerEndpoint{
						{
							Address: net.NewIPOrDomain(net.LocalHostIP),
							Port:    uint32(serverPort),
							User: []*protocol.User{
								{
									Account: account,
								},
							},
						},
					},
				}),
			},
		},
		App: []*serial.TypedMessage{
			serial.ToTypedMessage(&log.Config{
				ErrorLogLevel: log.LogLevel_Debug,
				ErrorLogType:  log.LogType_Console,
			}),
		},
	}

	servers, err := InitializeServerConfigs(serverConfig, clientConfig)
	assert.Error(err).IsNil()

	var wg sync.WaitGroup
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
				IP:   []byte{127, 0, 0, 1},
				Port: int(clientPort),
			})
			assert.Error(err).IsNil()

			payload := make([]byte, 10240*1024)
			rand.Read(payload)

			nBytes, err := conn.Write([]byte(payload))
			assert.Error(err).IsNil()
			assert.Int(nBytes).Equals(len(payload))

			response := readFrom(conn, time.Second*20, 10240*1024)
			assert.Bytes(response).Equals(xor([]byte(payload)))
			assert.Error(conn.Close()).IsNil()
			wg.Done()
		}()
	}
	wg.Wait()

	CloseAllServers(servers)
}
