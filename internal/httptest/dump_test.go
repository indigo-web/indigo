package httptest

import (
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/initialize"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/settings"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDump(t *testing.T) {
	q := query.NewQuery(keyvalue.New())
	q.Set([]byte("hello=world&foo=bar"))

	client := dummy.NewCircularClient([]byte("Hello, world!")).OneTime()
	body := initialize.NewBody(client, settings.Default().Body)
	request := initialize.NewRequest(settings.Default(), dummy.NewNopConn(), body)
	request.Headers = headers.New().
		Add("hello", "world").
		Add("foo", "bar")
	request.Query = q
	request.Method = method.GET
	request.Path = "/"
	request.Proto = proto.HTTP11
	request.ContentLength = 13
	body.Init(request)
	dumped, err := Dump(request)
	require.NoError(t, err)
	want := "GET /?hello=world&foo=bar HTTP/1.1\r\nhello: world\r\nfoo: bar\r\nContent-Length: 13\r\n\r\nHello, world!"
	require.Equal(t, want, dumped)
}
