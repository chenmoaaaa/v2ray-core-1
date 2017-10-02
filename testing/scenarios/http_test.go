package scenarios

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/whatedcgveg/v2ray-core"
	"github.com/whatedcgveg/v2ray-core/app/proxyman"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/common/serial"
	"github.com/whatedcgveg/v2ray-core/proxy/freedom"
	v2http "github.com/whatedcgveg/v2ray-core/proxy/http"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
	v2httptest "github.com/whatedcgveg/v2ray-core/testing/servers/http"
)

func TestHttpConformance(t *testing.T) {
	assert := assert.On(t)

	httpServerPort := pickPort()
	httpServer := &v2httptest.Server{
		Port:        httpServerPort,
		PathHandler: make(map[string]http.HandlerFunc),
	}
	_, err := httpServer.Start()
	assert.Error(err).IsNil()
	defer httpServer.Close()

	serverPort := pickPort()
	serverConfig := &core.Config{
		Inbound: []*proxyman.InboundHandlerConfig{
			{
				ReceiverSettings: serial.ToTypedMessage(&proxyman.ReceiverConfig{
					PortRange: net.SinglePortRange(serverPort),
					Listen:    net.NewIPOrDomain(net.LocalHostIP),
				}),
				ProxySettings: serial.ToTypedMessage(&v2http.ServerConfig{}),
			},
		},
		Outbound: []*proxyman.OutboundHandlerConfig{
			{
				ProxySettings: serial.ToTypedMessage(&freedom.Config{}),
			},
		},
	}

	servers, err := InitializeServerConfigs(serverConfig)
	assert.Error(err).IsNil()

	{
		transport := &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				return url.Parse("http://127.0.0.1:" + serverPort.String())
			},
		}

		client := &http.Client{
			Transport: transport,
		}

		resp, err := client.Get("http://127.0.0.1:" + httpServerPort.String())
		assert.Error(err).IsNil()
		assert.Int(resp.StatusCode).Equals(200)

		content, err := ioutil.ReadAll(resp.Body)
		assert.Error(err).IsNil()
		assert.String(string(content)).Equals("Home")

	}

	CloseAllServers(servers)
}
