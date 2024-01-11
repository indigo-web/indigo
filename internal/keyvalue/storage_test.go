package keyvalue

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestKeyValueStorage(t *testing.T) {
	testValues := func(t *testing.T, kv *Storage) {
		for _, tc := range []struct {
			Key    string
			Values []string
		}{
			{
				Key:    "Hello",
				Values: []string{"world"},
			},
			{
				Key:    "Some",
				Values: []string{"multiple", "values"},
			},
			{
				Key:    "sOME",
				Values: []string{"multiple", "values"},
			},
		} {
			value, found := kv.Get(tc.Key)
			require.True(t, found)
			require.Equal(t, tc.Values[0], value)

			values := kv.Values(tc.Key)
			require.Equal(t, tc.Values, values)
		}
	}

	t.Run("Value with manual filling", func(t *testing.T) {
		kv := New()
		kv.Add("Hello", "world")
		kv.Add("Some", "multiple")
		kv.Add("Some", "values")
		testValues(t, kv)
	})

	t.Run("Value with map instantiation", func(t *testing.T) {
		kv := NewFromMap(map[string][]string{
			"Hello": {"world"},
			"Some":  {"multiple", "values"},
		})
		testValues(t, kv)
	})

	t.Run("Has", func(t *testing.T) {
		kv := New()
		kv.Add("Hello", "world")
		require.True(t, kv.Has("Hello"))
		require.True(t, kv.Has("hELLO"))
		require.False(t, kv.Has("random"))
	})

	t.Run("Keys", func(t *testing.T) {
		kv := New()
		kv.Add("Hello", "world")
		kv.Add("sOME", "multiple")
		kv.Add("Some", "values")
		kv.Add("hELLO", "nether")
		require.Equal(t, []string{"Hello", "sOME"}, kv.Keys())
	})
}
