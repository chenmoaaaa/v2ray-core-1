package inbound

import (
	"context"

	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/app/proxyman"
	"github.com/whatedcgveg/v2ray-core/app/proxyman/mux"
	"github.com/whatedcgveg/v2ray-core/common/dice"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/proxy"
)

type AlwaysOnInboundHandler struct {
	proxy   proxy.Inbound
	workers []worker
	mux     *mux.Server
}

func NewAlwaysOnInboundHandler(ctx context.Context, tag string, receiverConfig *proxyman.ReceiverConfig, proxyConfig interface{}) (*AlwaysOnInboundHandler, error) {
	p, err := proxy.CreateInboundHandler(ctx, proxyConfig)
	if err != nil {
		return nil, err
	}

	h := &AlwaysOnInboundHandler{
		proxy: p,
		mux:   mux.NewServer(ctx),
	}

	nl := p.Network()
	pr := receiverConfig.PortRange
	address := receiverConfig.Listen.AsAddress()
	if address == nil {
		address = net.AnyIP
	}
	for port := pr.From; port <= pr.To; port++ {
		if nl.HasNetwork(net.Network_TCP) {
			log.Trace(newError("creating tcp worker on ", address, ":", port).AtDebug())
			worker := &tcpWorker{
				address:      address,
				port:         net.Port(port),
				proxy:        p,
				stream:       receiverConfig.StreamSettings,
				recvOrigDest: receiverConfig.ReceiveOriginalDestination,
				tag:          tag,
				dispatcher:   h.mux,
				sniffers:     receiverConfig.DomainOverride,
			}
			h.workers = append(h.workers, worker)
		}

		if nl.HasNetwork(net.Network_UDP) {
			worker := &udpWorker{
				tag:          tag,
				proxy:        p,
				address:      address,
				port:         net.Port(port),
				recvOrigDest: receiverConfig.ReceiveOriginalDestination,
				dispatcher:   h.mux,
			}
			h.workers = append(h.workers, worker)
		}
	}

	return h, nil
}

func (h *AlwaysOnInboundHandler) Start() error {
	for _, worker := range h.workers {
		if err := worker.Start(); err != nil {
			return err
		}
	}
	return nil
}

func (h *AlwaysOnInboundHandler) Close() {
	for _, worker := range h.workers {
		worker.Close()
	}
}

func (h *AlwaysOnInboundHandler) GetRandomInboundProxy() (proxy.Inbound, net.Port, int) {
	if len(h.workers) == 0 {
		return nil, 0, 0
	}
	w := h.workers[dice.Roll(len(h.workers))]
	return w.Proxy(), w.Port(), 9999
}
