package http1

import (
	"testing"
)

func BenchmarkHttpRequestsParser_Parse_GET(b *testing.B) {
	parser, _ := getParser()

	// ignoring all the errors because it has to be covered by tests

	simpleGET_1 := splitIntoParts(simpleGET, 1)
	simpleGET_2 := splitIntoParts(simpleGET, 2)
	simpleGET_3 := splitIntoParts(simpleGET, 3)
	simpleGET_5 := splitIntoParts(simpleGET, 5)
	simpleGET_10 := splitIntoParts(simpleGET, 10)

	testGet := func(b *testing.B, parts [][]byte) {
		for i := 0; i < b.N; i++ {
			for part := range parts {
				_, _, _ = parser.Parse(parts[part])
			}
		}
	}

	b.Run("SimpleGET_1", func(b *testing.B) {
		testGet(b, simpleGET_1)
	})

	b.Run("SimpleGET_2", func(b *testing.B) {
		testGet(b, simpleGET_2)
	})

	b.Run("SimpleGET_3", func(b *testing.B) {
		testGet(b, simpleGET_3)
	})

	b.Run("SimpleGET_5", func(b *testing.B) {
		testGet(b, simpleGET_5)
	})

	b.Run("SimpleGET_10", func(b *testing.B) {
		testGet(b, simpleGET_10)
	})

	b.Run("SimpleGET_Full", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(simpleGET)
		}
	})

	biggerGET_1 := splitIntoParts(biggerGET, 1)
	biggerGET_2 := splitIntoParts(biggerGET, 2)
	biggerGET_3 := splitIntoParts(biggerGET, 3)
	biggerGET_5 := splitIntoParts(biggerGET, 5)
	biggerGET_10 := splitIntoParts(biggerGET, 10)

	b.Run("BiggerGET_1", func(b *testing.B) {
		testGet(b, biggerGET_1)
	})

	b.Run("BiggerGET_2", func(b *testing.B) {
		testGet(b, biggerGET_2)
	})

	b.Run("BiggerGET_3", func(b *testing.B) {
		testGet(b, biggerGET_3)
	})

	b.Run("BiggerGET_5", func(b *testing.B) {
		testGet(b, biggerGET_5)
	})

	b.Run("BiggerGET_10", func(b *testing.B) {
		testGet(b, biggerGET_10)
	})

	b.Run("BiggerGET_Full", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(biggerGET)
		}
	})

	manyHeaders := []byte(
		"GET / HTTP/1.1\r\n" +
			"Header1: value1\r\n" +
			"Header2: value2\r\n" +
			"Header3: value3\r\n" +
			"ROFL Header: ROFL value\r\n" +
			"\r\n",
	)
	manyHeaders_1 := splitIntoParts(manyHeaders, 1)
	manyHeaders_2 := splitIntoParts(manyHeaders, 2)
	manyHeaders_3 := splitIntoParts(manyHeaders, 3)
	manyHeaders_5 := splitIntoParts(manyHeaders, 5)
	manyHeaders_10 := splitIntoParts(manyHeaders, 10)

	b.Run("ManyHeaders_1", func(b *testing.B) {
		testGet(b, manyHeaders_1)
	})

	b.Run("ManyHeaders_2", func(b *testing.B) {
		testGet(b, manyHeaders_2)
	})

	b.Run("ManyHeaders_3", func(b *testing.B) {
		testGet(b, manyHeaders_3)
	})

	b.Run("ManyHeaders_5", func(b *testing.B) {
		testGet(b, manyHeaders_5)
	})

	b.Run("ManyHeaders_10", func(b *testing.B) {
		testGet(b, manyHeaders_10)
	})

	b.Run("ManyHeaders_Full", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = parser.Parse(manyHeaders)
		}
	})
}
