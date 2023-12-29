package query

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/datastruct"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamsParser(t *testing.T) {
	t.Run("single pair", func(t *testing.T) {
		query := "hello=world"
		result := datastruct.NewKeyValue()
		err := Parse([]byte(query), result)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Equal(t, "world", result.Value("hello"))
	})

	t.Run("two pairs", func(t *testing.T) {
		query := "hello=world&lorem=ipsum"
		result := datastruct.NewKeyValue()
		err := Parse([]byte(query), result)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Equal(t, "world", result.Value("hello"))
		require.True(t, result.Has("lorem"))
		require.Equal(t, "ipsum", result.Value("lorem"))
	})

	t.Run("empty value before ampersand", func(t *testing.T) {
		query := "hello=&another=pair"
		result := datastruct.NewKeyValue()
		err := Parse([]byte(query), result)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Empty(t, result.Value("hello"))
	})

	t.Run("single entry without value", func(t *testing.T) {
		query := "hello="
		result := datastruct.NewKeyValue()
		err := Parse([]byte(query), result)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Empty(t, result.Value("hello"))
	})

	t.Run("empty key", func(t *testing.T) {
		query := "=world"
		err := Parse([]byte(query), datastruct.NewKeyValue())
		require.ErrorIs(t, err, status.ErrBadQuery)
	})

	t.Run("ampersand without continuation at the end", func(t *testing.T) {
		query := "hello=world&"
		result := datastruct.NewKeyValue()
		err := Parse([]byte(query), result)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Equal(t, "world", result.Value("hello"))
	})

	t.Run("flag", func(t *testing.T) {
		result := datastruct.NewKeyValue()

		for _, paramsString := range []string{
			"lorem&hello=world&foo=bar",
			"hello=world&lorem&foo=bar",
			"hello=world&foo=bar&lorem",
		} {
			err := Parse([]byte(paramsString), result)
			require.NoError(t, err, paramsString)
			require.True(t, result.Has("hello"), paramsString)
			require.Equal(t, "world", result.Value("hello"), paramsString)
			require.True(t, result.Has("foo"), paramsString)
			require.Equal(t, "bar", result.Value("foo"), paramsString)
			require.True(t, result.Has("lorem"), paramsString)
			require.Equal(t, defaultEmptyValueContent, result.Value("lorem"), paramsString)
		}
	})

	t.Run("single flag", func(t *testing.T) {
		query := "lorem"
		result := datastruct.NewKeyValue()
		err := Parse([]byte(query), result)
		require.NoError(t, err)
		require.True(t, result.Has("lorem"))
		require.Equal(t, defaultEmptyValueContent, result.Value("lorem"))
	})
}
