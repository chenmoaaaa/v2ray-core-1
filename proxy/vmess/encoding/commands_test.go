package encoding_test

import (
	"testing"

	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/protocol"
	"github.com/whatedcgveg/v2ray-core/common/uuid"
	. "github.com/whatedcgveg/v2ray-core/proxy/vmess/encoding"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestSwitchAccount(t *testing.T) {
	assert := assert.On(t)

	sa := &protocol.CommandSwitchAccount{
		Port:     1234,
		ID:       uuid.New(),
		AlterIds: 1024,
		Level:    128,
		ValidMin: 16,
	}

	buffer := buf.New()
	err := MarshalCommand(sa, buffer)
	assert.Error(err).IsNil()

	cmd, err := UnmarshalCommand(1, buffer.BytesFrom(2))
	assert.Error(err).IsNil()

	sa2, ok := cmd.(*protocol.CommandSwitchAccount)
	assert.Bool(ok).IsTrue()
	assert.Pointer(sa.Host).IsNil()
	assert.Pointer(sa2.Host).IsNil()
	assert.Port(sa.Port).Equals(sa2.Port)
	assert.String(sa.ID.String()).Equals(sa2.ID.String())
	assert.Uint16(sa.AlterIds).Equals(sa2.AlterIds)
	assert.Byte(byte(sa.Level)).Equals(byte(sa2.Level))
	assert.Byte(sa.ValidMin).Equals(sa2.ValidMin)
}
