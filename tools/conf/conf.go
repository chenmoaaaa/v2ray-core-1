package conf

import (
	"io"

	"github.com/whatedcgveg/v2ray-core"
	jsonconf "github.com/whatedcgveg/ext/tools/conf/serial"
)

//go:generate go run $GOPATH/src/github.com/whatedcgveg/v2ray-core/tools/generrorgen/main.go -pkg conf -path Tools,Conf

func init() {
	core.RegisterConfigLoader(core.ConfigFormat_JSON, func(input io.Reader) (*core.Config, error) {
		return jsonconf.LoadJSONConfig(input)
	})
}
