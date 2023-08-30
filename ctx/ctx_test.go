package ctx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWithValue(t *testing.T) {
	t.Run("put and get value", func(t *testing.T) {
		ctx := WithValue(context.Background(), "hello", "world")
		value := ctx.Value("hello")
		require.NotNil(t, value)
		require.Equal(t, "world", value.(string))
	})

	t.Run("multiple layers", func(t *testing.T) {
		ctx := WithValue(context.Background(), "hello", "world")
		ctx = WithValue(ctx, "key", "value")
		value := ctx.Value("hello")
		require.NotNil(t, value)
		require.Equal(t, "world", value.(string))
	})

	t.Run("int key", func(t *testing.T) {
		ctx := WithValue(context.Background(), 4, 2)
		value := ctx.Value(4)
		require.NotNil(t, value)
		require.Equal(t, 2, value.(int))
	})
}
