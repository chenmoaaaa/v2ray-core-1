package udp

import (
	"context"
	"sync"

	"github.com/whatedcgveg/v2ray-core/app/dispatcher"
	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/transport/ray"
)

type ResponseCallback func(payload *buf.Buffer)

type Dispatcher struct {
	sync.RWMutex
	conns      map[net.Destination]ray.InboundRay
	dispatcher dispatcher.Interface
}

func NewDispatcher(dispatcher dispatcher.Interface) *Dispatcher {
	return &Dispatcher{
		conns:      make(map[net.Destination]ray.InboundRay),
		dispatcher: dispatcher,
	}
}

func (v *Dispatcher) RemoveRay(dest net.Destination) {
	v.Lock()
	defer v.Unlock()
	if conn, found := v.conns[dest]; found {
		conn.InboundInput().Close()
		conn.InboundOutput().Close()
		delete(v.conns, dest)
	}
}

func (v *Dispatcher) getInboundRay(ctx context.Context, dest net.Destination) (ray.InboundRay, bool) {
	v.Lock()
	defer v.Unlock()

	if entry, found := v.conns[dest]; found {
		return entry, true
	}

	log.Trace(newError("establishing new connection for ", dest))
	inboundRay, _ := v.dispatcher.Dispatch(ctx, dest)
	v.conns[dest] = inboundRay
	return inboundRay, false
}

func (v *Dispatcher) Dispatch(ctx context.Context, destination net.Destination, payload *buf.Buffer, callback ResponseCallback) {
	// TODO: Add user to destString
	log.Trace(newError("dispatch request to: ", destination).AtDebug())

	inboundRay, existing := v.getInboundRay(ctx, destination)
	outputStream := inboundRay.InboundInput()
	if outputStream != nil {
		if err := outputStream.Write(buf.NewMultiBufferValue(payload)); err != nil {
			v.RemoveRay(destination)
		}
	}
	if !existing {
		go func() {
			handleInput(inboundRay.InboundOutput(), callback)
			v.RemoveRay(destination)
		}()
	}
}

func handleInput(input ray.InputStream, callback ResponseCallback) {
	for {
		mb, err := input.Read()
		if err != nil {
			break
		}
		for _, b := range mb {
			callback(b)
		}
	}
}