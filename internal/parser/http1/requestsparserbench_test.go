package http1

import (
	"github.com/indigo-web/indigo/internal/requestgen"
	"strings"
	"testing"
)

func BenchmarkHttpRequestsParser_Parse_GET(b *testing.B) {
	parser, _ := getParser()

	b.Run("5 headers", func(b *testing.B) {
		request := requestgen.Generate(strings.Repeat("a", 500), 5)
		b.SetBytes(int64(len(request)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(request)
			parser.Release()
		}
	})

	b.Run("10 headers", func(b *testing.B) {
		request := requestgen.Generate(strings.Repeat("a", 500), 10)
		b.SetBytes(int64(len(request)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(request)
			parser.Release()
		}
	})

	b.Run("50 headers", func(b *testing.B) {
		request := requestgen.Generate(strings.Repeat("a", 500), 50)
		b.SetBytes(int64(len(request)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(request)
			parser.Release()
		}
	})

	b.Run("heavily escaped uri 20 headers", func(b *testing.B) {
		request := requestgen.Generate(strings.Repeat("%20", 500), 20)
		b.SetBytes(int64(len(request)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(request)
			parser.Release()
		}
	})
}
