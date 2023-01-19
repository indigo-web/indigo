package http

import (
	"github.com/fakefloordiv/indigo/internal/server/tcp/dummy"
	"net"
	"testing"
	"time"

	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/internal/pool"

	"github.com/fakefloordiv/indigo/http/headers"
	"github.com/fakefloordiv/indigo/http/url"
	"github.com/fakefloordiv/indigo/internal/alloc"
	"github.com/fakefloordiv/indigo/internal/parser/http1"
	render2 "github.com/fakefloordiv/indigo/internal/render"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/settings"
)

var (
	simpleGETRequest      = []byte("GET / HTTP/1.1\r\n\r\n")
	fiveHeadersGETRequest = []byte(
		"GET / HTTP/1.1\r\n" +
			"Hello: world\r\n" +
			"One: ok\r\n" +
			"Content-Type: nothing but true;q=0.9\r\n" +
			"Four: lorem ipsum\r\n" +
			"Mistake: is made here\r\n" +
			"\r\n",
	)
	tenHeadersGETRequest = []byte(
		"GET / HTTP/1.1\r\n" +
			"Hello: world\r\n" +
			"One: ok\r\n" +
			"Content-Type: nothing but true;q=0.9\r\n" +
			"Four: lorem ipsum\r\n" +
			"Mistake: is made here\r\n" +
			"Lorem: upsum\r\n" +
			"tired: of all this shit\r\n" +
			"Eight: finally only two left\r\n" +
			"my-brain: is not so creative\r\n" +
			"to-create: ten random headers from scratch\r\n" +
			"\r\n",
	)
	simpleGETWithHeader = []byte("GET /with-header HTTP/1.1\r\n\r\n")

	simplePOST = []byte("POST / HTTP/1.1\r\nContent-Length: 13\r\n\r\nHello, world!")
)

type connMock struct {
	data []byte
}

func newConn(data []byte) net.Conn {
	return connMock{
		data: data,
	}
}

func (c connMock) Read(b []byte) (n int, err error) {
	copy(b, c.data)

	return len(b), nil
}

func (connMock) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (connMock) Close() error {
	return nil
}

func (connMock) LocalAddr() net.Addr {
	return nil
}

func (connMock) RemoteAddr() net.Addr {
	return nil
}

func (connMock) SetDeadline(time.Time) error {
	return nil
}

func (connMock) SetReadDeadline(time.Time) error {
	return nil
}

func (connMock) SetWriteDeadline(time.Time) error {
	return nil
}

func BenchmarkIndigo(b *testing.B) {
	router := inbuilt.NewRouter()
	root := router.Resource("/")
	root.Get(http.RespondTo)
	root.Post(func(request *http.Request) http.Response {
		_ = request.OnBody(func([]byte) error {
			return nil
		}, func() error {
			return nil
		})

		return http.RespondTo(request)
	})

	router.Get("/with-header", func(request *http.Request) http.Response {
		return http.RespondTo(request).WithHeader("Hello", "World")
	})

	router.Get("/with-two-headers", func(request *http.Request) http.Response {
		return http.RespondTo(request).
			WithHeader("Hello", "World").
			WithHeader("Lorem", "Ipsum")
	})

	router.OnStart()

	s := settings.Default()
	query := url.NewQuery(func() map[string][]byte {
		return make(map[string][]byte)
	})
	bodyReader := http1.NewBodyReader(dummy.NewNopClient(), settings.Default().Body)
	request := http.NewRequest(
		headers.NewHeaders(nil), query, http.NewResponse(), dummy.NewNopConn(), bodyReader,
	)
	keyAllocator := alloc.NewAllocator(
		s.Headers.MaxKeyLength*s.Headers.Number.Default,
		s.Headers.MaxKeyLength*s.Headers.Number.Maximal,
	)
	valAllocator := alloc.NewAllocator(
		int(s.Headers.ValueSpace.Default), int(s.Headers.ValueSpace.Maximal),
	)
	objPool := pool.NewObjectPool[[]string](20)
	startLineBuff := make([]byte, s.URL.MaxLength)
	parser := http1.NewHTTPRequestsParser(
		request, keyAllocator, valAllocator, objPool, startLineBuff, s.Headers,
	)

	client := dummy.NewNopClient()
	render := render2.NewRenderer(make([]byte, 0, 1024), nil, make(map[string][]string))

	server := NewHTTPServer(router).(*httpServer)
	go server.Run(client, request, bodyReader, render, parser)

	simpleGETClient := dummy.NewCircularClient(simpleGETRequest)
	b.Run("SimpleGET", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			server.RunOnce(simpleGETClient, request, bodyReader, render, parser)
		}
	})

	fiveHeadersGETClient := dummy.NewCircularClient(fiveHeadersGETRequest)
	b.Run("FiveHeadersGET", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			server.RunOnce(fiveHeadersGETClient, request, bodyReader, render, parser)
		}
	})

	tenHeadersGETClient := dummy.NewCircularClient(tenHeadersGETRequest)
	b.Run("TenHeadersGET", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			server.RunOnce(tenHeadersGETClient, request, bodyReader, render, parser)
		}
	})

	withRespHeadersGETClient := dummy.NewCircularClient(simpleGETWithHeader)
	b.Run("WithRespHeader", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			server.RunOnce(withRespHeadersGETClient, request, bodyReader, render, parser)
		}
	})

	// TODO: add here some special request with special client that is able to
	//       parse a body
	//b.Run("SimplePOST", func(b *testing.B) {
	//	for i := 0; i < b.N; i++ {
	//		_ = server.OnData(simplePOST)
	//	}
	//})
}
