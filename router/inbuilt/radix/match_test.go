package radix

import (
	"context"
	"testing"

	routertypes "github.com/indigo-web/indigo/v2/router/inbuilt/types"
	"github.com/stretchr/testify/require"
)

func TestNode_Match_Positive(t *testing.T) {
	tree := NewTree()

	payload := Payload{
		MethodsMap: routertypes.MethodsMap{},
		Allow:      "",
	}
	tree.MustInsert(MustParse(staticSample), payload)
	tree.MustInsert(MustParse(unnamedTemplateSample), payload)
	tree.MustInsert(MustParse(shortTemplateSample), payload)
	tree.MustInsert(MustParse(mediumTemplateSample), payload)
	tree.MustInsert(MustParse(longTemplateSample), payload)
	tree.MustInsert(MustParse("/"), payload)

	t.Run("StaticMatch", func(t *testing.T) {
		_, handler := tree.Match(context.Background(), staticSample)
		require.NotNil(t, handler)
	})

	t.Run("UnnamedMatch", func(t *testing.T) {
		ctx, handler := tree.Match(context.Background(), unnamedSample)
		require.Nil(t, ctx.Value(""))
		require.NotNil(t, handler)
	})

	t.Run("ShortTemplateMatch", func(t *testing.T) {
		ctx, handler := tree.Match(context.Background(), shortSample)
		require.NotNil(t, handler)
		require.Equal(t, "some-very-long-world", ctx.Value("world"))
	})

	t.Run("MediumTemplateMatch", func(t *testing.T) {
		ctx, handler := tree.Match(context.Background(), mediumSample)
		require.NotNil(t, handler)
		require.Equal(t, "world-finally-became", ctx.Value("world"))
		require.Equal(t, "good-as-fuck", ctx.Value("good"))
		require.Equal(t, "ok-let-it-be", ctx.Value("ok"))
	})

	t.Run("Root", func(t *testing.T) {
		_, handler := tree.Match(context.Background(), "/")
		require.NotNil(t, handler)
	})
}

func TestNode_Match_Negative(t *testing.T) {
	tree := NewTree()

	payload := Payload{
		MethodsMap: routertypes.MethodsMap{},
		Allow:      "",
	}
	tree.MustInsert(MustParse(staticSample), payload)
	tree.MustInsert(MustParse(shortTemplateSample), payload)
	tree.MustInsert(MustParse(mediumTemplateSample), payload)
	tree.MustInsert(MustParse(longTemplateSample), payload)

	t.Run("EmptyDynamicPath_WithTrailingSlash", func(t *testing.T) {
		_, handler := tree.Match(context.Background(), "/hello/")
		require.Nil(t, handler)
	})

	t.Run("EmptyDynamicPath_NoTrailingSlash", func(t *testing.T) {
		_, handler := tree.Match(context.Background(), "/hello")
		require.Nil(t, handler)
	})

	t.Run("EmptyDynamicPath_BetweenStatic", func(t *testing.T) {
		_, handler := tree.Match(context.Background(), "/hello//very/good/ok")
		require.Nil(t, handler)
	})
}
