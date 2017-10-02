package protocol_test

import (
	"testing"

	"github.com/whatedcgveg/v2ray-core/common/predicate"
	. "github.com/whatedcgveg/v2ray-core/common/protocol"
	"github.com/whatedcgveg/v2ray-core/common/uuid"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestCmdKey(t *testing.T) {
	assert := assert.On(t)

	id := NewID(uuid.New())
	assert.Bool(predicate.BytesAll(id.CmdKey(), 0)).IsFalse()
}
