package http1

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/coding"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"testing"
)

type NopClientWriter struct{}

func (n NopClientWriter) Write([]byte) error {
	return nil
}

func BenchmarkDumper(b *testing.B) {
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

	hdrs := headers.NewHeaders()
	response := http.NewResponse()
	body := NewBody(
		dummy.NewNopClient(), nil, coding.NewManager(0),
	)
	request := http.NewRequest(
		hdrs, query.NewQuery(nil), http.NewResponse(), dummy.NewNopConn(),
		body, nil, false,
	)
	client := NopClientWriter{}

	b.Run("DefaultResponse_NoDefHeaders", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		dumper := NewDumper(buff, nil, nil)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dumper.Dump(request.Proto, request, response, client)
		}
	})

	b.Run("DefaultResponse_1DefaultHeader", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		dumper := NewDumper(buff, nil, defaultHeadersSmall)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dumper.Dump(request.Proto, request, response, client)
		}
	})

	b.Run("DefaultResponse_3DefaultHeaders", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		dumper := NewDumper(buff, nil, defaultHeadersMedium)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dumper.Dump(request.Proto, request, response, client)
		}
	})

	b.Run("DefaultResponse_8DefaultHeaders", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		dumper := NewDumper(buff, nil, defaultHeadersBig)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dumper.Dump(request.Proto, request, response, client)
		}
	})

	b.Run("101SwitchingProtocol", func(b *testing.B) {
		resp := http.NewResponse().Code(status.SwitchingProtocols)
		buff := make([]byte, 0, 128)
		dumper := NewDumper(buff, nil, nil)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dumper.Dump(request.Proto, request, resp, client)
		}
	})
}
