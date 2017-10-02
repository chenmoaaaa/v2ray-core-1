package http_test

import (
	"bufio"
	"net/http"
	"strings"
	"testing"

	. "github.com/whatedcgveg/v2ray-core/proxy/http"
	"github.com/whatedcgveg/v2ray-core/testing/assert"

	_ "github.com/whatedcgveg/v2ray-core/transport/internet/tcp"
)

func TestHopByHopHeadersStrip(t *testing.T) {
	assert := assert.On(t)

	rawRequest := `GET /pkg/net/http/ HTTP/1.1
Host: golang.org
Connection: keep-alive,Foo, Bar
Foo: foo
Bar: bar
Proxy-Connection: keep-alive
Proxy-Authenticate: abc
User-Agent: Mozilla/5.0 (Macintosh; U; Intel Mac OS X; de-de) AppleWebKit/523.10.3 (KHTML, like Gecko) Version/3.0.4 Safari/523.10
Accept-Encoding: gzip
Accept-Charset: ISO-8859-1,UTF-8;q=0.7,*;q=0.7
Cache-Control: no-cache
Accept-Language: de,en;q=0.7,en-us;q=0.3

`
	b := bufio.NewReader(strings.NewReader(rawRequest))
	req, err := http.ReadRequest(b)
	assert.Error(err).IsNil()
	assert.String(req.Header.Get("Foo")).Equals("foo")
	assert.String(req.Header.Get("Bar")).Equals("bar")
	assert.String(req.Header.Get("Connection")).Equals("keep-alive,Foo, Bar")
	assert.String(req.Header.Get("Proxy-Connection")).Equals("keep-alive")
	assert.String(req.Header.Get("Proxy-Authenticate")).Equals("abc")

	StripHopByHopHeaders(req.Header)
	assert.String(req.Header.Get("Connection")).IsEmpty()
	assert.String(req.Header.Get("Foo")).IsEmpty()
	assert.String(req.Header.Get("Bar")).IsEmpty()
	assert.String(req.Header.Get("Proxy-Connection")).IsEmpty()
	assert.String(req.Header.Get("Proxy-Authenticate")).IsEmpty()
}
