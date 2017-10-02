package udp

import (
	"context"

	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/dice"
	"github.com/whatedcgveg/v2ray-core/common/net"
)

// Payload represents a single UDP payload.
type Payload struct {
	payload      *buf.Buffer
	source       net.Destination
	originalDest net.Destination
}

// PayloadHandler is function to handle Payload.
type PayloadHandler func(payload *buf.Buffer, source net.Destination, originalDest net.Destination)

// PayloadQueue is a queue of Payload.
type PayloadQueue struct {
	queue    []chan Payload
	callback PayloadHandler
}

// NewPayloadQueue returns a new PayloadQueue.
func NewPayloadQueue(option ListenOption) *PayloadQueue {
	queue := &PayloadQueue{
		callback: option.Callback,
		queue:    make([]chan Payload, option.Concurrency),
	}
	for i := range queue.queue {
		queue.queue[i] = make(chan Payload, 64)
		go queue.Dequeue(queue.queue[i])
	}
	return queue
}

func (v *PayloadQueue) Enqueue(payload Payload) {
	size := len(v.queue)
	idx := 0
	if size > 1 {
		idx = dice.Roll(size)
	}
	for i := 0; i < size; i++ {
		select {
		case v.queue[idx%size] <- payload:
			return
		default:
			idx++
		}
	}
}

func (v *PayloadQueue) Dequeue(queue <-chan Payload) {
	for payload := range queue {
		v.callback(payload.payload, payload.source, payload.originalDest)
	}
}

func (v *PayloadQueue) Close() {
	for _, queue := range v.queue {
		close(queue)
	}
}

type ListenOption struct {
	Callback            PayloadHandler
	ReceiveOriginalDest bool
	Concurrency         int
}

type Hub struct {
	conn   *net.UDPConn
	cancel context.CancelFunc
	queue  *PayloadQueue
	option ListenOption
}

func ListenUDP(address net.Address, port net.Port, option ListenOption) (*Hub, error) {
	if option.Concurrency < 1 {
		option.Concurrency = 1
	}
	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   address.IP(),
		Port: int(port),
	})
	if err != nil {
		return nil, err
	}
	log.Trace(newError("listening UDP on ", address, ":", port))
	if option.ReceiveOriginalDest {
		rawConn, err := udpConn.SyscallConn()
		if err != nil {
			return nil, newError("failed to get fd").Base(err)
		}
		err = rawConn.Control(func(fd uintptr) {
			if err := SetOriginalDestOptions(int(fd)); err != nil {
				log.Trace(newError("failed to set socket options").Base(err))
			}
		})
		if err != nil {
			return nil, newError("failed to control socket").Base(err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	hub := &Hub{
		conn:   udpConn,
		queue:  NewPayloadQueue(option),
		option: option,
		cancel: cancel,
	}
	go hub.start(ctx)
	return hub, nil
}

func (v *Hub) Close() {
	v.cancel()
	v.conn.Close()
}

func (v *Hub) WriteTo(payload []byte, dest net.Destination) (int, error) {
	return v.conn.WriteToUDP(payload, &net.UDPAddr{
		IP:   dest.Address.IP(),
		Port: int(dest.Port),
	})
}

func (v *Hub) start(ctx context.Context) {
	oobBytes := make([]byte, 256)
L:
	for {
		select {
		case <-ctx.Done():
			break L
		default:
		}

		buffer := buf.New()
		var noob int
		var addr *net.UDPAddr
		err := buffer.AppendSupplier(func(b []byte) (int, error) {
			n, nb, _, a, e := ReadUDPMsg(v.conn, b, oobBytes)
			noob = nb
			addr = a
			return n, e
		})

		if err != nil {
			log.Trace(newError("failed to read UDP msg").Base(err))
			buffer.Release()
			continue
		}

		payload := Payload{
			payload: buffer,
		}
		payload.source = net.UDPDestination(net.IPAddress(addr.IP), net.Port(addr.Port))
		if v.option.ReceiveOriginalDest && noob > 0 {
			payload.originalDest = RetrieveOriginalDest(oobBytes[:noob])
		}
		v.queue.Enqueue(payload)
	}
	v.queue.Close()
}

// Connection returns the net.Conn underneath this hub.
// Private: Visible for testing only
func (v *Hub) Connection() net.Conn {
	return v.conn
}

func (v *Hub) Addr() net.Addr {
	return v.conn.LocalAddr()
}
