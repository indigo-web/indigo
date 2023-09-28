package coding

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGZIPCoding(t *testing.T) {
	t.Run("compress and decompress", func(t *testing.T) {
		source := "Hello, world!"
		coder := NewGZIP(make([]byte, 1024))
		result, err := cycle(source, coder)
		require.NoError(t, err)
		require.Equal(t, source, string(result))
	})

	t.Run("multiple reads and writes", func(t *testing.T) {
		source1 := "Hello, world!"
		source2 := "lorem ipsum"
		coder := NewGZIP(make([]byte, 1024))

		result, err := cycle(source1, coder)
		require.NoError(t, err)
		require.Equal(t, source1, string(result))

		result, err = cycle(source2, coder)
		require.NoError(t, err)
		require.Equal(t, source2, string(result))
	})
}

func cycle(source string, coder Coding) ([]byte, error) {
	compressed, err := coder.Encode([]byte(source))
	if err != nil {
		return nil, err
	}

	decompressed, err := coder.Decode(compressed)
	return decompressed, err
}
