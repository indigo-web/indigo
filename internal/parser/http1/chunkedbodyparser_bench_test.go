package http1

import (
	"strings"
	"testing"

	"github.com/indigo-web/indigo/settings"
)

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

	parser := NewChunkedBodyParser(settings.Default().Body)

	b.Run("Small_Example", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(smallChunked, false)
		}
	})

	b.Run("Medium_15Repeats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(mediumChunked, false)
		}
	})

	b.Run("Big_100Repeats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(bigChunked, false)
		}
	})
}
