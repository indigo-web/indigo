package http1

import (
	"github.com/fakefloordiv/indigo/internal/body"
	"github.com/fakefloordiv/indigo/settings"
	"strings"
	"testing"
)

func nopBodyReader(gateway *body.Gateway) {
	for {
		<-gateway.Data
		gateway.Data <- nil
	}
}

func BenchmarkChunkedBodyParser(b *testing.B) {
	chunkedEnd := "0\r\n\r\n"
	smallChunked := []byte(
		"d\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n" + chunkedEnd,
	)
	mediumChunked := []byte(
		strings.Repeat("1a\r\nBut what's wrong with you?\r\n", 15) + chunkedEnd,
	)
	bigChunked := []byte(
		strings.Repeat("1a\r\nBut what's wrong with you?\r\n", 100) + chunkedEnd,
	)

	gateway := body.NewBodyGateway()
	parser := newChunkedBodyParser(gateway, settings.Default())
	go nopBodyReader(gateway)

	b.Run("Small_Example", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(smallChunked, nopDecoder, false)
		}
	})

	b.Run("Medium_15Repeats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(mediumChunked, nopDecoder, false)
		}
	})

	b.Run("Big_100Repeats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(bigChunked, nopDecoder, false)
		}
	})
}
