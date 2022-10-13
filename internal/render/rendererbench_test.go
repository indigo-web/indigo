package render

import (
	"testing"

	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/url"
	"github.com/fakefloordiv/indigo/types"
)

func nopWriter(_ []byte) error {
	return nil
}

func BenchmarkRenderer_Response(b *testing.B) {
	buff := make([]byte, 0, 1024)
	defaultHeadersSmall := map[string][]string{
		"Server": {"indigo"},
	}
	defaultHeadersMedium := map[string][]string{
		"Server":           {"indigo"},
		"Connection":       {"keep-alive"},
		"Accept-Encodings": {"identity"},
	}
	defaultHeadersBig := map[string][]string{
		"Server":           {"indigo"},
		"Connection":       {"keep-alive"},
		"Accept-Encodings": {"identity"},
		"Easter":           {"Egg"},
		"Multiple": {
			"choices",
			"variants",
			"ways",
			"solutions",
		},
		"Something": {"is not happening"},
		"Talking":   {"allowed"},
		"Lorem":     {"ipsum", "doremi"},
	}

	hdrs := headers.NewHeaders(make(map[string][]string))
	defaultRequest, _ := types.NewRequest(hdrs, url.NewQuery(nil), nil)

	b.Run("DefaultResponse_NoDefHeaders", func(b *testing.B) {
		renderer := NewRenderer(buff, nil)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			renderer.Response(defaultRequest, types.WithResponse, nopWriter)
		}
	})

	b.Run("DefaultResponse_SmallDefHeaders", func(b *testing.B) {
		renderer := NewRenderer(buff, defaultHeadersSmall)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			renderer.Response(defaultRequest, types.WithResponse, nopWriter)
		}
	})

	b.Run("DefaultResponse_MediumDefHeaders", func(b *testing.B) {
		renderer := NewRenderer(buff, defaultHeadersMedium)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			renderer.Response(defaultRequest, types.WithResponse, nopWriter)
		}
	})

	b.Run("DefaultResponse_BigDefHeaders", func(b *testing.B) {
		renderer := NewRenderer(buff, defaultHeadersBig)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			renderer.Response(defaultRequest, types.WithResponse, nopWriter)
		}
	})
}
