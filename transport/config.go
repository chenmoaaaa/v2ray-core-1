package transport

import (
	"github.com/whatedcgveg/v2ray-core/transport/internet"
)

// Apply applies this Config.
func (c *Config) Apply() error {
	if c == nil {
		return nil
	}
	if err := internet.ApplyGlobalTransportSettings(c.TransportSettings); err != nil {
		return err
	}
	return nil
}
