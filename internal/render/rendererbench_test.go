package render

import (
	"github.com/fakefloordiv/indigo/http"
	"testing"

	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/url"
)

func nopWriter(_ []byte) error {
	return nil
}

func BenchmarkRenderer_Response(b *testing.B) {
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
	defaultResponse := http.NewResponse()
	defaultRequest, _ := http.NewRequest(
		hdrs, url.NewQuery(nil), nil, nil, defaultResponse,
	)

	b.Run("DefaultResponse_NoDefHeaders", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		renderer := NewRenderer(buff, nil, nil)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = renderer.Response(defaultRequest, defaultResponse, nopWriter)
		}
	})

	b.Run("DefaultResponse_1DefaultHeader", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		renderer := NewRenderer(buff, nil, defaultHeadersSmall)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = renderer.Response(defaultRequest, defaultResponse, nopWriter)
		}
	})

	b.Run("DefaultResponse_3DefaultHeaders", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		renderer := NewRenderer(buff, nil, defaultHeadersMedium)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = renderer.Response(defaultRequest, defaultResponse, nopWriter)
		}
	})

	b.Run("DefaultResponse_8DefaultHeaders", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		renderer := NewRenderer(buff, nil, defaultHeadersBig)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = renderer.Response(defaultRequest, defaultResponse, nopWriter)
		}
	})
}
