package formdata

import (
	"fmt"
	"github.com/indigo-web/indigo/http/form"
	"github.com/indigo-web/indigo/http/status"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	singlePair := []byte(generatePairs(1))
	manyPairs := []byte(generatePairs(20))
	veryManyPairs := []byte(generatePairs(100))

	b.Run("single pair", benchmark(singlePair))
	b.Run("20 pairs", benchmark(manyPairs))
	b.Run("100 pairs", benchmark(veryManyPairs))
}

func benchmark(data []byte) func(b *testing.B) {
	buff := make([]byte, 0, len(data))
	dst := make(form.Form, 0, 1024)

	return func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = ParseURLEncoded(dst, data, buff)
		}
	}
}

func generatePairs(n int) string {
	var result string
	const (
		key   = "something"
		value = "somewhere"
	)

	for i := 0; i < n; i++ {
		result += fmt.Sprintf("%s%d=%s&", key, i, value)
	}

	return strings.TrimSuffix(result, "&")
}

func TestParseURLEncoded(t *testing.T) {
	t.Run("single pair", func(t *testing.T) {
		sample := "hello=world"
		result, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.NoError(t, err)
		require.Equal(t, form.Form{
			{Name: "hello", Value: "world"},
		}, result)
	})

	t.Run("two pairs", func(t *testing.T) {
		sample := "hello=world&lorem=ipsum"
		result, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.NoError(t, err)
		require.Equal(t, form.Form{
			{Name: "hello", Value: "world"},
			{Name: "lorem", Value: "ipsum"},
		}, result)
	})

	t.Run("empty value before ampersand", func(t *testing.T) {
		sample := "hello=&another=pair"
		result, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.NoError(t, err)
		require.Equal(t, form.Form{
			{Name: "hello"},
			{Name: "another", Value: "pair"},
		}, result)
	})

	t.Run("single entry without value", func(t *testing.T) {
		sample := "hello="
		result, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.NoError(t, err)
		require.Equal(t, form.Form{
			{Name: "hello"},
		}, result)
	})

	t.Run("empty key", func(t *testing.T) {
		sample := "=world"
		_, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.EqualError(t, err, status.ErrBadEncoding.Error())
	})

	t.Run("ampersand without continuation at the end", func(t *testing.T) {
		sample := "hello=world&"
		result, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.NoError(t, err)
		require.Equal(t, form.Form{
			{Name: "hello", Value: "world"},
		}, result)
	})

	t.Run("flag", func(t *testing.T) {
		{
			result, _, err := ParseURLEncoded([]form.Data{}, []byte("lorem&hello=world&foo=bar"), []byte{})
			require.NoError(t, err)
			require.Equal(t, form.Form{
				{Name: "lorem"},
				{Name: "hello", Value: "world"},
				{Name: "foo", Value: "bar"},
			}, result)
		}

		{
			result, _, err := ParseURLEncoded([]form.Data{}, []byte("hello=world&lorem&foo=bar"), []byte{})
			require.NoError(t, err)
			require.Equal(t, form.Form{
				{Name: "hello", Value: "world"},
				{Name: "lorem"},
				{Name: "foo", Value: "bar"},
			}, result)
		}

		{
			result, _, err := ParseURLEncoded([]form.Data{}, []byte("hello=world&foo=bar&lorem"), []byte{})
			require.NoError(t, err)
			require.Equal(t, form.Form{
				{Name: "hello", Value: "world"},
				{Name: "foo", Value: "bar"},
				{Name: "lorem"},
			}, result)
		}
	})

	t.Run("single flag", func(t *testing.T) {
		sample := "lorem"
		result, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.NoError(t, err)
		require.Equal(t, form.Form{{Name: "lorem"}}, result)
	})

	t.Run("encoded spaces", func(t *testing.T) {
		sample := "hel+lo=wo+rld"
		result, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.NoError(t, err)
		require.Equal(t, form.Form{
			{Name: "hel lo", Value: "wo rld"},
		}, result)
	})

	t.Run("url encoded", func(t *testing.T) {
		sample := "hel%20lo=wo%20rld%21"
		result, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.NoError(t, err)
		require.Equal(t, form.Form{
			{Name: "hel lo", Value: "wo rld!"},
		}, result)
	})

	t.Run("encoded plus char", func(t *testing.T) {
		sample := "hel%2blo=wo%2brld"
		result, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.NoError(t, err)
		require.Equal(t, form.Form{
			{Name: "hel+lo", Value: "wo+rld"},
		}, result)
	})

	t.Run("edgecases", func(t *testing.T) {
		for _, tc := range []string{
			"%07=hello",
			"he%5=ok",
			"hello=%5",
		} {
			_, _, err := ParseURLEncoded([]form.Data{}, []byte(tc), []byte{})
			require.EqualError(t, err, status.ErrBadEncoding.Error(), tc)
		}
	})

	t.Run("nonprintable in value", func(t *testing.T) {
		sample := "hello=%07"
		result, _, err := ParseURLEncoded([]form.Data{}, []byte(sample), []byte{})
		require.NoError(t, err)
		require.Equal(t, form.Form{
			{Name: "hello", Value: "\x07"},
		}, result)
	})
}
