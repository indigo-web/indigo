package cookie

import (
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("single pair", func(t *testing.T) {
		jar := keyvalue.New()
		require.NoError(t, Parse(jar, "a=b"))
		require.Equal(t, "b", jar.Value("a"))
		require.NoError(t, Parse(jar.Clear(), "a=b;"))
		require.Equal(t, "b", jar.Value("a"))
		require.NoError(t, Parse(jar.Clear(), "a=b; "))
		require.Equal(t, "b", jar.Value("a"))
	})

	t.Run("multiple pairs", func(t *testing.T) {
		jar := keyvalue.New()
		require.NoError(t, Parse(jar, "hello=world; men=in black"))
		require.Equal(t, "world", jar.Value("hello"))
		require.Equal(t, "in black", jar.Value("men"))
	})
}
