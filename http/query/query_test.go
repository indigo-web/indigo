package query

import (
	"github.com/indigo-web/indigo/http/headers"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	// test the laziness just in place
	header := headers.New()
	query := NewQuery(header)
	query.Set([]byte("hello=world"))
	require.Equal(t, "hello=world", string(query.raw))
	require.False(t, query.parsed)

	t.Run("get existing key", func(t *testing.T) {
		value, err := query.Get("hello")
		require.NoError(t, err)
		require.Equal(t, "world", value)
	})

	t.Run("get non-existing key", func(t *testing.T) {
		_, err := query.Get("lorem")
		require.ErrorIs(t, err, ErrNoSuchKey)
	})

	require.True(t, query.parsed)
}
