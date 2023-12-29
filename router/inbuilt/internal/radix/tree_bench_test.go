package radix

import (
	"github.com/indigo-web/indigo/internal/datastruct"
	"github.com/indigo-web/indigo/router/inbuilt/internal/types"
	"testing"
)

func BenchmarkTreeMatch(b *testing.B) {
	tree := NewTree()

	payload := Payload{
		MethodsMap: types.MethodsMap{},
		Allow:      "",
	}
	tree.MustInsert(MustParse(staticSample), payload)
	tree.MustInsert(MustParse(shortTemplateSample), payload)
	tree.MustInsert(MustParse(mediumTemplateSample), payload)
	tree.MustInsert(MustParse(longTemplateSample), payload)
	const paramsPreAlloc = 5
	params := datastruct.NewKeyValuePreAlloc(paramsPreAlloc)

	b.Run("simple static", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(staticSample, params)
		}
	})

	b.Run("short dynamic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(shortSample, params)
		}
	})

	b.Run("medium dynamic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(mediumSample, params)
		}
	})

	b.Run("long dynamic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(longSample, params)
		}
	})
}
