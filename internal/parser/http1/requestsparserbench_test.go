package http1

import (
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

	b.Run("BiggerGET", func(b *testing.B) {
		b.SetBytes(int64(len(biggerGET)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(biggerGET)
			parser.Release()
		}
	})

	tenHeaders := []byte(
		"GET / HTTP/1.1\r\n" +
			"Header1: value1\r\n" +
			"Header2: value2\r\n" +
			"Header3: value3\r\n" +
			"ROFL Header: ROFL value\r\n" +
			"Header-5: value 5\r\n" +
			"Header-6: haha lol\r\n" +
			"Header-7: rolling out of laugh\r\n" +
			"Header-8: sometimes I just wanna chicken fries\r\n" +
			"Header-9: but having only fried potatoes instead\r\n" +
			"Header-10: and this is sometimes really annoying\r\n" +
			"\r\n",
	)

	b.Run("TenHeaders", func(b *testing.B) {
		b.SetBytes(int64(len(tenHeaders)))
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(tenHeaders)
			parser.Release()
		}
	})
}
