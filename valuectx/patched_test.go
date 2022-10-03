package valuectx

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWithValue(t *testing.T) {
	t.Run("PutAndGetValue", func(t *testing.T) {
		ctx := WithValue(context.Background(), "hello", "world")
		value := ctx.Value("hello")
		require.NotNil(t, value)
		require.Equal(t, "world", ctx.Value("hello").(string))
	})

	t.Run("ValueInAnotherLayer", func(t *testing.T) {
		ctx := WithValue(context.Background(), "hello", "world")
		ctx = WithValue(ctx, "key", "value")
		value := ctx.Value("hello")
		require.NotNil(t, value)
		require.Equal(t, "world", value.(string))
	})
}
