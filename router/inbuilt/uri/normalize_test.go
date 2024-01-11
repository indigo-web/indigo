package uri

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNormalize(t *testing.T) {
	t.Run("single slash", func(t *testing.T) {
		norm := Normalize("/")
		require.Equal(t, "/", norm)
	})

	t.Run("empty", func(t *testing.T) {
		norm := Normalize("")
		require.Equal(t, "", norm)
	})

	t.Run("single trailing", func(t *testing.T) {
		norm := Normalize("/api/")
		require.Equal(t, "/api", norm)
	})

	t.Run("multiple trailing", func(t *testing.T) {
		norm := Normalize("/api/////")
		require.Equal(t, "/api", norm)
	})
}
