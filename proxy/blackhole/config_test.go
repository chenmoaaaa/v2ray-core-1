package blackhole_test

import (
	"bufio"
	"net/http"
	"testing"

	"github.com/whatedcgveg/v2ray-core/common/buf"
	. "github.com/whatedcgveg/v2ray-core/proxy/blackhole"
	"github.com/whatedcgveg/v2ray-core/testing/assert"
)

func TestHTTPResponse(t *testing.T) {
	assert := assert.On(t)

	buffer := buf.New()

	httpResponse := new(HTTPResponse)
	httpResponse.WriteTo(buf.NewWriter(buffer))

	reader := bufio.NewReader(buffer)
	response, err := http.ReadResponse(reader, nil)
	assert.Error(err).IsNil()
	assert.Int(response.StatusCode).Equals(403)
}
