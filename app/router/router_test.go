package router_test

import (
	"context"
	"testing"

	"github.com/whatedcgveg/v2ray-core/app"
	"github.com/whatedcgveg/v2ray-core/app/dispatcher"
	_ "github.com/whatedcgveg/v2ray-core/app/dispatcher/impl"
	"github.com/whatedcgveg/v2ray-core/app/dns"
	_ "github.com/whatedcgveg/v2ray-core/app/dns/server"
	"github.com/whatedcgveg/v2ray-core/app/proxyman"
	_ "github.com/whatedcgveg/v2ray-core/app/proxyman/outbound"
	. "github.com/whatedcgveg/v2ray-core/app/router"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/proxy"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestSimpleRouter(t *testing.T) {
	assert := assert.On(t)

	config := &Config{
		Rule: []*RoutingRule{
			{
				Tag: "test",
				NetworkList: &net.NetworkList{
					Network: []net.Network{net.Network_TCP},
				},
			},
		},
	}

	space := app.NewSpace()
	ctx := app.ContextWithSpace(context.Background(), space)
	assert.Error(app.AddApplicationToSpace(ctx, new(dns.Config))).IsNil()
	assert.Error(app.AddApplicationToSpace(ctx, new(dispatcher.Config))).IsNil()
	assert.Error(app.AddApplicationToSpace(ctx, new(proxyman.OutboundConfig))).IsNil()
	assert.Error(app.AddApplicationToSpace(ctx, config)).IsNil()
	assert.Error(space.Initialize()).IsNil()

	r := FromSpace(space)

	ctx = proxy.ContextWithTarget(ctx, net.TCPDestination(net.DomainAddress("v2ray.com"), 80))
	tag, err := r.TakeDetour(ctx)
	assert.Error(err).IsNil()
	assert.String(tag).Equals("test")
}
