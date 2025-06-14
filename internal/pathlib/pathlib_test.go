package pathlib

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Benchmark(b *testing.B) {
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

func Test(t *testing.T) {
	t.Run("replace root", func(t *testing.T) {
		path := NewPath("/", "./static")
		path.Set("/index.html")
		relative := path.Relative()
		require.Equal(t, "./static/index.html", relative)
	})

	t.Run("replace static", func(t *testing.T) {
		path := NewPath("/static/", ".")
		path.Set("/static/index.html")
		relative := path.Relative()
		require.Equal(t, "./index.html", relative)
	})

	t.Run("reuse with shorter path", func(t *testing.T) {
		path := NewPath("/", "./static")
		path.Set("/index.html")
		relative := path.Relative()
		require.Equal(t, "./static/index.html", relative)

		path.Set("/index")
		relative = path.Relative()
		require.Equal(t, "./static/index", relative)
	})
}
