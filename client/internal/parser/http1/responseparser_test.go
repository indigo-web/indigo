package http1

import (
	"github.com/indigo-web/indigo/client"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/utils/buffer"
	"github.com/stretchr/testify/require"
	"testing"
)

func compareResponse(t *testing.T, want, got client.Response) {
	require.Equal(t, want.Protocol, got.Protocol)
	require.Equal(t, int(want.Code), int(got.Code))
	if len(want.Status) > 0 {
		require.Equal(t, want.Status, got.Status)
	}

	for _, key := range want.Headers.Keys() {
		require.True(t, got.Headers.Has(key))
		require.Equal(t, want.Headers.Values(key), got.Headers.Values(key))
	}
}

func TestResponseParser(t *testing.T) {
	parser := NewParser(
		*buffer.NewBuffer[byte](0, 4096), *buffer.NewBuffer[byte](0, 4096),
	)

	t.Run("simple response", func(t *testing.T) {
		data := "HTTP/1.1 200 OK\r\n\r\n"
		parser.Init(headers.NewHeaders(), nil)
		headersCompleted, rest, err := parser.Parse([]byte(data))
		require.NoError(t, err)
		require.True(t, headersCompleted)
		require.Empty(t, rest)
		compareResponse(t, client.Response{
			Protocol: proto.HTTP11,
			Code:     status.OK,
			Status:   "OK",
			Headers:  headers.NewHeaders(),
		}, parser.Response())
	})

	t.Run("response with headers", func(t *testing.T) {
		data := "HTTP/1.1 200 OK\r\nHello: world\r\nhello: nether\r\n\r\n"
		parser.Init(headers.NewHeaders(), nil)
		headersCompleted, rest, err := parser.Parse([]byte(data))
		require.NoError(t, err)
		require.True(t, headersCompleted)
		require.Empty(t, rest)
		compareResponse(t, client.Response{
			Protocol: proto.HTTP11,
			Code:     status.OK,
			Status:   "OK",
			Headers: headers.FromMap(map[string][]string{
				"hello": {"world", "nether"},
			}),
		}, parser.Response())
	})
}
