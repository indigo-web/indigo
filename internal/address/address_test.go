package address

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("valid ip and port", func(t *testing.T) {
		addr, err := Parse("localhost:8080")
		require.NoError(t, err)
		require.Equal(t, "localhost", addr.Host)
		require.Equal(t, 8080, int(addr.Port))
	})

	t.Run("no ip but port", func(t *testing.T) {
		addr, err := Parse(":8080")
		require.NoError(t, err)
		require.Equal(t, DefaultHost, addr.Host)
		require.Equal(t, 8080, int(addr.Port))
	})

	t.Run("only ip", func(t *testing.T) {
		_, err := Parse("localhost")
		require.NotNil(t, err, "error expected, got nil instead")
		require.Equal(t, "no port given", err.Error())
	})

	t.Run("too big port", func(t *testing.T) {
		_, err := Parse(":65536")
		require.NotNil(t, err, "error expected, got nil instead")
		require.Equal(t, "invalid port: 65536", err.Error())
	})
}
