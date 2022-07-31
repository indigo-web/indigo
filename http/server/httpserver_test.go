package server

import (
	"indigo/http/parser"
	"indigo/tests"
	"indigo/types"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	simpleGETRequest      = []byte("GET / HTTP/1.1\r\nHello: world\r\n\r\n")
	helloWorldPOSTRequest = []byte("POST / HTTP/1.1\r\nContent-Length: 13\r\n\r\nHello, world!")
	nilRequest            = (*types.Request)(nil)
)

func getPollerOutput(reqChan <-chan *types.Request, errChan <-chan error) (*types.Request, error) {
	select {
	case req := <-reqChan:
		return req, nil
	case err := <-errChan:
		return nil, err
	}
}

func testParseNRequests(t *testing.T, n int, request []byte) {
	reqChan, errChan := make(requestsChan, 1), make(errorsChan, 2)
	mockedParser := &tests.HTTPParserMock{}
	handler := newHTTPHandler(HTTPHandlerArgs{
		Router:     nil,
		Request:    nilRequest,
		Parser:     mockedParser,
		RespWriter: nil,
	}, reqChan, errChan)

	for i := 0; i < n; i++ {
		mockedParser.Actions = append(mockedParser.Actions, tests.ParserRetVal{
			State: parser.RequestCompleted | parser.BodyCompleted,
			Extra: nil,
			Err:   nil,
		})

		errChan <- nil
		err := handler.OnData(request)
		require.Nil(t, err, "unwanted error")
		require.Equal(t, i+1, mockedParser.CallsCount(), "too much parser calls")
		require.Nil(t, mockedParser.GetError(), "unwanted error")

		req, reqErr := getPollerOutput(reqChan, errChan)
		require.Nil(t, reqErr, "unwanted error")
		require.Equal(t, req, nilRequest)
	}
}

func testParse2Parts(t *testing.T, mockedParser tests.HTTPParserMock, firstPart, secondPart []byte) {
	reqChan, errChan := make(requestsChan, 1), make(errorsChan, 2)
	errChan <- nil
	handler := newHTTPHandler(HTTPHandlerArgs{
		Router:     nil,
		Request:    nilRequest,
		Parser:     &mockedParser,
		RespWriter: nil,
	}, reqChan, errChan)

	err := handler.OnData(firstPart)
	require.Nil(t, err, "unwanted error")
	require.Nil(t, mockedParser.GetError(), "unwanted error")

	req, reqErr := getPollerOutput(reqChan, errChan)
	require.Nil(t, reqErr, "unwanted error")
	require.Equal(t, req, nilRequest)

	errChan <- nil
	err = handler.OnData(secondPart)
	require.Nil(t, err, "unwanted error")
	require.Nil(t, mockedParser.GetError(), "unwanted error")

	req, reqErr = getPollerOutput(reqChan, errChan)
	require.Nil(t, reqErr, "unwanted error")
	require.Equal(t, req, nilRequest)
}

func testSimpleCase(t *testing.T, request []byte) {
	t.Run("RunOnce", func(t *testing.T) {
		testParseNRequests(t, 1, request)
	})

	t.Run("SplitRequestInto2Parts", func(t *testing.T) {
		firstPart := request[:len(request)/2]
		secondPart := request[len(request)/2:]

		mockedParser := tests.HTTPParserMock{
			Actions: []tests.ParserRetVal{
				{parser.Pending, nil, nil},
				{parser.RequestCompleted | parser.BodyCompleted, nil, nil},
			},
		}

		testParse2Parts(t, mockedParser, firstPart, secondPart)
	})
}

func testSimpleManyRequests(t *testing.T, request []byte) {
	t.Run("2Requests", func(t *testing.T) {
		testParseNRequests(t, 2, request)
	})

	t.Run("2RequestsWithExtra", func(t *testing.T) {
		// just copy to be sure no implicit shit will happen
		requestCopy := append(make([]byte, 0, len(request)), request...)
		firstRequest := append(requestCopy, request[:len(request)/2]...)
		secondRequest := request[len(request)/2:]

		mockedParser := tests.HTTPParserMock{
			Actions: []tests.ParserRetVal{
				{parser.RequestCompleted | parser.BodyCompleted, request[:len(request)/2], nil},
				{parser.Pending, nil, nil},
				{parser.RequestCompleted | parser.BodyCompleted, nil, nil},
			},
		}

		testParse2Parts(t, mockedParser, firstRequest, secondRequest)
	})

	t.Run("5Requests", func(t *testing.T) {
		testParseNRequests(t, 5, request)
	})
}

func TestHTTPServerGETRequest(t *testing.T) {
	testSimpleCase(t, simpleGETRequest)
}

func TestHTTPServerPOSTRequest(t *testing.T) {
	testSimpleCase(t, helloWorldPOSTRequest)
}

func TestHTTPServerManyGETRequests(t *testing.T) {
	testSimpleManyRequests(t, simpleGETRequest)
}

func TestHTTPServerManyPOSTRequests(t *testing.T) {
	testSimpleManyRequests(t, helloWorldPOSTRequest)
}
