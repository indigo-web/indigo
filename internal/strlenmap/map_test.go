package strlenmap

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMap(t *testing.T) {
	t.Run("brand-new instance", func(t *testing.T) {
		m := New[string]()
		_, found := m.Get("")
		require.False(t, found)
		_, found = m.Get("any key")
		require.False(t, found)
	})

	t.Run("distributed by buckets", func(t *testing.T) {
		m := New[string]()
		m.Insert("hello", "world")
		m.Insert("hi", "sky")

		testKey(t, m, "", "", false)
		testKey(t, m, "hallo", "", false)
		testKey(t, m, "hello", "world", true)
		testKey(t, m, "hi", "sky", true)
	})

	t.Run("multiple entries in a single bucket", func(t *testing.T) {
		m := New[string]()
		m.Insert("hello", "world")
		m.Insert("hallo", "sky")
		m.Insert("borat", "best film fr")

		testKey(t, m, "hello", "world", true)
		testKey(t, m, "hallo", "sky", true)
		testKey(t, m, "borat", "best film fr", true)
	})
}

func testKey(t *testing.T, m *Map[string], key string, wantedValue string, wantedFound bool) {
	value, found := m.Get(key)
	require.Equal(t, wantedFound, found)
	require.Equal(t, wantedValue, value)
}
