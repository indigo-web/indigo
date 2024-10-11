package qparams

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/internal/urlencoded"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParamsParser(t *testing.T) {
	t.Run("single pair", func(t *testing.T) {
		query := "hello=world"
		result := keyvalue.New()
		err := Parse([]byte(query), Into(result), urlencoded.Decode)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Equal(t, "world", result.Value("hello"))
	})

	t.Run("two pairs", func(t *testing.T) {
		query := "hello=world&lorem=ipsum"
		result := keyvalue.New()
		err := Parse([]byte(query), Into(result), urlencoded.Decode)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Equal(t, "world", result.Value("hello"))
		require.True(t, result.Has("lorem"))
		require.Equal(t, "ipsum", result.Value("lorem"))
	})

	t.Run("empty value before ampersand", func(t *testing.T) {
		query := "hello=&another=pair"
		result := keyvalue.New()
		err := Parse([]byte(query), Into(result), urlencoded.Decode)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Empty(t, result.Value("hello"))
	})

	t.Run("single entry without value", func(t *testing.T) {
		query := "hello="
		result := keyvalue.New()
		err := Parse([]byte(query), Into(result), urlencoded.Decode)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Empty(t, result.Value("hello"))
	})

	t.Run("empty key", func(t *testing.T) {
		query := "=world"
		err := Parse([]byte(query), Into(keyvalue.New()), urlencoded.Decode)
		require.ErrorIs(t, err, status.ErrBadRequest)
	})

	t.Run("ampersand without continuation at the end", func(t *testing.T) {
		query := "hello=world&"
		result := keyvalue.New()
		err := Parse([]byte(query), Into(result), urlencoded.Decode)
		require.NoError(t, err)
		require.True(t, result.Has("hello"))
		require.Equal(t, "world", result.Value("hello"))
	})

	t.Run("flag", func(t *testing.T) {
		result := keyvalue.New()

		for _, paramsString := range []string{
			"lorem&hello=world&foo=bar",
			"hello=world&lorem&foo=bar",
			"hello=world&foo=bar&lorem",
		} {
			err := Parse([]byte(paramsString), Into(result), urlencoded.Decode)
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
		result := keyvalue.New()
		err := Parse([]byte(query), Into(result), urlencoded.Decode)
		require.NoError(t, err)
		require.True(t, result.Has("lorem"))
		require.Equal(t, defaultEmptyValueContent, result.Value("lorem"))
	})

	t.Run("encoded spaces", func(t *testing.T) {
		query := "hel+lo=wo+rld"
		result := keyvalue.New()
		err := Parse([]byte(query), Into(result), urlencoded.Decode)
		require.NoError(t, err)
		require.True(t, result.Has("hel lo"))
		require.Equal(t, "wo rld", result.Value("hel lo"))
	})

	t.Run("url encoded", func(t *testing.T) {
		query := "hel%20lo=wo%20rld%21"
		result := keyvalue.New()
		err := Parse([]byte(query), Into(result), urlencoded.Decode)
		require.NoError(t, err)
		require.True(t, result.Has("hel lo"))
		require.Equal(t, "wo rld!", result.Value("hel lo"))
	})

	t.Run("encoded plus char", func(t *testing.T) {
		query := "hel%2blo=wo%2brld"
		result := keyvalue.New()
		err := Parse([]byte(query), Into(result), urlencoded.Decode)
		require.NoError(t, err)
		require.True(t, result.Has("hel+lo"))
		require.Equal(t, "wo+rld", result.Value("hel+lo"))
	})
}
