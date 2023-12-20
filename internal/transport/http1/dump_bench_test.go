package http1

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/coding"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

type NopClientWriter struct{}

func (n NopClientWriter) Write([]byte) error {
	return nil
}

func BenchmarkDumper(b *testing.B) {
	defaultHeadersSmall := map[string]string{
		"Server": "indigo",
	}
	defaultHeadersMedium := map[string]string{
		"Server":           "indigo",
		"Connection":       "keep-alive",
		"Accept-Encodings": "identity",
	}
	defaultHeadersBig := map[string]string{
		"Server":           "indigo",
		"Connection":       "keep-alive",
		"Accept-Encodings": "identity",
		"Easter":           "Egg",
		"Many":             "choices, variants, ways, solutions",
		"Something":        "is not happening",
		"Talking":          "allowed",
		"Lorem":            "ipsum, doremi",
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

	b.Run("no body no def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		dumper := NewDumper(buff, nil, nil)
		respSize, err := estimateResponseSize(dumper, request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dumper.Dump(request.Proto, request, response, client)
		}
	})

	b.Run("with 4kb body", func(b *testing.B) {
		response := http.NewResponse().String(strings.Repeat("a", 4096))
		buff := make([]byte, 0, 8192)
		dumper := NewDumper(buff, nil, nil)
		respSize, err := estimateResponseSize(dumper, request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dumper.Dump(request.Proto, request, response, client)
		}
	})

	b.Run("no body 1 def header", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		dumper := NewDumper(buff, nil, defaultHeadersSmall)
		respSize, err := estimateResponseSize(dumper, request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dumper.Dump(request.Proto, request, response, client)
		}
	})

	b.Run("no body 3 def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		dumper := NewDumper(buff, nil, defaultHeadersMedium)
		respSize, err := estimateResponseSize(dumper, request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dumper.Dump(request.Proto, request, response, client)
		}
	})

	b.Run("no body 8 def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		dumper := NewDumper(buff, nil, defaultHeadersBig)
		respSize, err := estimateResponseSize(dumper, request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = dumper.Dump(request.Proto, request, response, client)
		}
	})

	b.Run("with pre-dump", func(b *testing.B) {
		preResp := http.NewResponse().Code(status.SwitchingProtocols)
		buff := make([]byte, 0, 128)
		dumper := NewDumper(buff, nil, nil)
		respSize, err := estimatePreDumpSize(dumper, request, preResp, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			dumper.PreDump(request.Proto, preResp)
			_ = dumper.Dump(request.Proto, request, response, client)
		}
	})
}

func estimateResponseSize(
	dumper *Dumper, req *http.Request, resp *http.Response,
) (int64, error) {
	writer := dummy.NewSinkholeWriter()
	err := dumper.Dump(req.Proto, req, resp, writer)

	return int64(len(writer.Data)), err
}

func estimatePreDumpSize(
	dumper *Dumper, req *http.Request, predump, resp *http.Response,
) (int64, error) {
	writer := dummy.NewSinkholeWriter()
	dumper.PreDump(req.Proto, predump)
	err := dumper.Dump(req.Proto, req, resp, writer)

	return int64(len(writer.Data)), err
}
