package dummy

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockClient(t *testing.T) {
	t.Run("no looping", func(t *testing.T) {
		slices := [][]byte{
			[]byte("Hello"), []byte("world!"),
		}
		client := NewMockClient(slices...)

		for _, slice := range slices {
			got, err := client.Read()
			require.NoError(t, err)
			require.Equal(t, string(slice), string(got))
		}

		_, err := client.Read()
		require.EqualError(t, err, io.EOF.Error())
	})

	t.Run("looped slices", func(t *testing.T) {
		slices := [][]byte{
			[]byte("Hello"), []byte("world"), []byte("!"),
		}
		client := NewMockClient(slices...).LoopReads()
		for i := 0; i < len(slices)*2; i++ {
			data, err := client.Read()
			require.NoError(t, err)
			require.Equal(t, string(slices[i%len(slices)]), string(data))
		}
	})
}
