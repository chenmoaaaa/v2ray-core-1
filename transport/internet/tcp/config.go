package tcp

import (
	"github.com/whatedcgveg/v2ray-core/common"
	"github.com/whatedcgveg/v2ray-core/transport/internet"
)

func init() {
	common.Must(internet.RegisterProtocolConfigCreator(internet.TransportProtocol_TCP, func() interface{} {
		return new(Config)
	}))
}
