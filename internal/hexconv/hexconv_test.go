package hexconv

import (
	"github.com/indigo-web/utils/hex"
	"strings"
	"testing"
)

func benchLocal(b *testing.B, str string) {
	b.SetBytes(int64(len(str)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result uint64

		for j := range str {
			result = (result << 4) | uint64(Parse(str[j]))
		}
	}
}

func benchUtils(b *testing.B, str string) {
	b.SetBytes(int64(len(str)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result uint64

		for j := range str {
			result = (result << 4) | uint64(hex.Un(str[j]))
		}
	}
}

func BenchmarkParse(b *testing.B) {
	b.Run("short", func(b *testing.B) {
		benchLocal(b, "123456789abcdef")
	})

	b.Run("long", func(b *testing.B) {
		benchLocal(b, strings.Repeat("123456789abcdef", 100))
	})
}

func BenchmarkParseUtils(b *testing.B) {
	// this benchmark uses hex package from utils/hex
	b.Run("short", func(b *testing.B) {
		benchUtils(b, "123456789abcdef")
	})

	b.Run("long", func(b *testing.B) {
		benchUtils(b, strings.Repeat("123456789abcdef", 100))
	})
}
