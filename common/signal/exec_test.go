package signal_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/whatedcgveg/v2ray-core/common/signal"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestErrorOrFinish2_Error(t *testing.T) {
	assert := assert.On(t)

	c1 := make(chan error, 1)
	c2 := make(chan error, 2)
	c := make(chan error, 1)

	go func() {
		c <- ErrorOrFinish2(context.Background(), c1, c2)
	}()

	c1 <- errors.New("test")
	err := <-c
	assert.String(err.Error()).Equals("test")
}

func TestErrorOrFinish2_Error2(t *testing.T) {
	assert := assert.On(t)

	c1 := make(chan error, 1)
	c2 := make(chan error, 2)
	c := make(chan error, 1)

	go func() {
		c <- ErrorOrFinish2(context.Background(), c1, c2)
	}()

	c2 <- errors.New("test")
	err := <-c
	assert.String(err.Error()).Equals("test")
}

func TestErrorOrFinish2_NoneError(t *testing.T) {
	assert := assert.On(t)

	c1 := make(chan error, 1)
	c2 := make(chan error, 2)
	c := make(chan error, 1)

	go func() {
		c <- ErrorOrFinish2(context.Background(), c1, c2)
	}()

	close(c1)
	select {
	case <-c:
		t.Fail()
	default:
	}

	close(c2)
	err := <-c
	assert.Error(err).IsNil()
}

func TestErrorOrFinish2_NoneError2(t *testing.T) {
	assert := assert.On(t)

	c1 := make(chan error, 1)
	c2 := make(chan error, 2)
	c := make(chan error, 1)

	go func() {
		c <- ErrorOrFinish2(context.Background(), c1, c2)
	}()

	close(c2)
	select {
	case <-c:
		t.Fail()
	default:
	}

	close(c1)
	err := <-c
	assert.Error(err).IsNil()
}
