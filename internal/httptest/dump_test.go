package httptest

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/datastruct"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/internal/transport/http1"
	"github.com/indigo-web/indigo/settings"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDump(t *testing.T) {
	hdrs := headers.New()
	hdrs.Add("hello", "world")
	hdrs.Add("foo", "bar")
	hdrs.Add("Content-Length", "13")
	q := query.NewQuery(datastruct.NewKeyValue())
	q.Set([]byte("hello=world&foo=bar"))
	client := dummy.NewCircularClient([]byte("Hello, world!"))
	client.OneTime()
	body := http1.NewBody(client, nil, settings.Default().Body)
	request := http.NewRequest(hdrs, q, nil, dummy.NewNopConn(), body, nil)
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
