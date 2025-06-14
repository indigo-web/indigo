package strutil

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFold(t *testing.T) {
	require.True(t, CmpFold("HELLO", "hello"))
	require.True(t, CmpFold("\r\n\r\n", "\r\n\r\n"))
	require.False(t, CmpFold("\v\t", "\r\t"))
}
