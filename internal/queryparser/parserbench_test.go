package queryparser

import "testing"

const initialQueryMapSize = 10

var queriesMap = make(map[string][]byte, initialQueryMapSize)

func BenchmarkParse(b *testing.B) {
	queriesFactory := func() map[string][]byte {
		return make(map[string][]byte, initialQueryMapSize)
	}
	queriesFactoryNoAlloc := func() map[string][]byte {
		return queriesMap
	}

	singlePair := []byte("something=somewhere")
	manyPairs := []byte("something=somewhere&lorem=ipsum&good=bad&bad=good&paradox=life&dog=cat&cat=dog&life=bad")

	b.Run("SinglePair", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Parse(singlePair, queriesFactory)
		}
	})

	b.Run("ManyPairs", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Parse(manyPairs, queriesFactory)
		}
	})

	b.Run("SinglePairNoAlloc", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Parse(singlePair, queriesFactoryNoAlloc)
		}
	})

	b.Run("ManyPairsNoAlloc", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = Parse(manyPairs, queriesFactoryNoAlloc)
		}
	})
}
