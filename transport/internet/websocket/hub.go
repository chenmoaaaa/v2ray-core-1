package websocket

import (
	"context"
	"crypto/tls"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/common"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/transport/internet"
	v2tls "github.com/whatedcgveg/v2ray-core/transport/internet/tls"
)

type requestHandler struct {
	path string
	ln   *Listener
}

func (h *requestHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != h.path {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	conn, err := converttovws(writer, request)
	if err != nil {
		log.Trace(newError("failed to convert to WebSocket connection").Base(err))
		return
	}

	h.ln.addConn(h.ln.ctx, internet.Connection(conn))
}

type Listener struct {
	sync.Mutex
	ctx       context.Context
	listener  net.Listener
	tlsConfig *tls.Config
	config    *Config
	addConn   internet.AddConnection
}

func ListenWS(ctx context.Context, address net.Address, port net.Port, addConn internet.AddConnection) (internet.Listener, error) {
	networkSettings := internet.TransportSettingsFromContext(ctx)
	wsSettings := networkSettings.(*Config)

	l := &Listener{
		ctx:     ctx,
		config:  wsSettings,
		addConn: addConn,
	}
	if securitySettings := internet.SecuritySettingsFromContext(ctx); securitySettings != nil {
		tlsConfig, ok := securitySettings.(*v2tls.Config)
		if ok {
			l.tlsConfig = tlsConfig.GetTLSConfig()
		}
	}

	err := l.listenws(address, port)

	return l, err
}

func (ln *Listener) listenws(address net.Address, port net.Port) error {
	netAddr := address.String() + ":" + strconv.Itoa(int(port.Value()))
	var listener net.Listener
	if ln.tlsConfig == nil {
		l, err := net.Listen("tcp", netAddr)
		if err != nil {
			return newError("failed to listen TCP ", netAddr).Base(err)
		}
		listener = l
	} else {
		l, err := tls.Listen("tcp", netAddr, ln.tlsConfig)
		if err != nil {
			return newError("failed to listen TLS ", netAddr).Base(err)
		}
		listener = l
	}
	ln.listener = listener

	go func() {
		http.Serve(listener, &requestHandler{
			path: ln.config.GetNormailzedPath(),
			ln:   ln,
		})
	}()

	return nil
}

func converttovws(w http.ResponseWriter, r *http.Request) (*connection, error) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  32 * 1024,
		WriteBufferSize: 32 * 1024,
	}
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		return nil, err
	}

	return &connection{wsc: conn}, nil
}

func (ln *Listener) Addr() net.Addr {
	return ln.listener.Addr()
}

func (ln *Listener) Close() error {
	return ln.listener.Close()
}

func init() {
	common.Must(internet.RegisterTransportListener(internet.TransportProtocol_WebSocket, ListenWS))
}
