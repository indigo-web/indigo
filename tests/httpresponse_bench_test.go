package tests

import (
	"bytes"
	"indigo/internal"
	"testing"
)

func RenderHTTPResponse(buff []byte,
	protocol, code, status []byte,
	headers map[string][]byte,
	body []byte) []byte {
	buff = append(buff, protocol...)
	buff = append(buff, code...)
	buff = append(buff, status...)

	for key, value := range headers {
		buff = append(append(append(buff, internal.S2B(key)...), ':', ' '), value...)
		buff = append(buff, '\n')
	}

	return append(buff, body...)
}

func RenderHTTPResponseGoodHeaders(buff []byte,
	protocol, code, status []byte,
	headers map[string][]byte,
	body []byte) []byte {
	buff = append(append(append(buff, protocol...), code...), status...)

	for key, value := range headers {
		buff = append(append(buff, internal.S2B(key)...), value...)
	}

	return append(buff, body...)
}

func RenderHTTPResponseBytesBufferMut(buff bytes.Buffer,
	protocol, code, status []byte,
	headers map[string][]byte,
	body []byte) {
	buff.Write(protocol)
	buff.Write(code)
	buff.Write(status)

	for key, value := range headers {
		buff.Write(internal.S2B(key))
		buff.Write([]byte(": "))
		buff.Write(value)
	}

	buff.Write(body)
}

func BenchmarkRenderHTTPResponse(b *testing.B) {
	//listener, err := getTCPSock("localhost", 5005)
	//if err != nil {
	//	b.Fatalf("unexpected error: %s", err)
	//}
	//defer listener.Close()
	//
	//go idleTCPConn("localhost", 5005)
	//conn, err := listener.Accept()
	//if err != nil {
	//	b.Fatalf("unexpected error on accepting connection: %s", err)
	//}
	//defer conn.Close()
	buff := make([]byte, 0, 200)
	var (
		protocol = []byte("HTTP/1.1 ")
		code     = []byte("200 ")
		status   = []byte("OK\n")
		headers  = map[string][]byte{
			"Authorization": []byte("good"),
			"Server":        []byte("indigo"),
			"Wassup":        []byte("good"),
			"Easter":        []byte("egg"),
		}
		body = []byte("Hello, world! Lorem ipsum!")
	)

	b.Run("SourceSolution", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderHTTPResponse(buff, protocol, code, status, headers, body)
			buff = buff[:0]
		}
	})

	b.Run("PreRenderedHeaders", func(b *testing.B) {
		headers = map[string][]byte{
			"Authorization: ": []byte("good\n"),
			"Server: ":        []byte("indigo\n"),
			"Wassup: ":        []byte("good\n"),
			"Easter: ":        []byte("egg\n"),
		}

		for i := 0; i < b.N; i++ {
			RenderHTTPResponseGoodHeaders(buff, protocol, code, status, headers, body)
			buff = buff[:0]
		}
	})

	bytesBuff := bytes.Buffer{}
	bytesBuff.Grow(200)

	b.Run("BytesBuffer", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RenderHTTPResponseBytesBufferMut(bytesBuff, protocol, code, status, headers, body)
			bytesBuff.Reset()
		}
	})
}
