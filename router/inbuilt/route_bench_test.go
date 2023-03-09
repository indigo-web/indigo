package inbuilt

import (
	"strings"
	"testing"

	methods "github.com/indigo-web/indigo/http/method"
)

func BenchmarkRequestRouting(b *testing.B) {
	longURIRequest := getRequest()
	longURIRequest.Method = methods.GET
	longURIRequest.Path.String = "/" + strings.Repeat("a", 255)

	shortURIRequest := getRequest()
	shortURIRequest.Method = methods.GET
	shortURIRequest.Path.String = "/" + strings.Repeat("a", 15)

	unknownURIRequest := getRequest()
	unknownURIRequest.Method = methods.GET
	unknownURIRequest.Path.String = "/" + strings.Repeat("b", 255)

	unknownMethodRequest := getRequest()
	unknownMethodRequest.Method = methods.POST
	unknownMethodRequest.Path.String = longURIRequest.Path.String

	router := NewRouter()
	router.Get(longURIRequest.Path.String, nopHandler)
	router.Get(shortURIRequest.Path.String, nopHandler)

	router.OnStart()

	b.Run("LongURI", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.OnRequest(longURIRequest)
		}
	})

	b.Run("ShortURI", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.OnRequest(shortURIRequest)
		}
	})

	b.Run("UnknownURI", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.OnRequest(unknownURIRequest)
		}
	})

	b.Run("UnknownMethod", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			router.OnRequest(unknownMethodRequest)
		}
	})
}
