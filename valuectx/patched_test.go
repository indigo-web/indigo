package valuectx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithValue(t *testing.T) {
	ctx := WithValue(context.Background(), "hello", "world")
	require.Equal(t, "world", ctx.Value("hello"))
}
