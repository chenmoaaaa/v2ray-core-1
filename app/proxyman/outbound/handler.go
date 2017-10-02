package outbound

import (
	"context"
	"io"
	"time"

	"github.com/whatedcgveg/v2ray-core/app"
	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/app/proxyman"
	"github.com/whatedcgveg/v2ray-core/app/proxyman/mux"
	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/errors"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/proxy"
	"github.com/whatedcgveg/v2ray-core/transport/internet"
	"github.com/whatedcgveg/v2ray-core/transport/ray"
)

type Handler struct {
	config          *proxyman.OutboundHandlerConfig
	senderSettings  *proxyman.SenderConfig
	proxy           proxy.Outbound
	outboundManager proxyman.OutboundHandlerManager
	mux             *mux.ClientManager
}

func NewHandler(ctx context.Context, config *proxyman.OutboundHandlerConfig) (*Handler, error) {
	h := &Handler{
		config: config,
	}
	space := app.SpaceFromContext(ctx)
	if space == nil {
		return nil, newError("no space in context")
	}
	space.OnInitialize(func() error {
		ohm := proxyman.OutboundHandlerManagerFromSpace(space)
		if ohm == nil {
			return newError("no OutboundManager in space")
		}
		h.outboundManager = ohm
		return nil
	})

	if config.SenderSettings != nil {
		senderSettings, err := config.SenderSettings.GetInstance()
		if err != nil {
			return nil, err
		}
		switch s := senderSettings.(type) {
		case *proxyman.SenderConfig:
			h.senderSettings = s
		default:
			return nil, newError("settings is not SenderConfig")
		}
	}

	proxyHandler, err := config.GetProxyHandler(ctx)
	if err != nil {
		return nil, err
	}

	if h.senderSettings != nil && h.senderSettings.MultiplexSettings != nil && h.senderSettings.MultiplexSettings.Enabled {
		config := h.senderSettings.MultiplexSettings
		if config.Concurrency < 1 || config.Concurrency > 1024 {
			return nil, newError("invalid mux concurrency: ", config.Concurrency)
		}
		h.mux = mux.NewClientManager(proxyHandler, h, config)
	}

	h.proxy = proxyHandler
	return h, nil
}

// Dispatch implements proxy.Outbound.Dispatch.
func (h *Handler) Dispatch(ctx context.Context, outboundRay ray.OutboundRay) {
	if h.mux != nil {
		err := h.mux.Dispatch(ctx, outboundRay)
		if err != nil {
			log.Trace(newError("failed to process outbound traffic").Base(err))
		}
	} else {
		err := h.proxy.Process(ctx, outboundRay, h)
		// Ensure outbound ray is properly closed.
		if err != nil && errors.Cause(err) != io.EOF {
			log.Trace(newError("failed to process outbound traffic").Base(err))
			outboundRay.OutboundOutput().CloseError()
		} else {
			outboundRay.OutboundOutput().Close()
		}
		outboundRay.OutboundInput().CloseError()
	}
}

// Dial implements proxy.Dialer.Dial().
func (h *Handler) Dial(ctx context.Context, dest net.Destination) (internet.Connection, error) {
	if h.senderSettings != nil {
		if h.senderSettings.ProxySettings.HasTag() {
			tag := h.senderSettings.ProxySettings.Tag
			handler := h.outboundManager.GetHandler(tag)
			if handler != nil {
				log.Trace(newError("proxying to ", tag).AtDebug())
				ctx = proxy.ContextWithTarget(ctx, dest)
				stream := ray.NewRay(ctx)
				go handler.Dispatch(ctx, stream)
				return NewConnection(stream), nil
			}

			log.Trace(newError("failed to get outbound handler with tag: ", tag).AtWarning())
		}

		if h.senderSettings.Via != nil {
			ctx = internet.ContextWithDialerSource(ctx, h.senderSettings.Via.AsAddress())
		}

		if h.senderSettings.StreamSettings != nil {
			ctx = internet.ContextWithStreamSettings(ctx, h.senderSettings.StreamSettings)
		}
	}

	return internet.Dial(ctx, dest)
}

var (
	_ buf.MultiBufferReader = (*Connection)(nil)
	_ buf.MultiBufferWriter = (*Connection)(nil)
)

type Connection struct {
	stream     ray.Ray
	closed     bool
	localAddr  net.Addr
	remoteAddr net.Addr

	bytesReader io.Reader
	reader      buf.Reader
	writer      buf.Writer
}

func NewConnection(stream ray.Ray) *Connection {
	return &Connection{
		stream: stream,
		localAddr: &net.TCPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: 0,
		},
		remoteAddr: &net.TCPAddr{
			IP:   []byte{0, 0, 0, 0},
			Port: 0,
		},
		bytesReader: buf.ToBytesReader(stream.InboundOutput()),
		reader:      stream.InboundOutput(),
		writer:      stream.InboundInput(),
	}
}

// Read implements net.Conn.Read().
func (v *Connection) Read(b []byte) (int, error) {
	if v.closed {
		return 0, io.EOF
	}
	return v.bytesReader.Read(b)
}

func (v *Connection) ReadMultiBuffer() (buf.MultiBuffer, error) {
	return v.reader.Read()
}

// Write implements net.Conn.Write().
func (v *Connection) Write(b []byte) (int, error) {
	if v.closed {
		return 0, io.ErrClosedPipe
	}
	return buf.ToBytesWriter(v.writer).Write(b)
}

func (v *Connection) WriteMultiBuffer(mb buf.MultiBuffer) error {
	if v.closed {
		return io.ErrClosedPipe
	}
	return v.writer.Write(mb)
}

// Close implements net.Conn.Close().
func (v *Connection) Close() error {
	v.closed = true
	v.stream.InboundInput().Close()
	v.stream.InboundOutput().CloseError()
	return nil
}

// LocalAddr implements net.Conn.LocalAddr().
func (v *Connection) LocalAddr() net.Addr {
	return v.localAddr
}

// RemoteAddr implements net.Conn.RemoteAddr().
func (v *Connection) RemoteAddr() net.Addr {
	return v.remoteAddr
}

// SetDeadline implements net.Conn.SetDeadline().
func (v *Connection) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline implements net.Conn.SetReadDeadline().
func (v *Connection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline implement net.Conn.SetWriteDeadline().
func (v *Connection) SetWriteDeadline(t time.Time) error {
	return nil
}