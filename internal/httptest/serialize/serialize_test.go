package serialize

import (
	"github.com/indigo-web/indigo/internal/httptest/parse"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDump(t *testing.T) {
	t.Run("Content-Length", func(t *testing.T) {
		input := "GET /?hello=world&foo=bar HTTP/1.1\r\n" +
			"Hello: world\r\n" +
			"Lorem: ipsum\r\n" +
			"Transfer-Encoding: identity\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"Hello, world!"
		request, err := parse.HTTP11Request(input)
		require.NoError(t, err)
		serialized, err := Request(request)
		require.NoError(t, err)
		require.Equal(t, input, serialized)
	})

	t.Run("Chunked encoding", func(t *testing.T) {
		headers := "GET /?hello=world&foo=bar HTTP/1.1\r\n" +
			"Hello: world\r\n" +
			"Lorem: ipsum\r\n" +
			"Content-Encoding: gzip\r\n" +
			"Transfer-Encoding: chunked\r\n" +
			"\r\n"
		chunkedBody := "d\r\nHello, world!\r\n0\r\n\r\n"
		wantBody := "Hello, world!"
		request, err := parse.HTTP11Request(headers + chunkedBody)
		require.NoError(t, err)
		serialized, err := Request(request)
		require.NoError(t, err)
		require.Equal(t, headers+wantBody, serialized)
	})
}
