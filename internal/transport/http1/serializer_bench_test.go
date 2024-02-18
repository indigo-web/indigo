package http1

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
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

func BenchmarkSerializer(b *testing.B) {
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

	hdrs := headers.New()
	response := http.NewResponse()
	body := NewBody(
		dummy.NewNopClient(), nil, config.Default().Body,
	)
	request := http.NewRequest(
		hdrs, query.NewQuery(nil), http.NewResponse(), dummy.NewNopConn(),
		body, nil,
	)
	client := NopClientWriter{}

	b.Run("no body no def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := NewSerializer(buff, 128, nil)
		respSize, err := estimateResponseSize(serializer, request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, request, response, client)
		}
	})

	b.Run("with 4kb body", func(b *testing.B) {
		response := http.NewResponse().String(strings.Repeat("a", 4096))
		buff := make([]byte, 0, 8192)
		serializer := NewSerializer(buff, 128, nil)
		respSize, err := estimateResponseSize(serializer, request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, request, response, client)
		}
	})

	b.Run("no body 1 def header", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := NewSerializer(buff, 128, defaultHeadersSmall)
		respSize, err := estimateResponseSize(serializer, request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, request, response, client)
		}
	})

	b.Run("no body 3 def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := NewSerializer(buff, 128, defaultHeadersMedium)
		respSize, err := estimateResponseSize(serializer, request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, request, response, client)
		}
	})

	b.Run("no body 8 def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := NewSerializer(buff, 128, defaultHeadersBig)
		respSize, err := estimateResponseSize(serializer, request, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, request, response, client)
		}
	})

	b.Run("pre-write", func(b *testing.B) {
		preResp := http.NewResponse().Code(status.SwitchingProtocols)
		buff := make([]byte, 0, 128)
		serializer := NewSerializer(buff, 128, nil)
		respSize, err := estimatePreWriteSize(serializer, request, preResp, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			serializer.PreWrite(request.Proto, preResp)
			_ = serializer.Write(request.Proto, request, response, client)
		}
	})

	// TODO: add benchmarking chunked body
}

func estimateResponseSize(
	serializer *Serializer, req *http.Request, resp *http.Response,
) (int64, error) {
	writer := dummy.NewSinkholeWriter()
	err := serializer.Write(req.Proto, req, resp, writer)

	return int64(len(writer.Data)), err
}

func estimatePreWriteSize(
	serializer *Serializer, req *http.Request, preWrite, resp *http.Response,
) (int64, error) {
	writer := dummy.NewSinkholeWriter()
	serializer.PreWrite(req.Proto, preWrite)
	err := serializer.Write(req.Proto, req, resp, writer)

	return int64(len(writer.Data)), err
}
