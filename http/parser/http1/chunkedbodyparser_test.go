package http1

import (
	"indigo/errors"
	"indigo/internal"
	"indigo/settings"
	"testing"

	"github.com/stretchr/testify/require"
)

func bodyReader(gateway *internal.BodyGateway, ch chan []byte) {
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

func finalize(gateway *internal.BodyGateway) {
	gateway.Data <- nil
}

func testDifferentPartSizes(t *testing.T, request []byte, wantBody string) {
	for i := 1; i < len(request); i += 1 {
		gateway := internal.NewBodyGateway()
		ch := make(chan []byte)
		go bodyReader(gateway, ch)

		parser := newChunkedBodyParser(gateway, settings.Default())

		parts := splitIntoParts(request, i)
		for j, part := range parts {
			done, extra, err := parser.Parse(part)
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
	testDifferentPartSizes(t, chunked, wantedBody)
}

func TestChunkedBodyParser_Parse_LFOnly(t *testing.T) {
	chunked := []byte("d\nHello, world!\n1a\nBut what's wrong with you?\nf\nFinally am here\n0\n\n")
	wantedBody := "Hello, world!But what's wrong with you?Finally am here"
	testDifferentPartSizes(t, chunked, wantedBody)
}

func TestChunkedBodyParser_Parse_Negative(t *testing.T) {
	t.Run("BeginWithCRLF", func(t *testing.T) {
		chunked := []byte("\r\nd\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n")
		gateway := internal.NewBodyGateway()
		ch := make(chan []byte)
		go bodyReader(gateway, ch)

		parser := newChunkedBodyParser(gateway, settings.Default())
		done, extra, err := parser.Parse(chunked)
		require.True(t, done)
		require.Empty(t, extra)
		require.EqualError(t, err, errors.ErrBadRequest.Error())
	})
}
