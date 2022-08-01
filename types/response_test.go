package types

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResponse_GrowHeaders(t *testing.T) {
	t.Run("GrowExplicitly", func(t *testing.T) {
		response := NewResponse()
		require.Empty(t, response.Headers)

		response = response.GrowHeaders(1)
		require.Equal(t, 1, cap(response.Headers))
		require.Empty(t, response.Headers)

		response = response.WithHeader("Hello", "World")
		require.Equal(t, 1, cap(response.Headers))

		response = response.GrowHeaders(2)
		require.Contains(t, response.Headers, renderHeader([]byte("Hello"), []byte("World")))
		require.Equal(t, 2, cap(response.Headers))
	})

	t.Run("GrowImplicitly", func(t *testing.T) {
		response := NewResponse()
		response = response.WithHeader("Hello", "world")
		response = response.WithHeader("Easter", "Egg")
		// slice growth algorithm may differ from version to version
		// so let's make algorithm-independent test
		currentCap := cap(response.Headers)
		response = response.GrowHeaders(1)
		require.Equal(t, currentCap, cap(response.Headers))
	})
}

func TestRenderHeader(t *testing.T) {
	key, value := []byte("Hello"), []byte("World")
	rendered := renderHeader(key, value)
	require.Equal(t, []byte("Hello: World"), rendered)
}

func TestRenderResponse(t *testing.T) {
	t.Run("NoHeadersNoBody", func(t *testing.T) {
		response := Response{
			Code:    200,
			Headers: nil,
			Body:    nil,
		}
		rendered := response.render([]byte("HTTP/1.1 "))
		// https://www.w3.org/Protocols/rfc2616/rfc2616-sec4.html
		// server MUST always set content-length header in response
		// even if no body is presented (then value 0 is set as value)
		want := "HTTP/1.1 200 OK\r\nContent-Length: 0\r\n\r\n"
		require.Equal(t, want, string(rendered))
	})

	t.Run("SingleHeader", func(t *testing.T) {
		key, value := []byte("Hello"), []byte("World")

		response := Response{
			Code: 200,
			Headers: [][]byte{
				renderHeader(key, value),
			},
			Body: nil,
		}
		rendered := response.render([]byte("HTTP/1.1 "))
		want := "HTTP/1.1 200 OK\r\nHello: World\r\nContent-Length: 0\r\n\r\n"
		require.Equal(t, want, string(rendered))
	})

	t.Run("HeadersAndBody", func(t *testing.T) {
		key, value := []byte("Hello"), []byte("World")

		response := Response{
			Code: 200,
			Headers: [][]byte{
				renderHeader(key, value),
			},
			Body: []byte("Hello, world!"),
		}
		rendered := response.render([]byte("HTTP/1.1 "))
		want := "HTTP/1.1 200 OK\r\nHello: World\r\nContent-Length: 13\r\n\r\nHello, world!"
		require.Equal(t, want, string(rendered))
	})
}
