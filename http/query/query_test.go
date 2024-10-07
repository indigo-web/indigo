package query

import (
	"github.com/indigo-web/indigo/http/headers"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	// test the laziness just in place
	header := headers.New()
	raw := New(header)
	raw.Update([]byte("hello=world"))
	query, err := raw.Cook()
	require.NoError(t, err)

	t.Run("get existing key", func(t *testing.T) {
		value, found := query.Get("hello")
		require.True(t, found)
		require.Equal(t, "world", value)
	})

	t.Run("get non-existing key", func(t *testing.T) {
		value, found := query.Get("lorem")
		require.False(t, found)
		require.Empty(t, value)
	})
}
