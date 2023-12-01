package pathlib

import "testing"

func BenchmarkPathlib(b *testing.B) {
	path := NewPath("/", "/static")
	given := "/index.html"

	b.Run("replace prefixes", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			path.Set(given)
			_ = path.Relative()
		}
	})
}
