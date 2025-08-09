package status

import (
	"math/rand/v2"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	for _, code := range KnownCodes {
		require.Equal(t, strconv.Itoa(int(code)), StringCode(code))
	}
}

func Benchmark(b *testing.B) {
	code := KnownCodes[rand.IntN(len(KnownCodes))]
	b.ResetTimer()

	for range b.N {
		_ = StringCode(code)
	}
}
