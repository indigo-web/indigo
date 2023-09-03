package render

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/decoder"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/parser/http1"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/settings"
	"testing"
)

type NopClientWriter struct{}

func (n NopClientWriter) Write([]byte) error {
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
	response := http.NewResponse()
	bodyReader := http1.NewBodyReader(
		dummy.NewNopClient(), http1.NewChunkedBodyParser(settings.Default().Body), decoder.NewManager(0),
	)
	request := http.NewRequest(
		hdrs, query.NewQuery(nil), http.NewResponse(), dummy.NewNopConn(),
		http.NewBody(bodyReader), nil, false,
	)
	client := NopClientWriter{}

	b.Run("DefaultResponse_NoDefHeaders", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		renderer := NewEngine(buff, nil, nil)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = renderer.Write(request.Proto, request, response, client)
		}
	})

	b.Run("DefaultResponse_1DefaultHeader", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		renderer := NewEngine(buff, nil, defaultHeadersSmall)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = renderer.Write(request.Proto, request, response, client)
		}
	})

	b.Run("DefaultResponse_3DefaultHeaders", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		renderer := NewEngine(buff, nil, defaultHeadersMedium)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = renderer.Write(request.Proto, request, response, client)
		}
	})

	b.Run("DefaultResponse_8DefaultHeaders", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		renderer := NewEngine(buff, nil, defaultHeadersBig)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = renderer.Write(request.Proto, request, response, client)
		}
	})

	b.Run("101SwitchingProtocol", func(b *testing.B) {
		resp := http.NewResponse().WithCode(status.SwitchingProtocols)
		buff := make([]byte, 0, 128)
		renderer := NewEngine(buff, nil, nil)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = renderer.Write(request.Proto, request, resp, client)
		}
	})
}
