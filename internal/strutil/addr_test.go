package strutil

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNormalize(t *testing.T) {
	require.Equal(t, "pavlo:80", NormalizeAddress("pavlo:80"))
	require.Equal(t, "0.0.0.0:80", NormalizeAddress(":80"))
}
