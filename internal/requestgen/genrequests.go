package requestgen

import (
	"github.com/indigo-web/indigo/http/headers"
	"strconv"
	"strings"
)

func Headers(n int) headers.Headers {
	hdrs := headers.NewPrealloc(n)

	for i := 0; i < n-1; i++ {
		hdrs.Add("some-random-header-name-nobody-cares-about"+strconv.Itoa(i), strings.Repeat("b", 100))
	}

	return hdrs.Add("Host", "localhost")
}

func HeadersBlock(hdrs headers.Headers) (buff []byte) {
	for _, pair := range hdrs.Unwrap() {
		buff = append(buff, pair.Key+": "+pair.Value+"\r\n"...)
	}

	return buff
}

func Generate(uri string, hdrs headers.Headers) (request []byte) {
	request = append(request, "GET /"+uri+" HTTP/1.1\r\n"...)
	request = append(request, HeadersBlock(hdrs)...)

	return append(request, '\r', '\n')
}
