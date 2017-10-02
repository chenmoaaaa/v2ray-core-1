package serial_test

import (
	"testing"

	"github.com/whatedcgveg/v2ray-core/common"
	"github.com/whatedcgveg/v2ray-core/common/buf"
	. "github.com/whatedcgveg/v2ray-core/common/serial"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestUint32(t *testing.T) {
	assert := assert.On(t)

	x := uint32(458634234)
	s1 := Uint32ToBytes(x, []byte{})
	s2 := buf.New()
	common.Must(s2.AppendSupplier(WriteUint32(x)))
	assert.Bytes(s1).Equals(s2.Bytes())
}
