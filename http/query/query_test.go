package query

import (
	"github.com/indigo-web/indigo/http/headers"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	// here we test laziness of query

	// just test that passed buffer's content will not be used
	header := headers.NewHeaders(nil)
	query := NewQuery(header)
	query.Set([]byte("hello=world"))
	require.Equal(t, "hello=world", string(query.raw))
	require.False(t, query.parsed)

	t.Run("GetExistingKey", func(t *testing.T) {
		value, err := query.Get("hello")
		require.NoError(t, err)
		require.Equal(t, "world", value)
	})

	t.Run("GetNonExistingKey", func(t *testing.T) {
		_, err := query.Get("lorem")
		require.ErrorIs(t, err, ErrNoSuchKey)
	})
}
