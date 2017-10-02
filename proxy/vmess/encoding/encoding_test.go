package encoding_test

import (
	"context"
	"testing"

	"github.com/whatedcgveg/v2ray-core/common/buf"
	"github.com/whatedcgveg/v2ray-core/common/net"
	"github.com/whatedcgveg/v2ray-core/common/protocol"
	"github.com/whatedcgveg/v2ray-core/common/serial"
	"github.com/whatedcgveg/v2ray-core/common/uuid"
	"github.com/whatedcgveg/v2ray-core/proxy/vmess"
	. "github.com/whatedcgveg/v2ray-core/proxy/vmess/encoding"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestRequestSerialization(t *testing.T) {
	assert := assert.On(t)

	user := &protocol.User{
		Level: 0,
		Email: "test@v2ray.com",
	}
	account := &vmess.Account{
		Id:      uuid.New().String(),
		AlterId: 0,
	}
	user.Account = serial.ToTypedMessage(account)

	expectedRequest := &protocol.RequestHeader{
		Version:  1,
		User:     user,
		Command:  protocol.RequestCommandTCP,
		Address:  net.DomainAddress("www.v2ray.com"),
		Port:     net.Port(443),
		Security: protocol.Security(protocol.SecurityType_AES128_GCM),
	}

	buffer := buf.New()
	client := NewClientSession(protocol.DefaultIDHash)
	client.EncodeRequestHeader(expectedRequest, buffer)

	buffer2 := buf.New()
	buffer2.Append(buffer.Bytes())

	ctx, cancel := context.WithCancel(context.Background())
	sessionHistory := NewSessionHistory(ctx)

	userValidator := vmess.NewTimedUserValidator(ctx, protocol.DefaultIDHash)
	userValidator.Add(user)

	server := NewServerSession(userValidator, sessionHistory)
	actualRequest, err := server.DecodeRequestHeader(buffer)
	assert.Error(err).IsNil()

	assert.Byte(expectedRequest.Version).Equals(actualRequest.Version)
	assert.Byte(byte(expectedRequest.Command)).Equals(byte(actualRequest.Command))
	assert.Byte(byte(expectedRequest.Option)).Equals(byte(actualRequest.Option))
	assert.Address(expectedRequest.Address).Equals(actualRequest.Address)
	assert.Port(expectedRequest.Port).Equals(actualRequest.Port)
	assert.Byte(byte(expectedRequest.Security)).Equals(byte(actualRequest.Security))

	_, err = server.DecodeRequestHeader(buffer2)
	// anti replay attack
	assert.Error(err).IsNotNil()

	cancel()
}
