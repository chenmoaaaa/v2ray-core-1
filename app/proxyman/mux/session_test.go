package mux_test

import (
	"testing"

	. "github.com/whatedcgveg/v2ray-core/app/proxyman/mux"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestSessionManagerAdd(t *testing.T) {
	assert := assert.On(t)

	m := NewSessionManager()

	s := m.Allocate()
	assert.Uint16(s.ID).Equals(1)

	s = m.Allocate()
	assert.Uint16(s.ID).Equals(2)

	s = &Session{
		ID: 4,
	}
	m.Add(s)
	assert.Uint16(s.ID).Equals(4)
}

func TestSessionManagerClose(t *testing.T) {
	assert := assert.On(t)

	m := NewSessionManager()
	s := m.Allocate()

	assert.Bool(m.CloseIfNoSession()).IsFalse()
	m.Remove(s.ID)
	assert.Bool(m.CloseIfNoSession()).IsTrue()
}
