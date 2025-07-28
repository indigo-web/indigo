package strutil

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func BenchmarkFoldSafe(b *testing.B) {
	bench := func(str string) func(b *testing.B) {
		return func(b *testing.B) {
			x, y := strings.ToLower(str), strings.ToUpper(str)
			b.SetBytes(int64(len(x)))
			b.ResetTimer()

			for range b.N {
				_ = CmpFoldSafe(x, y)
			}
		}
	}

	const sample8 = "abcdefgh"
	b.Run("8 byte", bench(sample8))
	b.Run("32 byte", bench(strings.Repeat(sample8, 4)))
	b.Run("128 byte", bench(strings.Repeat(sample8, 16)))
	b.Run("512 byte", bench(strings.Repeat(sample8, 64)))
}

func BenchmarkFoldFast(b *testing.B) {
	bench := func(str string) func(b *testing.B) {
		return func(b *testing.B) {
			x, y := strings.ToLower(str), strings.ToUpper(str)
			b.SetBytes(int64(len(x)))
			b.ResetTimer()

			for range b.N {
				_ = CmpFoldFast(x, y)
			}
		}
	}

	const sample8 = "abcdefgh"
	b.Run("8 byte", bench(sample8))
	b.Run("32 byte", bench(strings.Repeat(sample8, 4)))
	b.Run("128 byte", bench(strings.Repeat(sample8, 16)))
	b.Run("512 byte", bench(strings.Repeat(sample8, 64)))
}

func TestFoldSafe(t *testing.T) {
	require.True(t, CmpFoldSafe("HELLO", "hello"))
	require.True(t, CmpFoldSafe("\r\n\r\n", "\r\n\r\n"))
	require.False(t, CmpFoldSafe("\v\t", "\r\t"))
}

func TestFoldFast(t *testing.T) {
	require.True(t, CmpFoldFast("HELLO", "hello"))
	require.True(t, CmpFoldFast("HELLOWORLD", "helloworld"))
	require.True(t, CmpFoldFast("identity", "IDENTITY"))
	require.True(t, CmpFoldFast("\r\n\r\n", "\r\n\r\n"))
	require.False(t, CmpFoldFast("\v\t", "\r\t"))
}
