package http

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetHeaders(t *testing.T) {
	t.Run("SetNewHeaders", func(t *testing.T) {
		headers := make(Headers, 2)
		headers.Set([]byte("Hello"), []byte("World"))
		headers.Set([]byte("Test"), []byte("Richtig"))
		require.Equal(t, 2, len(headers), "wanted exactly 2 headers")
	})
	t.Run("SetAndGetNewHeaders", func(t *testing.T) {
		headers := make(Headers, 2)
		headers.Set([]byte("Hello"), []byte("World"))
		headers.Set([]byte("Test"), []byte("Richtig"))
		require.Contains(t, headers, "Hello")
		require.Contains(t, headers, "Test")

		require.Equal(t, []byte("World"), headers["Hello"])
		require.Equal(t, []byte("Richtig"), headers["Test"])
	})
	t.Run("SetAlreadyExisting", func(t *testing.T) {
		headers := make(Headers, 1)
		headers.Set([]byte("Hello"), []byte("World"))
		require.Contains(t, headers, "Hello")

		headers.Set([]byte("Hello"), []byte("Heaven"))
		require.Equal(t, []byte("Heaven"), headers["Hello"])
	})
}
