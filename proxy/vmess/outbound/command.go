package outbound

import (
	"time"

	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/common/protocol"
	"github.com/whatedcgveg/v2ray-core/common/serial"
	"github.com/whatedcgveg/v2ray-core/proxy/vmess"
)

func (h *Handler) handleSwitchAccount(cmd *protocol.CommandSwitchAccount) {
	account := &vmess.Account{
		Id:      cmd.ID.String(),
		AlterId: uint32(cmd.AlterIds),
		SecuritySettings: &protocol.SecurityConfig{
			Type: protocol.SecurityType_LEGACY,
		},
	}

	user := &protocol.User{
		Email:   "",
		Level:   cmd.Level,
		Account: serial.ToTypedMessage(account),
	}
	dest := net.TCPDestination(cmd.Host, cmd.Port)
	until := time.Now().Add(time.Duration(cmd.ValidMin) * time.Minute)
	h.serverList.AddServer(protocol.NewServerSpec(dest, protocol.BeforeTime(until), user))
}

func (h *Handler) handleCommand(dest net.Destination, cmd protocol.ResponseCommand) {
	switch typedCommand := cmd.(type) {
	case *protocol.CommandSwitchAccount:
		if typedCommand.Host == nil {
			typedCommand.Host = dest.Address
		}
		h.handleSwitchAccount(typedCommand)
	default:
	}
}
