package inbuilt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafePath(t *testing.T) {
	for _, tc := range []string{
		"/",
		"/./",
		"/./.",
		"././.",
	} {
		require.True(t, isSafe(tc))
	}

	for _, tc := range []string{
		"/..",
		"../",
		"/../",
	} {
		require.False(t, isSafe(tc))
	}
}
