package huffman

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestHuffman(t *testing.T) {
	test := func(t *testing.T, str string) {
		decompressed, ok := Decompress(Compress(str, nil))
		require.True(t, ok)
		require.Equal(t, str, decompressed)
	}

	t.Run("single frequent letter", func(t *testing.T) {
		test(t, "a")
	})

	t.Run("single infrequent letter", func(t *testing.T) {
		test(t, "\x00")
	})

	t.Run("short string", func(t *testing.T) {
		test(t, "abcdef")
	})

	t.Run("long string", func(t *testing.T) {
		test(t, strings.Repeat("abcdef", 100))
	})

	t.Run("long string of infrequent chars", func(t *testing.T) {
		test(t, strings.Repeat("\x00\xfa\xfb\xfc\xfd", 100))
	})

	t.Run("invalid code", func(t *testing.T) {
		// single bit in the end is zero
		_, ok := Decompress([]byte{0b11111111, 0b11111111, 0b11111001, 0b10111011})
		require.False(t, ok)

		// has no free bits at all
		_, ok = Decompress([]byte{0b00011000, 0b11000110, 0b00111000, 0b11100011})
		require.True(t, ok)
	})
}
