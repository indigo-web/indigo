package dummy

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCircularClient(t *testing.T) {
	t.Run("single slice", func(t *testing.T) {
		client := NewClient([]byte("Hello, world!"))
		data, err := client.Read()
		require.NoError(t, err)
		require.Equal(t, "Hello, world!", string(data))
		data, err = client.Read()
		require.NoError(t, err)
		require.Equal(t, "Hello, world!", string(data))
	})

	t.Run("multiple slices", func(t *testing.T) {
		slices := [][]byte{
			[]byte("Hello"), []byte("world"), []byte("!"),
		}
		client := NewClient(slices...)
		for i := 0; i < len(slices)*2; i++ {
			data, err := client.Read()
			require.NoError(t, err)
			require.Equal(t, string(slices[i%len(slices)]), string(data))
		}
	})
}
