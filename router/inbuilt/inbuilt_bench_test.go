package inbuilt

import (
	"context"
	"github.com/indigo-web/indigo/http"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/http/method"
)

func BenchmarkRouter_OnRequest_Static(b *testing.B) {
	r := New()

	GETRootRequest := getRequest()
	GETRootRequest.Path.String = "/"
	GETRootRequest.Method = method.GET
	r.Get(GETRootRequest.Path.String, http.Respond)

	longURIRequest := getRequest()
	longURIRequest.Method = method.GET
	longURIRequest.Path.String = "/" + strings.Repeat("a", 65534)
	r.Get(longURIRequest.Path.String, http.Respond)

	mediumURIRequest := getRequest()
	mediumURIRequest.Method = method.GET
	mediumURIRequest.Path.String = "/" + strings.Repeat("a", 50)
	r.Get(mediumURIRequest.Path.String, http.Respond)

	unknownURIRequest := getRequest()
	unknownURIRequest.Method = method.GET
	unknownURIRequest.Path.String = "/" + strings.Repeat("b", 65534)

	unknownMethodRequest := getRequest()
	unknownMethodRequest.Method = method.POST
	unknownMethodRequest.Path.String = longURIRequest.Path.String

	emptyCtx := context.Background()

	if err := r.OnStart(); err != nil {
		panic(err)
	}

	b.Run("GET root", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(GETRootRequest)
			GETRootRequest.Ctx = emptyCtx
		}
	})

	b.Run("GET long uri", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(longURIRequest)
			longURIRequest.Ctx = emptyCtx
		}
	})

	b.Run("GET medium uri", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(mediumURIRequest)
			mediumURIRequest.Ctx = emptyCtx
		}
	})

	b.Run("unknown uri", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(unknownURIRequest)
			unknownURIRequest.Ctx = emptyCtx
		}
	})

	b.Run("unknown method", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			r.OnRequest(unknownMethodRequest)
			unknownMethodRequest.Ctx = emptyCtx
		}
	})
}
