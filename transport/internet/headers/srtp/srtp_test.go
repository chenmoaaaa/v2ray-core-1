package srtp_test

import (
	"testing"

	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
	. "github.com/whatedcgveg/v2ray-core/transport/internet/headers/srtp"
)

func TestSRTPWrite(t *testing.T) {
	assert := assert.On(t)

	content := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}
	srtp := SRTP{}

	payload := buf.NewLocal(2048)
	payload.AppendSupplier(srtp.Write)
	payload.Append(content)

	assert.Int(payload.Len()).Equals(len(content) + srtp.Size())
}
