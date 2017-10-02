package utp_test

import (
	"testing"

	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
	. "github.com/whatedcgveg/v2ray-core/transport/internet/headers/utp"
)

func TestUTPWrite(t *testing.T) {
	assert := assert.On(t)

	content := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	utp := UTP{}

	payload := buf.NewLocal(2048)
	payload.AppendSupplier(utp.Write)
	payload.Append(content)

	assert.Int(payload.Len()).Equals(len(content) + utp.Size())
}
