package rmap

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/method"
	"strings"
	"testing"
)

func BenchmarkRoutesMap(b *testing.B) {
	const (
		shortPath  = "/"
		mediumPath = "/hello/world/some/text/here"
	)

	var (
		longPath    = "/" + strings.Repeat("a", 65534)
		unknownPath = longPath[:len(longPath)-1]
	)

	m := New()
	m.Add(shortPath, method.GET, http.Respond)
	m.Add(mediumPath, method.GET, http.Respond)
	m.Add(longPath, method.GET, http.Respond)

	b.Run("short path", func(b *testing.B) {
		b.SetBytes(int64(len(shortPath)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m.Get(shortPath)
		}
	})

	b.Run("medium path", func(b *testing.B) {
		b.SetBytes(int64(len(mediumPath)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m.Get(mediumPath)
		}
	})

	b.Run("long path", func(b *testing.B) {
		b.SetBytes(int64(len(longPath)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m.Get(longPath)
		}
	})

	b.Run("unknown path", func(b *testing.B) {
		b.SetBytes(int64(len(unknownPath)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			m.Get(unknownPath)
		}
	})
}
