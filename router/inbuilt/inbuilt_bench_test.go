package inbuilt

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/parser/http1"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/settings"
	"testing"
)

func BenchmarkRouter_OnRequest_Static(b *testing.B) {
	r := NewRouter()
	r.Get("/", http.RespondTo)
	r.Post("/", http.RespondTo)
	r.Get(
		"/some/very/long/path/that/is/not/gonna/end/somewhere/in/close/future/or/no/haha/I/lied",
		http.RespondTo,
	)

	r.OnStart()

	body := http1.NewBodyReader(dummy.NewNopClient(), settings.Default().Body)
	GETShortReq := http.NewRequest(
		headers.NewHeaders(nil), query.Query{}, http.NewResponse(), dummy.NewNopConn(), body,
		nil, false,
	)
	GETShortReq.Path.String = "/"
	GETShortReq.Method = method.GET

	POSTShortReq := http.NewRequest(
		headers.NewHeaders(nil), query.Query{}, http.NewResponse(), dummy.NewNopConn(), body,
		nil, false,
	)
	POSTShortReq.Path.String = "/"
	POSTShortReq.Method = method.POST

	GETLongReq := http.NewRequest(
		headers.NewHeaders(nil), query.Query{}, http.NewResponse(), dummy.NewNopConn(), body,
		nil, false,
	)
	GETLongReq.Path.String = "/some/very/long/path/that/is/not/gonna/end/somewhere/in/close/future/or/no/haha/I/lied"
	GETLongReq.Method = method.GET

	b.Run("GET_Static_Short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(GETShortReq)
		}
	})

	b.Run("POST_Static_Short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(POSTShortReq)
		}
	})

	b.Run("GET_Static_Long", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(GETLongReq)
		}
	})
}
