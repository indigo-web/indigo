package arena

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func pushSegment(t *testing.T, arena *Arena, text string) {
	ok := arena.Append([]byte(text))
	require.True(t, ok)
	segment := arena.Finish()
	require.Equal(t, text, string(segment))
}

func TestArena(t *testing.T) {
	t.Run("NoOverflow", func(t *testing.T) {
		rawArena := NewArena(10, 20)
		arena := &rawArena
		pushSegment(t, arena, "Hello")
		pushSegment(t, arena, "Here")
	})

	t.Run("YesOverflow", func(t *testing.T) {
		rawArena := NewArena(10, 20)
		Arena := &rawArena
		// "Hello, World!" is 13 characters length, so it will force the Arena
		// to grow an underlying slice
		pushSegment(t, Arena, "Hello, ")
		pushSegment(t, Arena, "World!")
	})

	t.Run("SizeLimitOverflow", func(t *testing.T) {
		rawArena := NewArena(10, 20)
		Arena := &rawArena
		pushSegment(t, Arena, "Hello, ")
		pushSegment(t, Arena, "World!")
		pushSegment(t, Arena, "Lorem ")
		// at this point, we have reached 19 elements in underlying slice
		ok := Arena.Append([]byte("overflow"))
		require.False(t, ok)
	})
}
