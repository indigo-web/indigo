package strutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalize(t *testing.T) {
	require.Equal(t, "pavlo:80", NormalizeAddress("pavlo:80"))
	require.Equal(t, "0.0.0.0:80", NormalizeAddress(":80"))
}
