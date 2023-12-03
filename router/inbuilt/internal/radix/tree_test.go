package radix

import (
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNode_Match_Positive(t *testing.T) {
	tree := NewTree()
	params := make(Params)

	payload := Payload{
		MethodsMap: types.MethodsMap{},
	}
	tree.MustInsert(MustParse(staticSample), payload)
	tree.MustInsert(MustParse(unnamedTemplateSample), payload)
	tree.MustInsert(MustParse(shortTemplateSample), payload)
	tree.MustInsert(MustParse(mediumTemplateSample), payload)
	tree.MustInsert(MustParse(longTemplateSample), payload)
	tree.MustInsert(MustParse("/"), payload)

	t.Run("StaticMatch", func(t *testing.T) {
		handler := tree.Match(params, staticSample)
		require.NotNil(t, handler)
	})

	t.Run("UnnamedMatch", func(t *testing.T) {
		handler := tree.Match(params, unnamedSample)
		require.Empty(t, params[""])
		require.NotNil(t, handler)
	})

	t.Run("ShortTemplateMatch", func(t *testing.T) {
		handler := tree.Match(params, shortSample)
		require.NotNil(t, handler)
		require.Equal(t, "some-very-long-world", params["world"])
	})

	t.Run("MediumTemplateMatch", func(t *testing.T) {
		handler := tree.Match(params, mediumSample)
		require.NotNil(t, handler)
		require.Equal(t, "world-finally-became", params["world"])
		require.Equal(t, "good-as-fuck", params["good"])
		require.Equal(t, "ok-let-it-be", params["ok"])
	})

	t.Run("Root", func(t *testing.T) {
		handler := tree.Match(params, "/")
		require.NotNil(t, handler)
	})
}

func TestNode_Match_Negative(t *testing.T) {
	tree := NewTree()
	params := make(Params)

	payload := Payload{
		MethodsMap: types.MethodsMap{},
	}
	tree.MustInsert(MustParse(staticSample), payload)
	tree.MustInsert(MustParse(shortTemplateSample), payload)
	tree.MustInsert(MustParse(mediumTemplateSample), payload)
	tree.MustInsert(MustParse(longTemplateSample), payload)

	t.Run("EmptyDynamicPath_WithTrailingSlash", func(t *testing.T) {
		handler := tree.Match(params, "/hello/")
		require.Nil(t, handler)
	})

	t.Run("EmptyDynamicPath_NoTrailingSlash", func(t *testing.T) {
		handler := tree.Match(params, "/hello")
		require.Nil(t, handler)
	})

	t.Run("EmptyDynamicPath_BetweenStatic", func(t *testing.T) {
		handler := tree.Match(params, "/hello//very/good/ok")
		require.Nil(t, handler)
	})
}
