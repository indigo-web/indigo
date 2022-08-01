package parser

import (
	"github.com/stretchr/testify/require"
	"indigo/errors"
	"testing"
)

func TestParserReuseAbilityChunkedRequest(t *testing.T) {
	parser, request := getParser()

	rawRequest := []byte("POST / HTTP/1.1\r\n" +
		"Content-Type: some content type\r\n" +
		"Host: indigo.dev\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"\r\nd\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n")

	wantedRequest := WantedRequest{
		Method:   "POST",
		Path:     "/",
		Protocol: "HTTP/1.1",
		Headers: testHeaders{
			"content-type":      "some content type",
			"host":              "indigo.dev",
			"transfer-encoding": "chunked",
		},
		Body:                 "Hello, world!But what's wrong with you?Finally am here",
		StrictHeadersCompare: true,
	}

	bodyChan := make(chan []byte, 1)
	go readBody(request, bodyChan)
	err, _ := FeedParser(parser, rawRequest, 5)

	require.NoError(t, err)
	require.NoError(t, compareRequests(wantedRequest, *request))
	require.Equal(t, []byte(wantedRequest.Body), <-bodyChan)

	request.Reset()

	go readBody(request, bodyChan)
	err, _ = FeedParser(parser, rawRequest, 5)

	require.NoError(t, err)
	require.NoError(t, compareRequests(wantedRequest, *request))
	require.Equal(t, []byte(wantedRequest.Body), <-bodyChan)
}

func TestChunkedTransferEncodingFullRequestBody(t *testing.T) {
	rawRequest := "POST / HTTP/1.1\r\n" +
		"Content-Type: some content type\r\n" +
		"Host: indigo.dev\r\n" +
		"Transfer-Encoding: chunked\r\n" +
		"\r\nd\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n"

	wantBody := "d\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n"
	wantedRequest := WantedRequest{
		Method:   "POST",
		Path:     "/",
		Protocol: "HTTP/1.1",
		Headers: testHeaders{
			"content-type":      "some content type",
			"host":              "indigo.dev",
			"transfer-encoding": "chunked",
		},
		StrictHeadersCompare: true,
	}

	parser, request := getParser()
	state, extra, err := parser.Parse([]byte(rawRequest))

	require.Equal(t, RequestCompleted, state, "wanted completion flag but got false")
	require.Equal(t, []byte(wantBody), extra)
	require.NoError(t, err)
	require.NoError(t, compareRequests(wantedRequest, *request))
	require.Equal(t, []byte(wantBody), extra)
}

func TestChunkOverflow(t *testing.T) {
	request, pipe := newRequest()
	parser := NewChunkedBodyParser(pipe, 65535)
	data := []byte("d\r\nHello, world! Overflow here\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n")

	bodyChan := make(chan []byte, 5)
	go readBody(&request, bodyChan)
	done, _, err := parser.Feed(data)

	require.True(t, done, "wanted completion flag but got false")
	require.ErrorIs(t, errors.ErrInvalidChunkSplitter, err)

	<-bodyChan
}

func TestChunkTooSmall(t *testing.T) {
	request, pipe := newRequest()
	parser := NewChunkedBodyParser(pipe, 65535)
	data := []byte("d\r\nHello, ...\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n")

	bodyChan := make(chan []byte, 5)
	go readBody(&request, bodyChan)
	done, _, err := parser.Feed(data)

	require.True(t, done, "wanted completion flag but got false")
	require.ErrorIs(t, errors.ErrInvalidChunkSplitter, err)

	<-bodyChan
}

func TestMixChunkSplitters(t *testing.T) {
	request, pipe := newRequest()
	parser := NewChunkedBodyParser(pipe, 65535)
	data := []byte("d\r\nHello, world!\n1a\r\nBut what's wrong with you?\nf\nFinally am here\r\n0\r\n\n")

	bodyChan := make(chan []byte, 5)
	go readBody(&request, bodyChan)
	done, _, err := parser.Feed(data)

	require.True(t, done, "wanted completion flag but got false")
	require.NoError(t, err)

	<-bodyChan
}

func TestWithDifferentBlockSizes(t *testing.T) {
	data := []byte("d\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n")

	for i := 1; i <= len(data); i++ {
		request, pipe := newRequest()
		bodyChan := make(chan []byte, 5)
		go readBody(&request, bodyChan)
		parser := NewChunkedBodyParser(pipe, 65535)

		for j := 0; j < len(data); j += i {
			end := j + i
			if end > len(data) {
				end = len(data)
			}

			done, _, err := parser.Feed(data[j:end])

			require.NoError(t, err)
			require.False(t, done && end < len(data), "unwanted completion flag")
		}

		<-bodyChan
	}
}
