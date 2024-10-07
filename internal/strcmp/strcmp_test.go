package strcmp

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFold(t *testing.T) {
	require.True(t, Fold("HELLO", "hello"))
	require.True(t, Fold("\r\n\r\n", "\r\n\r\n"))
	require.False(t, Fold("\v\t", "\r\t"))
}
