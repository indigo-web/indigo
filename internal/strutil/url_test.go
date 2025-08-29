package strutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestURLDecode(t *testing.T) {
	t.Run("base", func(t *testing.T) {
		res, ok := URLDecode("%61")
		require.True(t, ok)
		require.Equal(t, "a", res)

		for i, tc := range []string{"abc", "%61bc", "a%62c", "ab%63", "%61%62%63"} {
			res, ok = URLDecode(tc)
			require.True(t, ok, i)
			require.Equal(t, "abc", res, i)
		}
	})

	t.Run("unsafe char", func(t *testing.T) {
		res, ok := URLDecode("%61%2f")
		require.True(t, ok)
		require.Equal(t, "a%2f", res)
	})

	t.Run("unsafe normalization", func(t *testing.T) {
		res, ok := URLDecode("%61%2F")
		require.True(t, ok)
		require.Equal(t, "a%2f", res)
	})
}
