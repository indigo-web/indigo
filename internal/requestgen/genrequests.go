package requestgen

import (
	"strconv"
	"strings"
)

func Generate(uri string, headersNum int) (request []byte) {
	request = append(request, "GET /"+uri+" HTTP/1.1\r\n"...)

	for i := 0; i < headersNum-1; i++ {
		request = append(request,
			"some-random-header-name-nobody-cares-about"+strconv.Itoa(i)+": "+
				strings.Repeat("b", 100)+"\r\n"...,
		)
	}

	request = append(request, "Host: www.google.com\r\n"...)

	return append(request, '\r', '\n')
}
