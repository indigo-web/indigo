package qparams

import (
	"fmt"
	"github.com/indigo-web/indigo/internal/urlencoded"
	"strings"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	singlePair := []byte(generatePairs(1))
	manyPairs := []byte(generatePairs(20))
	veryManyPairs := []byte(generatePairs(100))

	b.Run("single pair", benchmark(singlePair))
	b.Run("20 pairs", benchmark(manyPairs))
	b.Run("100 pairs", benchmark(veryManyPairs))
}

func benchmark(data []byte) func(b *testing.B) {
	return func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = Parse(data, func(string, string) {}, urlencoded.Decode)
		}
	}
}

func generatePairs(n int) string {
	var result string
	const (
		key   = "something"
		value = "somewhere"
	)

	for i := 0; i < n; i++ {
		result += fmt.Sprintf("%s%d=%s&", key, i, value)
	}

	return strings.TrimSuffix(result, "&")
}
