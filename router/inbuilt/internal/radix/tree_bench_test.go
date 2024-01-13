package radix

import (
	"github.com/indigo-web/indigo/internal/keyvalue"
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
	params := keyvalue.NewPreAlloc(paramsPreAlloc)

	b.Run("simple static", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(staticSample, params)
			params.Clear()
		}
	})

	b.Run("short dynamic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(shortSample, params)
			params.Clear()
		}
	})

	b.Run("medium dynamic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(mediumSample, params)
			params.Clear()
		}
	})

	b.Run("long dynamic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tree.Match(longSample, params)
			params.Clear()
		}
	})
}
