package tcp

import "github.com/whatedcgveg/v2ray-core/common/errors"

func newError(values ...interface{}) *errors.Error {
	return errors.New(values...).Path("Transport", "Internet", "TCP")
}
