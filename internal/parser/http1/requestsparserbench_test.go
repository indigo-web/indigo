package http1

import (
	"testing"
)

func BenchmarkHttpRequestsParser_Parse_GET(b *testing.B) {
	parser, _ := getParser()

	// ignoring all the errors because it has to be covered by tests

	simpleGET_1 := splitIntoParts(simpleGET, 1)
	simpleGET_10 := splitIntoParts(simpleGET, 10)

	testGet := func(b *testing.B, parts [][]byte) {
		for i := 0; i < b.N; i++ {
			for part := range parts {
				_, _, _ = parser.Parse(parts[part])
			}

			parser.Release()
		}
	}

	b.Run("SimpleGET_1", func(b *testing.B) {
		testGet(b, simpleGET_1)
	})

	b.Run("SimpleGET_10", func(b *testing.B) {
		testGet(b, simpleGET_10)
	})

	b.Run("SimpleGET_Full", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(simpleGET)
			parser.Release()
		}
	})

	biggerGET_1 := splitIntoParts(biggerGET, 1)
	biggerGET_10 := splitIntoParts(biggerGET, 10)

	b.Run("BiggerGET_1", func(b *testing.B) {
		testGet(b, biggerGET_1)
	})

	b.Run("BiggerGET_10", func(b *testing.B) {
		testGet(b, biggerGET_10)
	})

	b.Run("BiggerGET_Full", func(b *testing.B) {
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

	b.Run("TenHeaders_Full", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(tenHeaders)
			parser.Release()
		}
	})
}
