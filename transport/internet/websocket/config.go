package websocket

import (
	"github.com/whatedcgveg/v2ray-core/common"
	"github.com/whatedcgveg/v2ray-core/transport/internet"
)

func (c *Config) GetNormailzedPath() string {
	path := c.Path
	if len(path) == 0 {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
}

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(internet.TransportProtocol_WebSocket, func() interface{} {
		return new(Config)
	}))
}
