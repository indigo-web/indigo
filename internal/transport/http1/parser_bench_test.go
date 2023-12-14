package http1

import (
	"github.com/indigo-web/indigo/internal/requestgen"
	"strings"
	"testing"
)

func BenchmarkParser(b *testing.B) {
	parser, request := getParser()

	b.Run("5 headers", func(b *testing.B) {
		data := requestgen.Generate(strings.Repeat("a", 500), 5)
		b.SetBytes(int64(len(data)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			b.ReportAllocs()
			_ = request.Clear()
		}
	})

	b.Run("10 headers", func(b *testing.B) {
		data := requestgen.Generate(strings.Repeat("a", 500), 10)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			_ = request.Clear()
		}
	})

	b.Run("50 headers", func(b *testing.B) {
		data := requestgen.Generate(strings.Repeat("a", 500), 50)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			_ = request.Clear()
		}
	})

	b.Run("heavily escaped uri 10 headers", func(b *testing.B) {
		data := requestgen.Generate(strings.Repeat("%20", 500), 10)
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(data)
			_ = request.Clear()
		}
	})
}
