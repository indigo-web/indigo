package buffer

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func pushSegment(t *testing.T, buff Buffer, text string) Buffer {
	ok := buff.Append([]byte(text))
	require.True(t, ok)
	segment := buff.Finish()
	require.Equal(t, text, string(segment))
	return buff
}

func BenchmarkBuffer(b *testing.B) {
	buff := New(1024, 4096)
	smallString := []byte(strings.Repeat("a", 1023))
	bigString := []byte(strings.Repeat("a", 4095))

	b.Run("no overflow", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(smallString)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = buff.Append(smallString)
			buff.Clear()
		}
	})

	b.Run("with overflow", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(bigString)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = buff.Append(bigString)
			buff.Clear()
			buff.memory = buff.memory[0:0:1024]
		}
	})
}

func TestBuffer(t *testing.T) {
	t.Run("no overflow", func(t *testing.T) {
		buff := New(10, 20)
		buff = pushSegment(t, buff, "Hello")
		buff = pushSegment(t, buff, "Here")
	})

	t.Run("with overflow", func(t *testing.T) {
		buff := New(10, 20)
		// "Hello, World!" is 13 characters length, so it will force the Buffer
		// to grow an underlying slice
		buff = pushSegment(t, buff, "Hello, ")
		buff = pushSegment(t, buff, "World!")
	})

	t.Run("overflow over the limit", func(t *testing.T) {
		buff := New(10, 20)
		buff = pushSegment(t, buff, "Hello, ")
		buff = pushSegment(t, buff, "World!")
		buff = pushSegment(t, buff, "Lorem ")
		// at this point, we have reached 19 elements in underlying slice
		ok := buff.Append([]byte("overflow"))
		require.False(t, ok)
	})

	t.Run("segment length", func(t *testing.T) {
		buff := New(10, 20)
		require.True(t, buff.Append([]byte("Hello, ")))
		require.True(t, buff.Append([]byte("World!")))
		require.Equal(t, 13, buff.SegmentLength())
	})

	t.Run("discard bytes", func(t *testing.T) {
		testDiscard(t, 13)
		testDiscard(t, 50)
	})

	t.Run("discard segment", func(t *testing.T) {
		buff := New(50, 50)
		require.True(t, buff.Append([]byte("Hello")))
		buff.Finish()
		require.True(t, buff.Append([]byte("World")))
		buff.Discard(0)
		require.Equal(t, "Hello", string(buff.memory))
	})

	t.Run("truncate", func(t *testing.T) {
		testTrunc(t, 1)
		testTrunc(t, 5)
	})
}

func testDiscard(t *testing.T, n int) {
	buff := New(10, 20)
	require.True(t, buff.Append([]byte("Hello, world!")))
	segment := buff.Finish()
	buff.Discard(n)
	require.True(t, buff.Append([]byte("Hello!")))
	newSegment := buff.Finish()
	require.Equal(t, "Hello!", string(newSegment))
	require.Equal(t, "Hello! world!", string(segment))
}

func testTrunc(t *testing.T, n int) {
	buff := New(10, 20)
	require.True(t, buff.Append([]byte("Hello, world!")))
	segment := buff.Finish()
	require.True(t, buff.Append([]byte("Hi?")))
	buff.Trunc(n)
	require.Equal(t, "Hello, world!", string(segment))

	orig := "Hi?"
	if n > len(orig) {
		n = len(orig)
	}

	require.Equal(t, orig[:len(orig)-n], string(buff.Finish()))
}
