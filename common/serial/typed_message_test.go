package serial_test

import (
	"testing"

	. "github.com/whatedcgveg/v2ray-core/common/serial"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestGetInstance(t *testing.T) {
	assert := assert.On(t)

	p, err := GetInstance("")
	assert.Pointer(p).IsNil()
	assert.Error(err).IsNotNil()
}
