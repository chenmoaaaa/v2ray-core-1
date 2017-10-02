package kcp_test

import (
	"testing"

	"github.com/whatedcgveg/v2ray-core/testing/assert"
	. "github.com/whatedcgveg/v2ray-core/transport/internet/kcp"
)

func TestSimpleAuthenticator(t *testing.T) {
	assert := assert.On(t)

	cache := make([]byte, 512)

	payload := []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g'}

	auth := NewSimpleAuthenticator()
	b := auth.Seal(cache[:0], nil, payload, nil)
	c, err := auth.Open(cache[:0], nil, b, nil)
	assert.Error(err).IsNil()
	assert.Bytes(c).Equals(payload)
}

func TestSimpleAuthenticator2(t *testing.T) {
	assert := assert.On(t)

	cache := make([]byte, 512)

	payload := []byte{'a', 'b'}

	auth := NewSimpleAuthenticator()
	b := auth.Seal(cache[:0], nil, payload, nil)
	c, err := auth.Open(cache[:0], nil, b, nil)
	assert.Error(err).IsNil()
	assert.Bytes(c).Equals(payload)
}
