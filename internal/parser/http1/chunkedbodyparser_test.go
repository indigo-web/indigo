package http1

import (
	"github.com/fakefloordiv/indigo/http/status"
	"testing"

	"github.com/fakefloordiv/indigo/internal/body"

	"github.com/fakefloordiv/indigo/settings"

	"github.com/stretchr/testify/require"
)

func bodyReader(gateway *body.Gateway, ch chan []byte) {
	var body []byte

	for {
		data := <-gateway.Data
		if data == nil {
			break
		}

		body = append(body, data...)
		gateway.Data <- nil
	}

	ch <- body
}

func finalize(gateway *body.Gateway) {
	gateway.Data <- nil
}

func testDifferentPartSizes(t *testing.T, request []byte, wantBody string, trailer bool) {
	for i := 1; i < len(request); i += 1 {
		gateway := body.NewBodyGateway()
		ch := make(chan []byte)
		go bodyReader(gateway, ch)

		parser := newChunkedBodyParser(gateway, settings.Default())

		parts := splitIntoParts(request, i)
		for j, part := range parts {
			done, extra, err := parser.Parse(part, nopDecoder, trailer)
			require.Empty(t, extra)
			require.NoErrorf(t, err, "happened with part size: %d", i)

			if done {
				require.True(t, j+1 == len(parts),
					"done before the whole request was fed")
			}
		}

		finalize(gateway)
		require.Equal(t, wantBody, string(<-ch))
	}
}

func TestChunkedBodyParser_Parse(t *testing.T) {
	chunked := []byte("d\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n")
	wantedBody := "Hello, world!But what's wrong with you?Finally am here"
	testDifferentPartSizes(t, chunked, wantedBody, false)
}

func TestChunkedBodyParser_Parse_LFOnly(t *testing.T) {
	chunked := []byte("d\nHello, world!\n1a\nBut what's wrong with you?\nf\nFinally am here\n0\n\n")
	wantedBody := "Hello, world!But what's wrong with you?Finally am here"
	testDifferentPartSizes(t, chunked, wantedBody, false)
}

func TestChunkedBodyParser_Parse_FooterHeaders(t *testing.T) {
	chunked := []byte("7\r\nMozilla\r\n9\r\nDeveloper\r\n7\r\nNetwork\r\n0\r\nExpires: date here\r\n\r\n")
	wantedBody := "MozillaDeveloperNetwork"
	testDifferentPartSizes(t, chunked, wantedBody, true)
}

func TestChunkedBodyParser_Parse_Negative(t *testing.T) {
	t.Run("BeginWithCRLF", func(t *testing.T) {
		chunked := []byte("\r\nd\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n")
		gateway := body.NewBodyGateway()
		ch := make(chan []byte)
		go bodyReader(gateway, ch)

		parser := newChunkedBodyParser(gateway, settings.Default())
		done, extra, err := parser.Parse(chunked, nopDecoder, false)
		require.True(t, done)
		require.Empty(t, extra)
		require.EqualError(t, err, status.ErrBadRequest.Error())
	})
}
