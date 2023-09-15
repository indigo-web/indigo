package queryparser

import (
	"github.com/indigo-web/indigo/http/headers"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	hdrs := headers.NewHeaders()

	singlePair := []byte("something=somewhere")
	manyPairs := []byte("something=somewhere&lorem=ipsum&good=bad&bad=good&paradox=life&dog=cat&cat=dog&life=bad")

	b.Run("SinglePair", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Parse(singlePair, hdrs)
			hdrs.Clear()
		}
	})

	b.Run("ManyPairs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Parse(manyPairs, hdrs)
			hdrs.Clear()
		}
	})
}
