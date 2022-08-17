package queryparser

import (
	"indigo/errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParse_Positive(t *testing.T) {
	t.Run("OnePair", func(t *testing.T) {
		query := "hello=world"
		parsed, err := Parse([]byte(query))
		require.NoError(t, err)
		require.Contains(t, parsed, "hello")
		require.Equal(t, "world", string(parsed["hello"]))
	})

	t.Run("TwoPairs", func(t *testing.T) {
		query := "hello=world&lorem=ipsum"
		parsed, err := Parse([]byte(query))
		require.NoError(t, err)
		require.Contains(t, parsed, "hello")
		require.Equal(t, "world", string(parsed["hello"]))
		require.Contains(t, parsed, "lorem")
		require.Equal(t, "ipsum", string(parsed["lorem"]))
	})
}

func TestParse_Negative(t *testing.T) {
	t.Run("EmptyValueBeforeAmpersand", func(t *testing.T) {
		query := "hello=&another=pair"
		parsed, err := Parse([]byte(query))
		require.NoError(t, err)
		require.Contains(t, parsed, "hello")
		require.Empty(t, parsed["hello"])
	})

	t.Run("EmptyValueInTheEnd", func(t *testing.T) {
		query := "hello="
		parsed, err := Parse([]byte(query))
		require.NoError(t, err)
		require.Contains(t, parsed, "hello")
		require.Empty(t, parsed["hello"])
	})

	t.Run("EmptyName", func(t *testing.T) {
		query := "=world"
		_, err := Parse([]byte(query))
		require.ErrorIs(t, err, errors.ErrBadQuery)
	})

	t.Run("AmpersandInTheEnd", func(t *testing.T) {
		query := "hello=world&"
		_, err := Parse([]byte(query))
		require.ErrorIs(t, err, errors.ErrBadQuery)
	})

	t.Run("OnlyKeyInTheEnd", func(t *testing.T) {
		query := "hello=world&lorem"
		_, err := Parse([]byte(query))
		require.ErrorIs(t, err, errors.ErrBadQuery)
	})
}
