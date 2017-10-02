package freedom

//go:generate go run $GOPATH/src/github.com/whatedcgveg/v2ray-core/tools/generrorgen/main.go -pkg freedom -path Proxy,Freedom

import (
	"context"
	"runtime"
	"time"

	"github.com/whatedcgveg/v2ray-core/app"
	"github.com/whatedcgveg/v2ray-core/app/dns"
	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/common"
	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/dice"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/common/retry"
	"github.com/whatedcgveg/v2ray-core/common/signal"
	"github.com/whatedcgveg/v2ray-core/proxy"
	"github.com/whatedcgveg/v2ray-core/transport/internet"
	"github.com/whatedcgveg/v2ray-core/transport/ray"
)

type Handler struct {
	domainStrategy Config_DomainStrategy
	timeout        uint32
	dns            dns.Server
	destOverride   *DestinationOverride
}

func New(ctx context.Context, config *Config) (*Handler, error) {
	space := app.SpaceFromContext(ctx)
	if space == nil {
		return nil, newError("no space in context")
	}
	f := &Handler{
		domainStrategy: config.DomainStrategy,
		timeout:        config.Timeout,
		destOverride:   config.DestinationOverride,
	}
	space.OnInitialize(func() error {
		if config.DomainStrategy == Config_USE_IP {
			f.dns = dns.FromSpace(space)
			if f.dns == nil {
				return newError("DNS server is not found in the space")
			}
		}
		return nil
	})
	return f, nil
}

func (v *Handler) ResolveIP(destination net.Destination) net.Destination {
	if !destination.Address.Family().IsDomain() {
		return destination
	}

	ips := v.dns.Get(destination.Address.Domain())
	if len(ips) == 0 {
		log.Trace(newError("DNS returns nil answer. Keep domain as is."))
		return destination
	}

	ip := ips[dice.Roll(len(ips))]
	newDest := net.Destination{
		Network: destination.Network,
		Address: net.IPAddress(ip),
		Port:    destination.Port,
	}
	log.Trace(newError("changing destination from ", destination, " to ", newDest))
	return newDest
}

func (v *Handler) Process(ctx context.Context, outboundRay ray.OutboundRay, dialer proxy.Dialer) error {
	destination, _ := proxy.TargetFromContext(ctx)
	if v.destOverride != nil {
		server := v.destOverride.Server
		destination = net.Destination{
			Network: destination.Network,
			Address: server.Address.AsAddress(),
			Port:    net.Port(server.Port),
		}
	}
	log.Trace(newError("opening connection to ", destination))

	input := outboundRay.OutboundInput()
	output := outboundRay.OutboundOutput()

	var conn internet.Connection
	if v.domainStrategy == Config_USE_IP && destination.Address.Family().IsDomain() {
		destination = v.ResolveIP(destination)
	}

	err := retry.ExponentialBackoff(5, 100).On(func() error {
		rawConn, err := dialer.Dial(ctx, destination)
		if err != nil {
			return err
		}
		conn = rawConn
		return nil
	})
	if err != nil {
		return newError("failed to open connection to ", destination).Base(err)
	}
	defer conn.Close()

	timeout := time.Second * time.Duration(v.timeout)
	if timeout == 0 {
		timeout = time.Minute * 5
	}
	ctx, timer := signal.CancelAfterInactivity(ctx, timeout)

	requestDone := signal.ExecuteAsync(func() error {
		var writer buf.Writer
		if destination.Network == net.Network_TCP {
			writer = buf.NewWriter(conn)
		} else {
			writer = buf.NewSequentialWriter(conn)
		}
		if err := buf.Copy(input, writer, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to process request").Base(err)
		}
		return nil
	})

	responseDone := signal.ExecuteAsync(func() error {
		defer output.Close()

		v2reader := buf.NewReader(conn)
		if err := buf.Copy(v2reader, output, buf.UpdateActivity(timer)); err != nil {
			return newError("failed to process response").Base(err)
		}
		return nil
	})

	if err := signal.ErrorOrFinish2(ctx, requestDone, responseDone); err != nil {
		input.CloseError()
		output.CloseError()
		return newError("connection ends").Base(err)
	}

	runtime.KeepAlive(timer)

	return nil
}

func init() {
	common.Must(common.RegisterConfig((*Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return New(ctx, config.(*Config))
	}))
}
