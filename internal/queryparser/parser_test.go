package queryparser

import (
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/status"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse_Positive(t *testing.T) {
	t.Run("OnePair", func(t *testing.T) {
		query := "hello=world"
		result := headers.NewHeaders(nil)
		err := Parse([]byte(query), result)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Equal(t, "world", result.Value("hello"))
	})

	t.Run("TwoPairs", func(t *testing.T) {
		query := "hello=world&lorem=ipsum"
		result := headers.NewHeaders(nil)
		err := Parse([]byte(query), result)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Equal(t, "world", result.Value("hello"))
		require.True(t, result.Has("lorem"))
		require.Equal(t, "ipsum", result.Value("lorem"))
	})
}

func TestParse_Negative(t *testing.T) {
	t.Run("EmptyValueBeforeAmpersand", func(t *testing.T) {
		query := "hello=&another=pair"
		result := headers.NewHeaders(nil)
		err := Parse([]byte(query), result)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Empty(t, result.Value("hello"))
	})

	t.Run("EmptyValueInTheEnd", func(t *testing.T) {
		query := "hello="
		result := headers.NewHeaders(nil)
		err := Parse([]byte(query), result)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Empty(t, result.Value("hello"))
	})

	t.Run("EmptyName", func(t *testing.T) {
		query := "=world"
		err := Parse([]byte(query), headers.NewHeaders(nil))
		require.ErrorIs(t, err, status.ErrBadQuery)
	})

	t.Run("AmpersandInTheEnd", func(t *testing.T) {
		query := "hello=world&"
		err := Parse([]byte(query), headers.NewHeaders(nil))
		require.ErrorIs(t, err, status.ErrBadQuery)
	})

	t.Run("OnlyKeyInTheEnd", func(t *testing.T) {
		query := "hello=world&lorem"
		err := Parse([]byte(query), headers.NewHeaders(nil))
		require.ErrorIs(t, err, status.ErrBadQuery)
	})
}
