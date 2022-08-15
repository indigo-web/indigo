package http1

import (
	"github.com/stretchr/testify/require"
	"indigo/internal"
	"indigo/settings"
	"testing"
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

func TestChunkedBodyParser_Parse(t *testing.T) {
	gateway := internal.NewBodyGateway()
	ch := make(chan []byte)
	go bodyReader(gateway, ch)

	chunked := []byte("d\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n")
	wantedBody := "Hello, world!But what's wrong with you?Finally am here"
	parser := newChunkedBodyParser(gateway, settings.Default())

	done, extra, err := parser.Parse(chunked)
	require.True(t, done)
	require.Empty(t, extra)
	require.NoError(t, err)
	finalize(gateway)
	require.Equal(t, wantedBody, string(<-ch))
}
