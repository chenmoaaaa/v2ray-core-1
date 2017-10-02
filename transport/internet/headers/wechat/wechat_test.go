package wechat_test

import (
	"testing"

	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
	. "github.com/whatedcgveg/v2ray-core/transport/internet/headers/wechat"
)

func TestUTPWrite(t *testing.T) {
	assert := assert.On(t)

	video := VideoChat{}

	payload := buf.NewLocal(2048)
	payload.AppendSupplier(video.Write)

	assert.Int(payload.Len()).Equals(video.Size())
}
