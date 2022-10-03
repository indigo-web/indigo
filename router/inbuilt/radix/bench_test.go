package radix

import (
	"context"
	routertypes "github.com/fakefloordiv/indigo/router/inbuilt/types"
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

func BenchmarkTreeMatch(b *testing.B) {
	tree := NewTree()

	payload := Payload{
		MethodsMap: routertypes.MethodsMap{},
		Allow:      "",
	}
	tree.MustInsert(MustParse(staticSample), payload)
	tree.MustInsert(MustParse(shortTemplateSample), payload)
	tree.MustInsert(MustParse(mediumTemplateSample), payload)
	tree.MustInsert(MustParse(longTemplateSample), payload)

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
