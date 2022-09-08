package render

import (
	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/url"
	"github.com/fakefloordiv/indigo/settings"
	"github.com/fakefloordiv/indigo/types"
	"testing"
)

func nopWriter(_ []byte) error {
	return nil
}

func BenchmarkRenderer_Response(b *testing.B) {
	buff := make([]byte, 0, 1024)
	defaultHeadersSmall := headers.Headers{
		"Server": {"indigo"},
	}
	defaultHeadersMedium := headers.Headers{
		"Server":           {"indigo"},
		"Connection":       {"keep-alive"},
		"Accept-Encodings": {"identity"},
	}
	defaultHeadersBig := headers.Headers{
		"Server":           {"indigo"},
		"Connection":       {"keep-alive"},
		"Accept-Encodings": {"identity"},
		"Easter":           {"Egg"},
		"Multiple":         {"choices", "variants", "ways"},
		"Something":        {"is not happening"},
		"Talking":          {"allowed"},
		"Lorem":            {"ipsum", "doremi"},
	}

	manager := headers.NewManager(settings.Default().Headers)
	defaultRequest, _ := types.NewRequest(&manager, url.NewQuery(nil))

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