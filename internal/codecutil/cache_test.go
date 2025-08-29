package codecutil

import (
	"testing"

	"github.com/indigo-web/indigo/http/codec"
	"github.com/stretchr/testify/require"
)

type mockCodec struct {
	Instantiated bool
}

func (m *mockCodec) Token() string {
	return "mock"
}

func (m *mockCodec) New() codec.Instance {
	m.Instantiated = true
	return nil
}

func TestLazyInstantiation(t *testing.T) {
	mock := new(mockCodec)
	cache := NewCache([]codec.Codec{mock}, "")
	require.False(t, mock.Instantiated)
	_ = cache.Get("rand")
	require.False(t, mock.Instantiated)
	_ = cache.Get("mock")
	require.True(t, mock.Instantiated)
}

func TestAcceptEncoding(t *testing.T) {
	t.Run("0", func(t *testing.T) {
		require.Equal(t, "identity", AcceptEncoding([]codec.Codec{}))
	})

	t.Run("1", func(t *testing.T) {
		require.Equal(t, "gzip", AcceptEncoding([]codec.Codec{codec.NewGZIP()}))
	})

	t.Run("2", func(t *testing.T) {
		require.Equal(t, "gzip, zstd", AcceptEncoding([]codec.Codec{codec.NewGZIP(), codec.NewZSTD()}))
	})
}
