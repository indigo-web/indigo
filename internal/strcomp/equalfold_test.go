package strcomp

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEqualFold(t *testing.T) {
	t.Run("equal strings", func(t *testing.T) {
		require.True(t, EqualFold("abc", "abc"))
	})

	t.Run("different cases", func(t *testing.T) {
		require.True(t, EqualFold("abc", "ABC"))
		require.True(t, EqualFold("ABC", "abc"))
		require.True(t, EqualFold("aBc", "AbC"))
	})

	t.Run("different strings equal length", func(t *testing.T) {
		require.False(t, EqualFold("abc", "def"))
	})

	t.Run("different strings by length", func(t *testing.T) {
		require.False(t, EqualFold("abc", "define"))
	})
}
