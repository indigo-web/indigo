package http

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func allocs(str string) int {
	return int(testing.Benchmark(func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Escape(str)
		}
	}).AllocsPerOp())
}

func TestEscape(t *testing.T) {
	t.Run("normal path", func(t *testing.T) {
		require.Equal(t, "/", Escape("/"))
		require.Zero(t, allocs("/hello-world"))
	})

	t.Run("with nonprintable", func(t *testing.T) {
		require.Equal(t, `/\0\n\?`, Escape("/\x00\n\x7f"))
		require.Equal(t, 1, allocs("/\x00\n\x7f"))
	})

	t.Run("all nonprintalbe are included", func(t *testing.T) {
		for i := byte(0); i < 0xff; i++ {
			assert.Truef(
				t, isASCIIPrintable(i) == (escapeByte(i) == 0),
				"not true for %s", strconv.Quote(string(i)),
			)
		}
	})
}
