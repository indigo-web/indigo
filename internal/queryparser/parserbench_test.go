package queryparser

import "testing"

func BenchmarkParse(b *testing.B) {
	const initialQueryMapSize = 10
	queriesMap := make(map[string]string, initialQueryMapSize)

	queriesFactory := func() map[string]string {
		return make(map[string]string, initialQueryMapSize)
	}
	queriesFactoryNoAlloc := func() map[string]string {
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
