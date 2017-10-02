package wechat

import (
	"context"

	"github.com/whatedcgveg/v2ray-core/common"
	"github.com/whatedcgveg/v2ray-core/common/dice"
	"github.com/whatedcgveg/v2ray-core/common/serial"
)

type VideoChat struct {
	sn int
}

func (vc *VideoChat) Size() int {
	return 13
}

// Write implements io.Writer.
func (vc *VideoChat) Write(b []byte) (int, error) {
	vc.sn++
	b = append(b[:0], 0xa1, 0x08)
	b = serial.IntToBytes(vc.sn, b)
	b = append(b, 0x10, 0x11, 0x18, 0x30, 0x22, 0x30)
	return 13, nil
}

func NewVideoChat(ctx context.Context, config interface{}) (interface{}, error) {
	return &VideoChat{
		sn: int(dice.RollUint16()),
	}, nil
}

func init() {
	common.Must(common.RegisterConfig((*VideoConfig)(nil), NewVideoChat))
}
