package inbuilt

import (
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/types"
	"strings"
	"testing"
)

func nopRender(_ types.Response) error {
	return nil
}

func BenchmarkRequestRouting(b *testing.B) {
	longURIRequest, _ := getRequest()
	longURIRequest.Method = methods.GET
	longURIRequest.Path = "/" + strings.Repeat("a", 255)

	shortURIRequest, _ := getRequest()
	shortURIRequest.Method = methods.GET
	shortURIRequest.Path = "/" + strings.Repeat("a", 15)

	unknownURIRequest, _ := getRequest()
	unknownURIRequest.Method = methods.GET
	unknownURIRequest.Path = "/" + strings.Repeat("b", 255)

	unknownMethodRequest, _ := getRequest()
	unknownMethodRequest.Method = methods.POST
	unknownMethodRequest.Path = longURIRequest.Path

	router := NewRouter()
	router.Get(longURIRequest.Path, nopHandler)
	router.Get(shortURIRequest.Path, nopHandler)

	b.Run("LongURI", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = router.OnRequest(longURIRequest, nopRender)
		}
	})

	b.Run("ShortURI", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = router.OnRequest(shortURIRequest, nopRender)
		}
	})

	b.Run("UnknownURI", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = router.OnRequest(unknownURIRequest, nopRender)
		}
	})

	b.Run("UnknownMethod", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = router.OnRequest(unknownMethodRequest, nopRender)
		}
	})
}
