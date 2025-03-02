package query

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"testing"

	"github.com/stretchr/testify/require"
)

func getQ(query string) Query {
	q := New(keyvalue.New(), config.Default())
	q.Update([]byte(query))
	return q
}

func TestQuery(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		q := getQ("")
		p, err := q.Cook()
		require.NoError(t, err)
		require.True(t, p.Empty())
	})

	t.Run("single pair", func(t *testing.T) {
		q := getQ("hello=world")
		p, err := q.Cook()
		require.NoError(t, err)
		require.Equal(t, "world", p.Value("hello"))
		require.Equal(t, 1, p.Len())
	})

	t.Run("last pair", func(t *testing.T) {
		q := getQ("a=b&hello=world")
		p, err := q.Cook()
		require.NoError(t, err)
		require.Equal(t, "world", p.Value("hello"))
		require.Equal(t, 2, p.Len())
	})

	t.Run("urlencoded", func(t *testing.T) {
		const query = "he%6c%6Co=wo%72ld"
		q := getQ(query)
		p, err := q.Cook()
		require.NoError(t, err)
		require.Equal(t, "world", p.Value("hello"))
		require.Equal(t, 1, p.Len())
		// actually, this check isn't mandatory. If query will try to modify
		// the query anyhow it'll segfault
		require.Equal(t, query, q.String())
	})

	t.Run("flag", func(t *testing.T) {
		q := getQ("a")
		p, err := q.Cook()
		require.NoError(t, err)
		require.Equal(t, "1", p.Value("a"))
		require.Equal(t, 1, p.Len())
	})

	t.Run("empty value", func(t *testing.T) {
		q := getQ("a=")
		p, err := q.Cook()
		require.NoError(t, err)
		require.Empty(t, p.Value("a"))
		require.Equal(t, 1, p.Len())
	})

	t.Run("empty flag", func(t *testing.T) {
		q := getQ("a&&b")
		_, err := q.Cook()
		require.EqualError(t, err, status.ErrBadRequest.Error())
	})

	t.Run("nonprintable in key", func(t *testing.T) {
		q := getQ("h\x00ello=world")
		_, err := q.Cook()
		require.EqualError(t, err, status.ErrBadRequest.Error())
	})

	t.Run("nonprintable in flag key", func(t *testing.T) {
		q := getQ("h\x00ello")
		_, err := q.Cook()
		require.EqualError(t, err, status.ErrBadRequest.Error())
	})

	t.Run("nonprintable in value", func(t *testing.T) {
		q := getQ("hello=w\x00rld&a=b")
		_, err := q.Cook()
		require.EqualError(t, err, status.ErrBadRequest.Error())
	})

	t.Run("nonprintable in last value", func(t *testing.T) {
		q := getQ("a=b&hello=w\x00rld")
		_, err := q.Cook()
		require.EqualError(t, err, status.ErrBadRequest.Error())
	})
}
