package httptest

import (
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("simple GET", func(t *testing.T) {
		data := "GET / HTTP/1.1\r\nContent-Length: 0\r\n\r\n"
		request, err := Parse(data)
		require.NoError(t, err)
		for _, err := range Compare(request, Request{
			Method: method.GET,
			Path:   "/",
			Proto:  "HTTP/1.1",
			Headers: headers.NewFromMap(map[string][]string{
				"Content-Length": {"0"},
			}),
			Body: "",
		}) {
			assert.NoError(t, err)
		}
	})

	t.Run("with headers", func(t *testing.T) {
		data := "GET / HTTP/1.1\r\nHello: World\r\nFoo: bar\r\nContent-Length: 0\r\n\r\n"
		request, err := Parse(data)
		require.NoError(t, err)
		for _, err := range Compare(request, Request{
			Method: method.GET,
			Path:   "/",
			Proto:  "HTTP/1.1",
			Headers: headers.NewFromMap(map[string][]string{
				"Hello":          {"World"},
				"Foo":            {"bar"},
				"Content-Length": {"0"},
			}),
			Body: "",
		}) {
			assert.NoError(t, err)
		}
	})

	t.Run("body and content-length", func(t *testing.T) {
		data := "POST /greets HTTP/1.1\r\nContent-Length: 13\r\n\r\nHello, world!"
		request, err := Parse(data)
		require.NoError(t, err)
		for _, err := range Compare(request, Request{
			Method: method.POST,
			Path:   "/greets",
			Proto:  "HTTP/1.1",
			Headers: headers.NewFromMap(map[string][]string{
				"Content-Length": {"13"},
			}),
			Body: "Hello, world!",
		}) {
			assert.NoError(t, err)
		}
	})

	t.Run("body with closing connection", func(t *testing.T) {
		data := "POST /greets HTTP/1.1\r\nConnection: close\r\n\r\nHello, world!"
		request, err := Parse(data)
		require.NoError(t, err)
		for _, err := range Compare(request, Request{
			Method: method.POST,
			Path:   "/greets",
			Proto:  "HTTP/1.1",
			Headers: headers.NewFromMap(map[string][]string{
				"Connection": {"close"},
			}),
			Body: "Hello, world!",
		}) {
			assert.NoError(t, err)
		}
	})
}
