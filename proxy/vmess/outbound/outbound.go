package outbound

//go:generate go run $GOPATH/src/github.com/whatedcgveg/v2ray-core/tools/generrorgen/main.go -pkg outbound -path Proxy,VMess,Outbound

import (
	"context"
	"runtime"
	"time"

	"github.com/whatedcgveg/v2ray-core/app"
	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/common"
	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/common/protocol"
	"github.com/whatedcgveg/v2ray-core/common/retry"
	"github.com/whatedcgveg/v2ray-core/common/signal"
	"github.com/whatedcgveg/v2ray-core/proxy"
	"github.com/whatedcgveg/v2ray-core/proxy/vmess"
	"github.com/whatedcgveg/v2ray-core/proxy/vmess/encoding"
	"github.com/whatedcgveg/v2ray-core/transport/internet"
	"github.com/whatedcgveg/v2ray-core/transport/ray"
)

// Handler is an outbound connection handler for VMess protocol.
type Handler struct {
	serverList   *protocol.ServerList
	serverPicker protocol.ServerPicker
}

func New(ctx context.Context, config *Config) (*Handler, error) {
	space := app.SpaceFromContext(ctx)
	if space == nil {
		return nil, newError("no space in context.")
	}

	serverList := protocol.NewServerList()
	for _, rec := range config.Receiver {
		serverList.AddServer(protocol.NewServerSpecFromPB(*rec))
	}
	handler := &Handler{
		serverList:   serverList,
		serverPicker: protocol.NewRoundRobinServerPicker(serverList),
	}

	return handler, nil
}

// Process implements proxy.Outbound.Process().
func (v *Handler) Process(ctx context.Context, outboundRay ray.OutboundRay, dialer proxy.Dialer) error {
	var rec *protocol.ServerSpec
	var conn internet.Connection

	err := retry.ExponentialBackoff(5, 200).On(func() error {
		rec = v.serverPicker.PickServer()
		rawConn, err := dialer.Dial(ctx, rec.Destination())
		if err != nil {
			return err
		}
		conn = rawConn

		return nil
	})
	if err != nil {
		return newError("failed to find an available destination").Base(err).AtWarning()
	}
	defer conn.Close()

	target, ok := proxy.TargetFromContext(ctx)
	if !ok {
		return newError("target not specified").AtError()
	}
	log.Trace(newError("tunneling request to ", target, " via ", rec.Destination()))

	command := protocol.RequestCommandTCP
	if target.Network == net.Network_UDP {
		command = protocol.RequestCommandUDP
	}
	//if target.Address.Family().IsDomain() && target.Address.Domain() == "v1.mux.com" {
	//	command = protocol.RequestCommandMux
	//}
	request := &protocol.RequestHeader{
		Version: encoding.Version,
		User:    rec.PickUser(),
		Command: command,
		Address: target.Address,
		Port:    target.Port,
		Option:  protocol.RequestOptionChunkStream,
	}

	rawAccount, err := request.User.GetTypedAccount()
	if err != nil {
		return newError("failed to get user account").Base(err).AtWarning()
	}
	account := rawAccount.(*vmess.InternalAccount)
	request.Security = account.Security

	if request.Security.Is(protocol.SecurityType_AES128_GCM) || request.Security.Is(protocol.SecurityType_NONE) || request.Security.Is(protocol.SecurityType_CHACHA20_POLY1305) {
		request.Option.Set(protocol.RequestOptionChunkMasking)
	}

	input := outboundRay.OutboundInput()
	output := outboundRay.OutboundOutput()

	session := encoding.NewClientSession(protocol.DefaultIDHash)

	ctx, timer := signal.CancelAfterInactivity(ctx, time.Minute*5)

	requestDone := signal.ExecuteAsync(func() error {
		writer := buf.NewBufferedWriter(conn)
		session.EncodeRequestHeader(request, writer)

		bodyWriter := session.EncodeRequestBody(request, writer)
		firstPayload, err := input.ReadTimeout(time.Millisecond * 500)
		if err != nil && err != buf.ErrReadTimeout {
			return newError("failed to get first payload").Base(err)
		}
		if !firstPayload.IsEmpty() {
			if err := bodyWriter.Write(firstPayload); err != nil {
				return newError("failed to write first payload").Base(err)
			}
			firstPayload.Release()
		}

		if err := writer.SetBuffered(false); err != nil {
			return err
		}

		if err := buf.Copy(input, bodyWriter, buf.UpdateActivity(timer)); err != nil {
			return err
		}

		if request.Option.Has(protocol.RequestOptionChunkStream) {
			if err := bodyWriter.Write(buf.NewMultiBuffer()); err != nil {
				return err
			}
		}
		return nil
	})

	responseDone := signal.ExecuteAsync(func() error {
		defer output.Close()

		reader := buf.NewBufferedReader(conn)
		header, err := session.DecodeResponseHeader(reader)
		if err != nil {
			return err
		}
		v.handleCommand(rec.Destination(), header.Command)

		reader.SetBuffered(false)
		bodyReader := session.DecodeResponseBody(request, reader)
		if err := buf.Copy(bodyReader, output, buf.UpdateActivity(timer)); err != nil {
			return err
		}

		return nil
	})

	if err := signal.ErrorOrFinish2(ctx, requestDone, responseDone); err != nil {
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
