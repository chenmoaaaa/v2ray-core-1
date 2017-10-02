package router

//go:generate go run $GOPATH/src/github.com/whatedcgveg/v2ray-core/tools/generrorgen/main.go -pkg router -path App,Router

import (
	"context"

	"github.com/whatedcgveg/v2ray-core/app"
	"github.com/whatedcgveg/v2ray-core/app/dns"
	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/common"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/proxy"
)

var (
	ErrNoRuleApplicable = newError("No rule applicable")
)

type Router struct {
	domainStrategy Config_DomainStrategy
	rules          []Rule
	dnsServer      dns.Server
}

func NewRouter(ctx context.Context, config *Config) (*Router, error) {
	space := app.SpaceFromContext(ctx)
	if space == nil {
		return nil, newError("no space in context")
	}
	r := &Router{
		domainStrategy: config.DomainStrategy,
		rules:          make([]Rule, len(config.Rule)),
	}

	space.OnInitialize(func() error {
		for idx, rule := range config.Rule {
			r.rules[idx].Tag = rule.Tag
			cond, err := rule.BuildCondition()
			if err != nil {
				return err
			}
			r.rules[idx].Condition = cond
		}

		r.dnsServer = dns.FromSpace(space)
		if r.dnsServer == nil {
			return newError("DNS is not found in the space")
		}
		return nil
	})
	return r, nil
}

func (r *Router) resolveIP(dest net.Destination) []net.Address {
	ips := r.dnsServer.Get(dest.Address.Domain())
	if len(ips) == 0 {
		return nil
	}
	dests := make([]net.Address, len(ips))
	for idx, ip := range ips {
		dests[idx] = net.IPAddress(ip)
	}
	return dests
}

func (r *Router) TakeDetour(ctx context.Context) (string, error) {
	for _, rule := range r.rules {
		if rule.Apply(ctx) {
			return rule.Tag, nil
		}
	}

	dest, ok := proxy.TargetFromContext(ctx)
	if !ok {
		return "", ErrNoRuleApplicable
	}

	if r.domainStrategy == Config_IpIfNonMatch && dest.Address.Family().IsDomain() {
		log.Trace(newError("looking up IP for ", dest))
		ipDests := r.resolveIP(dest)
		if ipDests != nil {
			ctx = proxy.ContextWithResolveIPs(ctx, ipDests)
			for _, rule := range r.rules {
				if rule.Apply(ctx) {
					return rule.Tag, nil
				}
			}
		}
	}

	return "", ErrNoRuleApplicable
}

func (*Router) Interface() interface{} {
	return (*Router)(nil)
}

func (*Router) Start() error {
	return nil
}

func (*Router) Close() {}

func FromSpace(space app.Space) *Router {
	app := space.GetApplication((*Router)(nil))
	if app == nil {
		return nil
	}
	return app.(*Router)
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewRouter(ctx, config.(*Config))
	}))
}
