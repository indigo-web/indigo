package radix

import (
	"context"
	"github.com/fakefloordiv/indigo/types"
	"testing"
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

func nopHandler(context.Context, *types.Request) types.Response {
	return types.OK()
}

func BenchmarkTreeMatch(b *testing.B) {
	tree := NewTree()

	tree.MustInsert(MustParse(staticSample), nopHandler)
	tree.MustInsert(MustParse(shortTemplateSample), nopHandler)
	tree.MustInsert(MustParse(mediumTemplateSample), nopHandler)
	tree.MustInsert(MustParse(longTemplateSample), nopHandler)

	b.Run("OnlyStatic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(context.Background(), staticSample)
		}
	})

	b.Run("Short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(context.Background(), shortSample)
		}
	})

	b.Run("Medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(context.Background(), mediumSample)
		}
	})

	b.Run("Long", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(context.Background(), longSample)
		}
	})
}
