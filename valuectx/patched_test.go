package valuectx

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWithValue(t *testing.T) {
	ctx := WithValue(context.Background(), "hello", "world")
	require.Equal(t, "world", ctx.Value("hello"))
}
