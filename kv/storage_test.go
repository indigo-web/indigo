package kv

import (
	"github.com/stretchr/testify/require"
	"slices"
	"testing"
)

func TestStorage(t *testing.T) {
	getHeaders := func() *Storage {
		return New().
			Add("Foo", "bar").
			Add("Hello", "World").
			Add("Lorem", "ipsum").
			Add("hello", "Pavlo")
	}

	t.Run("delete", func(t *testing.T) {
		kv := getHeaders().Delete("HELLO")

		want := []Pair{
			{"Foo", "bar"},
			{"Lorem", "ipsum"},
		}

		require.Equal(t, len(want), kv.Len())
		for _, p := range want {
			require.Equal(t, []string{p.Value}, slices.Collect(kv.Values(p.Key)))
		}

		indexOf := func(key string) int {
			for i, p := range want {
				if p.Key == key {
					return i
				}
			}

			return -1
		}

		for key, value := range kv.Pairs() {
			idx := indexOf(key)
			require.NotEqual(t, -1, idx)
			require.Equal(t, want[idx].Value, value)
		}
	})

	t.Run("set", func(t *testing.T) {
		kv := getHeaders().Set("HELLO", "no more Pavlo")

		want := []Pair{
			{"Foo", "bar"},
			{"HELLO", "no more Pavlo"},
			{"Lorem", "ipsum"},
		}

		require.Equal(t, len(want), kv.Len())
		for _, p := range want {
			require.Equal(t, []string{p.Value}, slices.Collect(kv.Values(p.Key)))
		}
	})

	t.Run("set new key", func(t *testing.T) {
		kv := New().
			Add("Pavlo", "the best").
			Set("Glory to", "Ukraine")

		want := []Pair{
			{"Pavlo", "the best"},
			{"Glory to", "Ukraine"},
		}

		require.Equal(t, len(want), kv.Len())
		for _, p := range want {
			require.Equal(t, []string{p.Value}, slices.Collect(kv.Values(p.Key)))
		}
	})

	t.Run("keys", func(t *testing.T) {
		kv := getHeaders().Delete("hello")
		require.Equal(t, []string{"Foo", "Lorem"}, slices.Collect(kv.Keys()))
	})

	t.Run("empty", func(t *testing.T) {
		kv := getHeaders()
		for key := range kv.Keys() {
			kv.Delete(key)
		}

		require.True(t, kv.Empty())
	})
}
