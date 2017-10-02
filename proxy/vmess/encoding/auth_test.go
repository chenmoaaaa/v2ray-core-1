package encoding_test

import (
	"crypto/rand"
	"testing"

	"github.com/whatedcgveg/v2ray-core/common"
	. "github.com/whatedcgveg/v2ray-core/proxy/vmess/encoding"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestFnvAuth(t *testing.T) {
	assert := assert.On(t)
	fnvAuth := new(FnvAuthenticator)

	expectedText := make([]byte, 256)
	_, err := rand.Read(expectedText)
	common.Must(err)

	buffer := make([]byte, 512)
	b := fnvAuth.Seal(buffer[:0], nil, expectedText, nil)
	b, err = fnvAuth.Open(buffer[:0], nil, b, nil)
	assert.Error(err).IsNil()
	assert.Int(len(b)).Equals(256)
	assert.Bytes(b).Equals(expectedText)
}
