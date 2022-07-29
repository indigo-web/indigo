package parser

import (
	"bytes"
	"fmt"
	"indigo/errors"
	"indigo/http"
	"indigo/internal"
	"indigo/settings"
	"indigo/types"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type testHeaders map[string]string

type parserer interface {
	Parse([]byte) (done bool, extra []byte, err error)
}

func FeedParser(parser parserer, data []byte, chunksSize int) (err error, extra []byte) {
	for i := 0; i < len(data); i += chunksSize {
		end := i + chunksSize

		if end > len(data) {
			end = len(data)
		}

		rawPiece := data[i:end]
		piece := make([]byte, len(rawPiece))
		copy(piece, rawPiece)

		_, extra, err = parser.Parse(piece)

		if err != nil {
			return err, extra
		}
	}

	return nil, extra
}

type WantedRequest struct {
	Method   string
	Path     string
	Protocol string
	Headers  testHeaders
	Body     string

	StrictHeadersCompare bool
}

func quote(data []byte) string {
	return strconv.Quote(string(data))
}

func newRequest() (types.Request, *internal.Pipe) {
	return types.NewRequest(nil, make(http.Headers, 5), nil, 0)
}

func getParser() (HTTPRequestsParser, *types.Request) {
	request, pipe := newRequest()
	parserSettings := settings.Prepare(settings.Settings{})
	return NewHTTPParser(&request, pipe, parserSettings), &request
}

func readBody(request *types.Request, bodyChan chan []byte) {
	body, _ := request.GetFullBody()
	bodyChan <- body
}

func compareRequests(wanted WantedRequest, body []byte, got types.Request) error {
	if http.Bytes2Method([]byte(wanted.Method)) != got.Method {
		gotMethod := http.Method2String(got.Method)
		return fmt.Errorf(
			"methods mismatching: wanted %s, got %s",
			wanted.Method, gotMethod)
	}
	if !bytes.Equal([]byte(wanted.Protocol+" "), got.Protocol.Raw()) {
		return fmt.Errorf(
			"protocols mismatching: wanted %s, got %s",
			strconv.Quote(wanted.Protocol), quote(got.Protocol.Raw()))
	}

	for key, value := range got.Headers {
		wantedValue, found := wanted.Headers[key]
		if !found {
			if wanted.StrictHeadersCompare {
				return fmt.Errorf(
					"unwanted header: %s (strict check)",
					strconv.Quote(key))
			}

			continue
		}

		if wantedValue != string(value) {
			return fmt.Errorf(
				"mismatching header %s values: wanted %s, got %s",
				key, strconv.Quote(wantedValue), quote(value))
		}
	}

	if len(wanted.Body) != len(body) {
		return fmt.Errorf(
			"mismatching bodies: wanted %s, got %s",
			strconv.Quote(wanted.Body), quote(body))
	}

	return nil
}

func testOrdinaryGETRequestParse(t *testing.T, chunkSize int) {
	parser, request := getParser()

	wantedRequest := WantedRequest{
		Method:   "GET",
		Path:     "/",
		Protocol: "HTTP/1.1",
		Headers: testHeaders{
			"Content-Type": "some content type",
			"Host":         "indigo.dev",
		},
		Body: "",
	}

	ordinaryGetRequest := []byte("GET / HTTP/1.1\r\nContent-Type: some content type\r\nHost: indigo.dev\r\n\r\n")

	if chunkSize == -1 {
		chunkSize = len(ordinaryGetRequest) + 1
	}

	bodyChan := make(chan []byte, 1)
	go readBody(request, bodyChan)
	err, extra := FeedParser(parser, ordinaryGetRequest, chunkSize)

	require.Nil(t, err, "unwanted error from parser")
	require.Empty(t, extra, "unwanted extra")
	require.Nil(t, compareRequests(wantedRequest, <-bodyChan, *request))
}

func TestOrdinaryGETRequestParse1Char(t *testing.T) {
	testOrdinaryGETRequestParse(t, 1)
}

func TestOrdinaryGETRequestParse2Chars(t *testing.T) {
	testOrdinaryGETRequestParse(t, 2)
}

func TestOrdinaryGETRequestParse5Chars(t *testing.T) {
	testOrdinaryGETRequestParse(t, 5)
}

func TestOrdinaryGETRequestParseFull(t *testing.T) {
	testOrdinaryGETRequestParse(t, -1)
}

func testInvalidGETRequest(t *testing.T, rawRequest []byte, errorWanted error) {
	parser, _ := getParser()
	err, extra := FeedParser(parser, rawRequest, 5)

	require.Empty(t, extra, "unwanted extra")
	require.Error(t, err, errorWanted)
}

func TestInvalidGETRequestMissingMethod(t *testing.T) {
	request := []byte("/ HTTP/1.1\r\nContent-Type: some content type\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidMethod)
}

func TestInvalidGETRequestEmptyMethod(t *testing.T) {
	request := []byte(" / HTTP/1.1\r\nContent-Type: some content type\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidMethod)
}

func TestInvalidGETRequestInvalidMethod(t *testing.T) {
	request := []byte("GETP / HTTP/1.1\r\nContent-Type: some content type\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidMethod)
}

func TestInvalidPOSTRequestExtraBody(t *testing.T) {
	rawRequest := []byte("POST / HTTP/1.1\r\nHost: indigo.dev\r\nContent-Length: 13\r\n\r\nHello, world! Extra body")
	parser, request := getParser()
	go readBody(request, make(chan []byte, 1))
	err, extra := FeedParser(parser, rawRequest, 5)

	require.Empty(t, extra, "unwanted extra")
	require.Error(t, err, errors.ErrInvalidMethod)
}

func TestInvalidGETRequestUnknownProtocol(t *testing.T) {
	request := []byte("GET / HTTP/1.2\r\nContent-Type: some content type\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrProtocolNotSupported)
}

func TestInvalidGETRequestEmptyPath(t *testing.T) {
	request := []byte("GET  HTTP/1.1\r\nContent-Type: some content type\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidPath)
}

func TestInvalidGETRequestMissingPath(t *testing.T) {
	request := []byte("GET HTTP/1.2\r\nContent-Type: some content type\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidPath)
}

func TestInvalidGETRequestInvalidHeaderNoColon(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\nContent-Type some content type\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidHeader)
}

func TestInvalidGETRequestInvalidHeaderEmptyKey(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n:some content type\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidHeader)
}

func TestInvalidGETRequestInvalidHeaderEmptyValue(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n\bContent-Type:\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidHeader)
}

func TestInvalidGETRequestInvalidHeaderNonPrintable(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\nContent\b-Type: some content type\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidHeader)
}

func TestInvalidGETRequestInvalidHeaderFirstCharNonPrintable(t *testing.T) {
	request := []byte("GET / HTTP/1.1\r\n\bContent-Type: some content type\r\nHost: indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidHeader)
}

func TestInvalidGETRequestNoSpaces(t *testing.T) {
	request := []byte("GET/HTTP/1.1\r\nContent-Typesomecontenttype\r\nHost:indigo.dev\r\n\r\n")
	testInvalidGETRequest(t, request, errors.ErrInvalidMethod)
}

func testOrdinaryPOSTRequestParse(t *testing.T, chunkSize int) {
	parser, request := getParser()
	ordinaryGetRequest := []byte("POST / HTTP/1.1\r\nContent-Type: some content type\r\nHost: indigo.dev" +
		"\r\nContent-Length: 13\r\n\r\nHello, world!")

	wantedRequest := WantedRequest{
		Method:   "POST",
		Path:     "/",
		Protocol: "HTTP/1.1",
		Headers: testHeaders{
			"Content-Type":   "some content type",
			"Host":           "indigo.dev",
			"Content-Length": "13",
		},
		Body:                 "Hello, world!",
		StrictHeadersCompare: true,
	}

	if chunkSize == -1 {
		chunkSize = len(ordinaryGetRequest)
	}

	bodyChan := make(chan []byte, 1)
	go readBody(request, bodyChan)
	err, extra := FeedParser(parser, ordinaryGetRequest, chunkSize)

	require.Nil(t, err)
	require.Empty(t, extra, "unwanted extra")
	require.Nil(t, compareRequests(wantedRequest, <-bodyChan, *request))
}

func TestOrdinaryPOSTRequestParse1Char(t *testing.T) {
	testOrdinaryPOSTRequestParse(t, 1)
}

func TestOrdinaryPOSTRequestParse2Chars(t *testing.T) {
	testOrdinaryPOSTRequestParse(t, 2)
}

func TestOrdinaryPOSTRequestParse5Chars(t *testing.T) {
	testOrdinaryPOSTRequestParse(t, 5)
}

func TestOrdinaryPOSTRequestParseFull(t *testing.T) {
	testOrdinaryPOSTRequestParse(t, -1)
}

func TestChromeGETRequest(t *testing.T) {
	rawRequest := "GET / HTTP/1.1\r\nHost: localhost:8080\r\nConnection: keep-alive\r\nCache-Control: max-age=0" +
		"\r\nsec-ch-ua: \" Not A;Brand\";v=\"99\", \"Chromium\";v=\"96\", \"Google Chrome\";v=\"96\"" +
		"\r\nsec-ch-ua-mobile: ?0\r\nsec-ch-ua-platform: \"Windows\"\r\nUpgrade-Insecure-Requests: 1" +
		"\r\nUser-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96" +
		".0.4664.110 Safari/537.36\r\n" +
		"Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8," +
		"application/signed-exchange;v=b3;q=0.9\r\nSec-Fetch-Site: none\r\nSec-Fetch-Mode: navigate" +
		"\r\nSec-Fetch-User: ?1\r\nSec-Fetch-Dest: document\r\nAccept-Encoding: gzip, deflate, br" +
		"\r\nAccept-Language: ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7,uk;q=0.6\r\nCookie: csrftoken=y1y3SinAMbiYy7yn9Oc" +
		"blqbudgdgdgdgddgdgdgdgdsgsdgsdgdgddgdGWnwfuDG; Goland-1dc491b=e03b2dgdgvdfgad0-b7ab-e4f8e1715c8b\r\n\r\n"

	parser, request := getParser()
	bodyChan := make(chan []byte, 1)
	go readBody(request, bodyChan)
	done, extra, err := parser.Parse([]byte(rawRequest))

	wantedRequest := WantedRequest{
		Method:   "GET",
		Path:     "/",
		Protocol: "HTTP/1.1",
		Headers: testHeaders{
			"Host":            "localhost:8080",
			"Content-Type":    "some content type",
			"Accept-Encoding": "gzip, deflate, br",
		},
		Body:                 "",
		StrictHeadersCompare: false,
	}

	require.Nil(t, err)
	require.True(t, done, "wanted completion flag")
	require.Empty(t, extra, "unwanted extra")
	require.Nil(t, compareRequests(wantedRequest, <-bodyChan, *request))
}

func TestParserReuseAbility(t *testing.T) {
	parser, request := getParser()

	rawRequest := []byte("GET / HTTP/1.1\r\nContent-Type: some content type\r\nHost: indigo.dev\r\n\r\n")
	bodyChan := make(chan []byte, 1)
	go readBody(request, bodyChan)
	err, extra := FeedParser(parser, rawRequest, 5)
	body := <-bodyChan

	wantedRequest := WantedRequest{
		Method:   "GET",
		Path:     "/",
		Protocol: "HTTP/1.1",
		Headers: testHeaders{
			"Content-Type": "some content type",
			"Host":         "indigo.dev",
		},
		Body:                 "",
		StrictHeadersCompare: true,
	}

	require.Nil(t, err)
	require.Nil(t, compareRequests(wantedRequest, body, *request))
	require.Empty(t, extra, "unwanted extra")

	go readBody(request, bodyChan)
	err, extra = FeedParser(parser, rawRequest, 5)
	body = <-bodyChan

	require.Nil(t, err)
	require.Nil(t, compareRequests(wantedRequest, body, *request))
}

func testOnlyLFGETRequest(t *testing.T, n int) {
	parser, request := getParser()

	rawRequest := []byte("GET / HTTP/1.1\nServer: indigo\n\n")

	if n < 0 {
		n = len(rawRequest)
	}

	wantedRequest := WantedRequest{
		Method:   "GET",
		Path:     "/",
		Protocol: "HTTP/1.1",
		Headers: testHeaders{
			"Server": "indigo",
		},
		Body:                 "",
		StrictHeadersCompare: true,
	}

	bodyChan := make(chan []byte, 1)
	go readBody(request, bodyChan)
	err, extra := FeedParser(parser, rawRequest, n)

	require.Nil(t, err)
	require.Empty(t, extra, "unwanted extra")
	require.Nil(t, compareRequests(wantedRequest, <-bodyChan, *request))
}

func TestOnlyLFGETRequestFull(t *testing.T) {
	testOnlyLFGETRequest(t, -1)
}

func TestOnlyLFGETRequest5Chars(t *testing.T) {
	testOnlyLFGETRequest(t, 5)
}

func TestOnlyLFGETRequest1Char(t *testing.T) {
	testOnlyLFGETRequest(t, 1)
}

func TestConnectionClose(t *testing.T) {
	parser, request := getParser()

	body := "Hello, I have a body for you!"
	rawRequest := []byte("POST / HTTP/1.1\r\nHost: indigo.dev\r\nConnection: close\r\n\r\n" + body)

	wantedRequest := WantedRequest{
		Method:   "POST",
		Path:     "/",
		Protocol: "HTTP/1.1",
		Headers: testHeaders{
			"Host":       "indigo.dev",
			"Connection": "close",
		},
		Body:                 body,
		StrictHeadersCompare: true,
	}

	bodyChan := make(chan []byte, 1)
	go readBody(request, bodyChan)
	err, extra := FeedParser(parser, rawRequest, 5)

	require.Nil(t, err)
	require.Empty(t, extra, "unwanted extra")

	// on Connection: close header, the finish is connection close
	// in this case, reading from socket returns empty byte
	// and this will be a completion mark for our parser
	var done bool
	done, extra, err = parser.Parse(nil)

	require.True(t, done, "wanted completion flag")
	require.Empty(t, extra, "unwanted extra")
	require.Error(t, err, errors.ErrConnectionClosed)
	require.Nil(t, compareRequests(wantedRequest, <-bodyChan, *request))
}
