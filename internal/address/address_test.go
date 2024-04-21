package address

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("localhost", func(t *testing.T) {
		require.True(t, IsLocalhost("localhost"))
	})

	t.Run("localhost with port", func(t *testing.T) {
		require.True(t, IsLocalhost("localhost:8080"))
	})

	t.Run("ip", func(t *testing.T) {
		require.True(t, IsIP("1.2.3.4"))
	})

	t.Run("ip with port", func(t *testing.T) {
		require.True(t, IsIP("1.2.3.4:8080"))
	})

	t.Run("invalid ip", func(t *testing.T) {
		require.False(t, IsIP("1.2.3.256"))
	})
}
