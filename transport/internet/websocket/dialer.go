package websocket

import (
	"context"

	"github.com/gorilla/websocket"
	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/common"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/transport/internet"
	"github.com/whatedcgveg/v2ray-core/transport/internet/tls"
)

// Dial dials a WebSocket connection to the given destination.
func Dial(ctx context.Context, dest net.Destination) (internet.Connection, error) {
	log.Trace(newError("creating connection to ", dest))

	conn, err := dialWebsocket(ctx, dest)
	if err != nil {
		return nil, newError("failed to dial WebSocket").Base(err)
	}
	return internet.Connection(conn), nil
}

func init() {
	common.Must(internet.RegisterTransportDialer(internet.TransportProtocol_WebSocket, Dial))
}

func dialWebsocket(ctx context.Context, dest net.Destination) (net.Conn, error) {
	src := internet.DialerSourceFromContext(ctx)
	wsSettings := internet.TransportSettingsFromContext(ctx).(*Config)

	commonDial := func(network, addr string) (net.Conn, error) {
		return internet.DialSystem(ctx, src, dest)
	}

	dialer := websocket.Dialer{
		NetDial:         commonDial,
		ReadBufferSize:  32 * 1024,
		WriteBufferSize: 32 * 1024,
	}

	protocol := "ws"

	if securitySettings := internet.SecuritySettingsFromContext(ctx); securitySettings != nil {
		tlsConfig, ok := securitySettings.(*tls.Config)
		if ok {
			protocol = "wss"
			dialer.TLSClientConfig = tlsConfig.GetTLSConfig()
			if dest.Address.Family().IsDomain() {
				dialer.TLSClientConfig.ServerName = dest.Address.Domain()
			}
		}
	}

	host := dest.NetAddr()
	if (protocol == "ws" && dest.Port == 80) || (protocol == "wss" && dest.Port == 443) {
		host = dest.Address.String()
	}
	uri := protocol + "://" + host + wsSettings.GetNormailzedPath()

	conn, resp, err := dialer.Dial(uri, nil)
	if err != nil {
		var reason string
		if resp != nil {
			reason = resp.Status
		}
		return nil, newError("failed to dial to (", uri, "): ", reason).Base(err)
	}

	return &connection{
		wsc: conn,
	}, nil
}
