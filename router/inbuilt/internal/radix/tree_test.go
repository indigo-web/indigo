package radix

import (
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	staticSample = "/hello/world/length/does/not/matter"

	unnamedTemplateSample = "/api/{}"
	unnamedSample         = "/api/v1"

	shortTemplateSample = "/hello/{world}"
	shortSample         = "/hello/some-very-long-world"

	mediumTemplateSample = "/hello/{world}/very/{good}/{ok}"
	mediumSample         = "/hello/world-finally-became/very/good-as-fuck/ok-let-it-be"

	longTemplateSample = "/hello/{world}/very/{good}/{ok}/{wanna}/somestatic/{finally}/good"
	longSample         = "/hello/world-finally-became/very/good-as-fuck/ok-let-it-be/i-wanna-/somestatic/finally-matcher-is-here/good"
)

func TestNode_Match_Positive(t *testing.T) {
	tree := NewTree()
	payload := Payload{
		MethodsMap: types.MethodsMap{},
	}
	tree.MustInsert(MustParse(staticSample), payload)
	tree.MustInsert(MustParse(unnamedTemplateSample), payload)
	tree.MustInsert(MustParse(shortTemplateSample), payload)
	tree.MustInsert(MustParse(mediumTemplateSample), payload)
	tree.MustInsert(MustParse(longTemplateSample), payload)
	tree.MustInsert(MustParse("/"), payload)

	t.Run("match static", func(t *testing.T) {
		params := keyvalue.New()
		handler := tree.Match(staticSample, params)
		require.NotNil(t, handler)
	})

	t.Run("unnamed match", func(t *testing.T) {
		params := keyvalue.New()
		handler := tree.Match(unnamedSample, params)
		require.Empty(t, params.Values(""))
		require.NotNil(t, handler)
	})

	t.Run("short template", func(t *testing.T) {
		params := keyvalue.New()
		handler := tree.Match(shortSample, params)
		require.NotNil(t, handler)
		require.Equal(t, "some-very-long-world", params.Value("world"))
	})

	t.Run("medium template", func(t *testing.T) {
		params := keyvalue.New()
		handler := tree.Match(mediumSample, params)
		require.NotNil(t, handler)
		require.Equal(t, "world-finally-became", params.Value("world"))
		require.Equal(t, "good-as-fuck", params.Value("good"))
		require.Equal(t, "ok-let-it-be", params.Value("ok"))
	})

	t.Run("root", func(t *testing.T) {
		params := keyvalue.New()
		handler := tree.Match("/", params)
		require.NotNil(t, handler)
	})
}

func TestNode_Match_Negative(t *testing.T) {
	tree := NewTree()
	payload := Payload{
		MethodsMap: types.MethodsMap{},
	}
	tree.MustInsert(MustParse(staticSample), payload)
	tree.MustInsert(MustParse(shortTemplateSample), payload)
	tree.MustInsert(MustParse(mediumTemplateSample), payload)
	tree.MustInsert(MustParse(longTemplateSample), payload)

	t.Run("static prefix", func(t *testing.T) {
		params := keyvalue.New()
		handler := tree.Match("/hello", params)
		require.Nil(t, handler)
		handler = tree.Match("/hello/", params)
		require.Nil(t, handler)
	})

	t.Run("empty dynamic section", func(t *testing.T) {
		params := keyvalue.New()
		handler := tree.Match("/hello//very/good/ok", params)
		require.Nil(t, handler)
	})
}
