package hexconv

import (
	"strings"
	"testing"
)

func benchLocal(b *testing.B, str string) {
	b.SetBytes(int64(len(str)))
	b.ResetTimer()

	for range b.N {
		var result uint64

		for j := range str {
			result = (result << 4) | uint64(Halfbyte[str[j]])
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
