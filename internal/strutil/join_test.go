package strutil

import (
	"iter"
	"testing"

	"github.com/stretchr/testify/require"
)

func asIterator(elems ...string) iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, elem := range elems {
			yield(elem)
		}
	}
}

func TestJoin(t *testing.T) {
	str := Join(asIterator(), ", ")
	require.Empty(t, str)

	str = Join(asIterator("hello"), ", ")
	require.Equal(t, "hello", str)

	str = Join(asIterator("hello", "world"), ", ")
	require.Equal(t, "hello, world", str)

	str = Join(asIterator("hello", "world", "as usual"), ", ")
	require.Equal(t, "hello, world, as usual", str)
}
