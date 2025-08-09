package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	t.Run("pure domain", func(t *testing.T) {
		require.Equal(t, "foo.example.com", Normalize("foo.example.com"))
	})

	t.Run("with default port", func(t *testing.T) {
		require.Equal(t, "foo.example.com", Normalize("foo.example.com:80"))
		require.Equal(t, "foo.example.com", Normalize("foo.example.com:443"))
	})

	t.Run("with different port", func(t *testing.T) {
		require.Equal(t, "foo.example.com:8080", Normalize("foo.example.com:8080"))
	})

	t.Run("with www prefix", func(t *testing.T) {
		require.Equal(t, "foo.example.com", Normalize("www.foo.example.com"))
	})

	t.Run("ip address", func(t *testing.T) {
		require.Equal(t, "1.1.1.1", Normalize("1.1.1.1:80"))
	})
}

func TestTrimPort(t *testing.T) {
	t.Run("with port", func(t *testing.T) {
		require.Equal(t, "localhost", TrimPort("localhost:8080"))
	})

	t.Run("without port", func(t *testing.T) {
		require.Equal(t, "localhost", TrimPort("localhost"))
	})
}
