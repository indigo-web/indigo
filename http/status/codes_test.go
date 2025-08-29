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

	// it used to panic when passing codes <100
	require.Equal(t, "", StringCode(99))
}

func Benchmark(b *testing.B) {
	code := KnownCodes[rand.IntN(len(KnownCodes))]
	b.ResetTimer()

	for range b.N {
		_ = StringCode(code)
	}
}

func TestString(t *testing.T) {
	require.Equal(t, "Nonstandard", String(1))
	require.Equal(t, "OK", String(200))
	require.Equal(t, "Nonstandard", String(maxCodeValue))
}
