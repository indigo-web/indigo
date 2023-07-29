package http1

import (
	"strconv"
	"strings"
	"testing"
)

func BenchmarkHttpRequestsParser_Parse_GET(b *testing.B) {
	parser, _ := getParser()

	b.Run("SimpleGET", func(b *testing.B) {
		b.SetBytes(int64(len(simpleGET)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(simpleGET)
			parser.Release()
		}
	})

	b.Run("5 headers", func(b *testing.B) {
		request := generateRequest(5, "www.google.com", 13)
		b.SetBytes(int64(len(request)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(request)
			parser.Release()
		}
	})

	b.Run("10 headers", func(b *testing.B) {
		request := generateRequest(10, "www.google.com", 13)
		b.SetBytes(int64(len(request)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, err := parser.Parse(request)
			if err != nil {
				panic(err)
			}
			parser.Release()
		}
	})

	b.Run("50 headers", func(b *testing.B) {
		request := generateRequest(50, "www.google.com", 13)
		b.SetBytes(int64(len(request)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, err := parser.Parse(request)
			if err != nil {
				panic(err)
			}
			parser.Release()
		}
	})
}

func generateRequest(headersNum int, hostValue string, contentLengthValue int) (request []byte) {
	request = append(request,
		"GET /"+strings.Repeat("a", 500)+" HTTP/1.1\r\n"...,
	)

	for i := 0; i < headersNum-2; i++ {
		request = append(request,
			"some-random-header-name-nobody-cares-about"+strconv.Itoa(i)+": "+
				strings.Repeat("b", 100)+"\r\n"...,
		)
	}

	request = append(request, "Host: "+hostValue+"\r\n"...)
	request = append(request, "Content-Length: "+strconv.Itoa(contentLengthValue)+"\r\n"...)

	return append(request, '\r', '\n')
}
