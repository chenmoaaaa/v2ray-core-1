package impl

//go:generate go run $GOPATH/src/github.com/whatedcgveg/v2ray-core/tools/generrorgen/main.go -pkg impl -path App,Dispatcher,Default

import (
	"context"
	"time"

	"github.com/whatedcgveg/v2ray-core/app"
	"github.com/whatedcgveg/v2ray-core/app/dispatcher"
	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/app/proxyman"
	"github.com/whatedcgveg/v2ray-core/app/router"
	"github.com/whatedcgveg/v2ray-core/common"
	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/proxy"
	"github.com/whatedcgveg/v2ray-core/transport/ray"
)

var (
	errSniffingTimeout = newError("timeout on sniffing")
)

var (
	_ app.Application = (*DefaultDispatcher)(nil)
)

// DefaultDispatcher is a default implementation of Dispatcher.
type DefaultDispatcher struct {
	ohm    proxyman.OutboundHandlerManager
	router *router.Router
}

// NewDefaultDispatcher create a new DefaultDispatcher.
func NewDefaultDispatcher(ctx context.Context, config *dispatcher.Config) (*DefaultDispatcher, error) {
	space := app.SpaceFromContext(ctx)
	if space == nil {
		return nil, newError("no space in context")
	}
	d := &DefaultDispatcher{}
	space.OnInitialize(func() error {
		d.ohm = proxyman.OutboundHandlerManagerFromSpace(space)
		if d.ohm == nil {
			return newError("OutboundHandlerManager is not found in the space")
		}
		d.router = router.FromSpace(space)
		return nil
	})
	return d, nil
}

// Start implements app.Application.
func (*DefaultDispatcher) Start() error {
	return nil
}

// Close implements app.Application.
func (*DefaultDispatcher) Close() {}

// Interface implements app.Application.
func (*DefaultDispatcher) Interface() interface{} {
	return (*dispatcher.Interface)(nil)
}

// Dispatch implements Dispatcher.Interface.
func (d *DefaultDispatcher) Dispatch(ctx context.Context, destination net.Destination) (ray.InboundRay, error) {
	if !destination.IsValid() {
		panic("Dispatcher: Invalid destination.")
	}
	ctx = proxy.ContextWithTarget(ctx, destination)

	outbound := ray.NewRay(ctx)
	sniferList := proxyman.ProtocoSniffersFromContext(ctx)
	if destination.Address.Family().IsDomain() || len(sniferList) == 0 {
		go d.routedDispatch(ctx, outbound, destination)
	} else {
		go func() {
			domain, err := snifer(ctx, sniferList, outbound)
			if err == nil {
				log.Trace(newError("sniffed domain: ", domain))
				destination.Address = net.ParseAddress(domain)
				ctx = proxy.ContextWithTarget(ctx, destination)
			}
			d.routedDispatch(ctx, outbound, destination)
		}()
	}
	return outbound, nil
}

func snifer(ctx context.Context, sniferList []proxyman.KnownProtocols, outbound ray.OutboundRay) (string, error) {
	payload := buf.New()
	defer payload.Release()

	sniffer := NewSniffer(sniferList)
	totalAttempt := 0
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			totalAttempt++
			if totalAttempt > 5 {
				return "", errSniffingTimeout
			}
			outbound.OutboundInput().Peek(payload)
			if !payload.IsEmpty() {
				domain, err := sniffer.Sniff(payload.Bytes())
				if err != ErrMoreData {
					return domain, err
				}
			}
			if payload.IsFull() {
				return "", ErrInvalidData
			}
			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (d *DefaultDispatcher) routedDispatch(ctx context.Context, outbound ray.OutboundRay, destination net.Destination) {
	dispatcher := d.ohm.GetDefaultHandler()
	if d.router != nil {
		if tag, err := d.router.TakeDetour(ctx); err == nil {
			if handler := d.ohm.GetHandler(tag); handler != nil {
				log.Trace(newError("taking detour [", tag, "] for [", destination, "]"))
				dispatcher = handler
			} else {
				log.Trace(newError("nonexisting tag: ", tag).AtWarning())
			}
		} else {
			log.Trace(newError("default route for ", destination))
		}
	}
	dispatcher.Dispatch(ctx, outbound)
}

func init() {
	common.Must(common.RegisterConfig((*dispatcher.Config)(nil), func(ctx context.Context, config interface{}) (interface{}, error) {
		return NewDefaultDispatcher(ctx, config.(*dispatcher.Config))
	}))
}
