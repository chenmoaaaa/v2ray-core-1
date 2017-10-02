package protocol

import (
	"time"

	"github.com/whatedcgveg/v2ray-core/common/dice"
	"github.com/whatedcgveg/v2ray-core/common/serial"
)

type Timestamp int64

func (v Timestamp) Bytes(b []byte) []byte {
	return serial.Int64ToBytes(int64(v), b)
}

type TimestampGenerator func() Timestamp

func NowTime() Timestamp {
	return Timestamp(time.Now().Unix())
}

func NewTimestampGenerator(base Timestamp, delta int) TimestampGenerator {
	return func() Timestamp {
		rangeInDelta := dice.Roll(delta*2) - delta
		return base + Timestamp(rangeInDelta)
	}
}
