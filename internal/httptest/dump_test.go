package httptest

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/internal/protocol/http1"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDump(t *testing.T) {
	q := query.New(keyvalue.New(), config.Default())
	q.Update([]byte("hello=world&foo=bar"))

	cfg := config.Default()
	client := dummy.NewCircularClient([]byte("Hello, world!")).OneTime()
	body := http1.NewBody(client, construct.Chunked(cfg.Body), cfg.Body)
	request := construct.Request(config.Default(), client, body)
	request.Headers = headers.New().
		Add("hello", "world").
		Add("foo", "bar")
	request.Query = q
	request.Method = method.GET
	request.Path = "/"
	request.Proto = proto.HTTP11
	request.ContentLength = 13
	body.Reset(request)
	dumped, err := Dump(request)
	require.NoError(t, err)
	want := "GET /?hello=world&foo=bar HTTP/1.1\r\nhello: world\r\nfoo: bar\r\nContent-Length: 13\r\n\r\nHello, world!"
	require.Equal(t, want, dumped)
}
