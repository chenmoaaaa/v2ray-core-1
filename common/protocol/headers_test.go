package protocol_test

import (
	"testing"

	. "github.com/whatedcgveg/v2ray-core/common/protocol"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestRequestOptionSet(t *testing.T) {
	assert := assert.On(t)

	var option RequestOption
	assert.Bool(option.Has(RequestOptionChunkStream)).IsFalse()

	option.Set(RequestOptionChunkStream)
	assert.Bool(option.Has(RequestOptionChunkStream)).IsTrue()

	option.Set(RequestOptionChunkMasking)
	assert.Bool(option.Has(RequestOptionChunkMasking)).IsTrue()
	assert.Bool(option.Has(RequestOptionChunkStream)).IsTrue()
}

func TestRequestOptionClear(t *testing.T) {
	assert := assert.On(t)

	var option RequestOption
	option.Set(RequestOptionChunkStream)
	option.Set(RequestOptionChunkMasking)

	option.Clear(RequestOptionChunkStream)
	assert.Bool(option.Has(RequestOptionChunkStream)).IsFalse()
	assert.Bool(option.Has(RequestOptionChunkMasking)).IsTrue()
}
