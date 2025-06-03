package http1

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/internal/requestgen"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

type NopClientWriter struct{}

func (n NopClientWriter) Write(p []byte) (int, error) {
	return len(p), nil
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

	response := http.NewResponse()
	request := construct.Request(config.Default(), dummy.NewNopClient(), NewBody(
		dummy.NewNopClient(), nil, config.Default().Body,
	))
	client := NopClientWriter{}

	b.Run("no body no def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := newSerializer(buff, 128, nil, request, client)
		size, err := estimateResponseSize(request, response, nil)
		require.NoError(b, err)
		b.SetBytes(size)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, response)
		}
	})

	b.Run("with 4kb body", func(b *testing.B) {
		response := http.NewResponse().String(strings.Repeat("a", 4096))
		buff := make([]byte, 0, 8192)
		serializer := newSerializer(buff, 128, nil, request, client)
		respSize, err := estimateResponseSize(request, response, nil)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, response)
		}
	})

	b.Run("no body 1 def header", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := newSerializer(buff, 128, defaultHeadersSmall, request, client)
		respSize, err := estimateResponseSize(request, response, defaultHeadersSmall)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, response)
		}
	})

	b.Run("no body 3 def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := newSerializer(buff, 128, defaultHeadersMedium, request, client)
		respSize, err := estimateResponseSize(request, response, defaultHeadersMedium)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, response)
		}
	})

	b.Run("no body 8 def headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		serializer := newSerializer(buff, 128, defaultHeadersBig, request, client)
		respSize, err := estimateResponseSize(request, response, defaultHeadersBig)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, response)
		}
	})

	b.Run("no body 15 headers", func(b *testing.B) {
		buff := make([]byte, 0, 1024)
		request := construct.Request(config.Default(), dummy.NewNopClient(), NewBody(
			dummy.NewNopClient(), nil, config.Default().Body,
		))
		request.Headers = requestgen.Headers(15)
		serializer := newSerializer(buff, 128, nil, request, client)
		size, err := estimateResponseSize(request, response, nil)
		require.NoError(b, err)
		b.SetBytes(size)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = serializer.Write(request.Proto, response)
		}
	})

	b.Run("pre-write", func(b *testing.B) {
		preResp := http.NewResponse().Code(status.SwitchingProtocols)
		buff := make([]byte, 0, 128)
		serializer := newSerializer(buff, 128, nil, request, client)
		respSize, err := estimatePreWriteSize(request, preResp, response)
		require.NoError(b, err)
		b.SetBytes(respSize)
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			serializer.PreWrite(request.Proto, preResp)
			_ = serializer.Write(request.Proto, response)
		}
	})

	// TODO: add benchmarking chunked body
}

func estimateResponseSize(req *http.Request, resp *http.Response, defHeaders map[string]string) (int64, error) {
	writer := dummy.NewSinkholeWriter()
	serializer := newSerializer(nil, 128, defHeaders, req, writer)
	err := serializer.Write(req.Proto, resp)

	return int64(len(writer.Data)), err
}

func estimatePreWriteSize(
	req *http.Request, preWrite, resp *http.Response,
) (int64, error) {
	writer := dummy.NewSinkholeWriter()
	serializer := newSerializer(nil, 128, nil, req, writer)
	serializer.PreWrite(req.Proto, preWrite)
	err := serializer.Write(req.Proto, resp)

	return int64(len(writer.Data)), err
}
