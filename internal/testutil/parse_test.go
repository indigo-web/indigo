package testutil

import (
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParseRequest(t *testing.T) {
	raw := "GET / HTTP/1.1\r\nHello: world\r\nContent-Length: 13\r\n\r\nHello, world!"
	request, err := ParseRequest(raw)
	require.NoError(t, err)
	require.Equal(t, method.GET, request.Method)
	require.Equal(t, "/", request.Path)
	require.Equal(t, proto.HTTP11, request.Protocol)
	require.Equal(t, "world", request.Headers.Value("hello"))
	require.Equal(t, "Hello, world!", request.Body)
}
