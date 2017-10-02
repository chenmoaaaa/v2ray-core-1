package vmess

import (
	"github.com/whatedcgveg/v2ray-core/app/log"
	"github.com/whatedcgveg/v2ray-core/common/dice"
	"github.com/whatedcgveg/v2ray-core/common/protocol"
	"github.com/whatedcgveg/v2ray-core/common/uuid"
)

type InternalAccount struct {
	ID       *protocol.ID
	AlterIDs []*protocol.ID
	Security protocol.Security
}

func (a *InternalAccount) AnyValidID() *protocol.ID {
	if len(a.AlterIDs) == 0 {
		return a.ID
	}
	return a.AlterIDs[dice.Roll(len(a.AlterIDs))]
}

func (a *InternalAccount) Equals(account protocol.Account) bool {
	vmessAccount, ok := account.(*InternalAccount)
	if !ok {
		return false
	}
	// TODO: handle AlterIds difference
	return a.ID.Equals(vmessAccount.ID)
}

func (a *Account) AsAccount() (protocol.Account, error) {
	id, err := uuid.ParseString(a.Id)
	if err != nil {
		log.Trace(newError("failed to parse ID").Base(err).AtError())
		return nil, err
	}
	protoID := protocol.NewID(id)
	return &InternalAccount{
		ID:       protoID,
		AlterIDs: protocol.NewAlterIDs(protoID, uint16(a.AlterId)),
		Security: a.SecuritySettings.AsSecurity(),
	}, nil
}
