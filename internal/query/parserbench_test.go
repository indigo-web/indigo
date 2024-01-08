package query

import (
	"fmt"
	"github.com/indigo-web/indigo/http/headers"
	"strings"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	params := headers.NewPrealloc(100)

	singlePair := []byte(generatePairs(1))
	manyPairs := []byte(generatePairs(20))
	veryManyPairs := []byte(generatePairs(100))

	b.Run("single pair", benchmark(singlePair, params))
	b.Run("20 pairs", benchmark(manyPairs, params))
	b.Run("100 pairs", benchmark(veryManyPairs, params))
}

func benchmark(data []byte, params headers.Headers) func(b *testing.B) {
	return func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = Parse(data, params)
			params.Clear()
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
