package strutil

import (
	"github.com/flrdv/uf"
	"github.com/stretchr/testify/require"
	"iter"
	"testing"
)

type strpair struct {
	K, V string
}

func collect(i iter.Seq2[string, string]) (pairs []strpair) {
	for k, v := range i {
		pairs = append(pairs, strpair{k, v})
	}

	return pairs
}

func strdup(str string) string {
	return string(uf.S2B(str))
}

func TestWalkKV(t *testing.T) {
	t.Run("single value", func(t *testing.T) {
		values := collect(WalkKV("abc"))
		require.Equal(t, 1, len(values))
		require.Equal(t, strpair{"abc", ""}, values[0])
	})

	t.Run("single pair", func(t *testing.T) {
		values := collect(WalkKV("abc=cba"))
		require.Equal(t, 1, len(values))
		require.Equal(t, strpair{"abc", "cba"}, values[0])
	})

	t.Run("multiple pairs", func(t *testing.T) {
		values := collect(WalkKV("abc=cba;hello=world;"))
		require.Equal(t, 3, len(values))
		require.Equal(t, strpair{"abc", "cba"}, values[0])
		require.Equal(t, strpair{"hello", "world"}, values[1])
		require.Equal(t, strpair{"", ""}, values[2])
	})

	t.Run("codings", func(t *testing.T) {
		values := collect(WalkKV(strdup("abc=cba; hello=\"world\"; k%20ey=value%21")))
		require.Equal(t, 3, len(values))
		require.Equal(t, strpair{"abc", "cba"}, values[0])
		require.Equal(t, strpair{"hello", "world"}, values[1])
		require.Equal(t, strpair{"k%20ey", "value%21"}, values[2])
	})
}
