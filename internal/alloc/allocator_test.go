package alloc

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func pushSegment(t *testing.T, allocator *Allocator, text string) {
	ok := allocator.Append([]byte(text))
	require.True(t, ok)
	segment := allocator.Finish()
	require.Equal(t, text, string(segment))
}

func TestAllocator(t *testing.T) {
	t.Run("NoOverflow", func(t *testing.T) {
		rawAllocator := NewAllocator(10, 20)
		allocator := &rawAllocator
		pushSegment(t, allocator, "Hello")
		pushSegment(t, allocator, "Here")
	})

	t.Run("YesOverflow", func(t *testing.T) {
		rawAllocator := NewAllocator(10, 20)
		allocator := &rawAllocator
		// "Hello, World!" is 13 characters length, so it will force the allocator
		// to grow an underlying slice
		pushSegment(t, allocator, "Hello, ")
		pushSegment(t, allocator, "World!")
	})

	t.Run("SizeLimitOverflow", func(t *testing.T) {
		rawAllocator := NewAllocator(10, 20)
		allocator := &rawAllocator
		pushSegment(t, allocator, "Hello, ")
		pushSegment(t, allocator, "World!")
		pushSegment(t, allocator, "Lorem ")
		// at this point, we have reached 19 elements in underlying slice
		ok := allocator.Append([]byte("overflow"))
		require.False(t, ok)
	})
}
