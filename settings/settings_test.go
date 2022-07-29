package settings

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func compareDefaultValues(t *testing.T, settings Settings) {
	require.Equal(t, defaultMaxBodyLength, settings.MaxBodyLength)
	require.Equal(t, defaultBodyBuffSize, settings.DefaultBodyBuffSize)
	require.Equal(t, defaultSockReadBuffSize, settings.SockReadBuffSize)
	require.Equal(t, defaultMaxHeaders, settings.MaxHeaders)
	require.Equal(t, defaultMaxURILength, settings.MaxURILength)
	require.Equal(t, defaultMaxHeaderKeyLength, settings.MaxHeaderKeyLength)
	require.Equal(t, defaultMaxHeaderValueLength, settings.MaxHeaderValueLength)
	require.Equal(t, defaultMaxBodyChunkLength, settings.MaxBodyChunkLength)
	require.Equal(t, defaultInfoLineBuffSize, settings.DefaultInfoLineBuffSize)
	require.Equal(t, defaultHeadersBuffSize, settings.DefaultHeadersBuffSize)
}

func TestPrepare(t *testing.T) {
	t.Run("AllNullValues", func(t *testing.T) {
		compareDefaultValues(t, Prepare(Settings{}))
	})
	t.Run("SomeNonNull", func(t *testing.T) {
		settings := Prepare(Settings{
			MaxURILength: 4096,
		})
		require.Equal(t, uint16(4096), settings.MaxURILength)
	})
}

func TestDefault(t *testing.T) {
	compareDefaultValues(t, Default())
}
