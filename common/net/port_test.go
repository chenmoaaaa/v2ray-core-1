package net_test

import (
	"testing"

	. "github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestPortRangeContains(t *testing.T) {
	assert := assert.On(t)

	portRange := &PortRange{
		From: 53,
		To:   53,
	}
	assert.Bool(portRange.Contains(Port(53))).IsTrue()
}
