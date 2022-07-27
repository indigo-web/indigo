package server

import (
	"indigo/tests"
	"indigo/types"
	"testing"
)

var (
	simpleRequest = []byte("GET / HTTP/1.1\r\nHello: world\r\n\r\n")
	nilRequest    = (*types.Request)(nil)
)

func getPollerOutput(reqChan <-chan *types.Request, errChan <-chan error) (*types.Request, error) {
	select {
	case req := <-reqChan:
		return req, nil
	case err := <-errChan:
		return nil, err
	}
}

func TestHTTPServerRunAbility(t *testing.T) {
	t.Run("RunOnce", func(t *testing.T) {
		mockedParser := tests.HTTPParserMock{
			Actions: []tests.ParserRetVal{
				{true, nil, nil},
			},
		}

		reqChan, errChan := make(requestsChan, 1), make(errorsChan, 2)
		errChan <- nil
		handler := newHTTPHandler(HTTPHandlerArgs{
			Router:     nil,
			Request:    nilRequest,
			Parser:     &mockedParser,
			RespWriter: nil,
		}, reqChan, errChan)

		if err := handler.OnData(simpleRequest); err != nil {
			t.Fatalf("unwanted error: %s", err)
		} else if callsCount := mockedParser.CallsCount(); callsCount != 1 {
			t.Fatalf("wanted exactly 1 call, got %d", callsCount)
		} else if err = mockedParser.GetError(); err != nil {
			t.Fatalf("unwanted error from mocked parser: %s", err)
		}

		req, reqErr := getPollerOutput(reqChan, errChan)
		if reqErr != nil {
			t.Fatalf("unwanted error in errChan: %s", reqErr)
		} else if req != nilRequest {
			t.Fatalf("mismatching wanted and got request objects")
		}
	})

	t.Run("SplitRequestInto2Parts", func(t *testing.T) {
		firstPart := simpleRequest[:len(simpleRequest)/2]
		secondPart := simpleRequest[len(simpleRequest)/2:]

		mockedParser := tests.HTTPParserMock{
			Actions: []tests.ParserRetVal{
				{false, nil, nil},
				{true, nil, nil},
			},
		}

		reqChan, errChan := make(requestsChan, 1), make(errorsChan, 2)
		errChan <- nil
		handler := newHTTPHandler(HTTPHandlerArgs{
			Router:     nil,
			Request:    nilRequest,
			Parser:     &mockedParser,
			RespWriter: nil,
		}, reqChan, errChan)

		if err := handler.OnData(firstPart); err != nil {
			t.Fatalf("unwanted error: %s", err)
		} else if callsCount := mockedParser.CallsCount(); callsCount != 1 {
			t.Fatalf("wanted exactly 1 call, got %d", callsCount)
		} else if err = mockedParser.GetError(); err != nil {
			t.Fatalf("unwanted error from mocked parser: %s", err)
		}

		if err := handler.OnData(secondPart); err != nil {
			t.Fatalf("unwanted error: %s", err)
		} else if callsCount := mockedParser.CallsCount(); callsCount != 2 {
			t.Fatalf("wanted already 2 calls, got %d", callsCount)
		} else if err = mockedParser.GetError(); err != nil {
			t.Fatalf("unwanted error from mocked parser: %s", err)
		}

		req, reqErr := getPollerOutput(reqChan, errChan)
		if reqErr != nil {
			t.Fatalf("unwanted error in errChan: %s", reqErr)
		} else if req != nilRequest {
			t.Fatalf("mismatching wanted and got request objects")
		}
	})

}

func TestHTTPServer2Requests(t *testing.T) {
	t.Run("2Requests", func(t *testing.T) {
		mockedParser := tests.HTTPParserMock{
			Actions: []tests.ParserRetVal{
				{true, nil, nil},
				{true, nil, nil},
			},
		}

		reqChan, errChan := make(requestsChan, 1), make(errorsChan, 2)
		errChan <- nil
		handler := newHTTPHandler(HTTPHandlerArgs{
			Router:     nil,
			Request:    nilRequest,
			Parser:     &mockedParser,
			RespWriter: nil,
		}, reqChan, errChan)

		err := handler.OnData(simpleRequest)
		if err != nil {
			t.Fatalf("unwanted error: %s", err)
		} else if callsCount := mockedParser.CallsCount(); callsCount != 1 {
			t.Fatalf("wanted exactly 1 call, got %d", callsCount)
		} else if err = mockedParser.GetError(); err != nil {
			t.Fatalf("unwanted error from mocked parser: %s", err)
		}

		req, reqErr := getPollerOutput(reqChan, errChan)
		if reqErr != nil {
			t.Fatalf("unwanted error in errChan: %s", reqErr)
		} else if req != nilRequest {
			t.Fatalf("mismatching wanted and got request objects")
		}

		errChan <- nil
		err = handler.OnData(simpleRequest)
		if err != nil {
			t.Fatalf("unwanted error: %s", err)
		} else if callsCount := mockedParser.CallsCount(); callsCount != 2 {
			t.Fatalf("wanted already 2 calls, got %d", callsCount)
		} else if err = mockedParser.GetError(); err != nil {
			t.Fatalf("unwanted error from mocked parser: %s", err)
		}

		req, reqErr = getPollerOutput(reqChan, errChan)
		if reqErr != nil {
			t.Fatalf("unwanted error in errChan: %s", reqErr)
		} else if req != nilRequest {
			t.Fatalf("mismatching wanted and got request objects")
		}
	})

	t.Run("2RequestsWithExtra", func(t *testing.T) {
		// just copy to be sure no implicit shit will happen
		request := append(make([]byte, 0, len(simpleRequest)), simpleRequest...)
		firstRequest := append(request, simpleRequest[:len(simpleRequest)/2]...)
		secondRequest := simpleRequest[len(simpleRequest)/2:]

		mockedParser := tests.HTTPParserMock{
			Actions: []tests.ParserRetVal{
				{true, nil, nil},
				{true, nil, nil},
			},
		}

		reqChan, errChan := make(requestsChan, 1), make(errorsChan, 2)
		errChan <- nil
		handler := newHTTPHandler(HTTPHandlerArgs{
			Router:     nil,
			Request:    nilRequest,
			Parser:     &mockedParser,
			RespWriter: nil,
		}, reqChan, errChan)

		err := handler.OnData(firstRequest)
		if err != nil {
			t.Fatalf("unwanted error: %s", err)
		} else if callsCount := mockedParser.CallsCount(); callsCount != 1 {
			t.Fatalf("wanted exactly 1 call, got %d", callsCount)
		} else if err = mockedParser.GetError(); err != nil {
			t.Fatalf("unwanted error from mocked parser: %s", err)
		}

		req, reqErr := getPollerOutput(reqChan, errChan)
		if reqErr != nil {
			t.Fatalf("unwanted error in errChan: %s", reqErr)
		} else if req != nilRequest {
			t.Fatalf("mismatching wanted and got request objects")
		}

		errChan <- nil
		err = handler.OnData(secondRequest)
		if err != nil {
			t.Fatalf("unwanted error: %s", err)
		} else if callsCount := mockedParser.CallsCount(); callsCount != 2 {
			t.Fatalf("wanted already 2 calls, got %d", callsCount)
		} else if err = mockedParser.GetError(); err != nil {
			t.Fatalf("unwanted error from mocked parser: %s", err)
		}

		req, reqErr = getPollerOutput(reqChan, errChan)
		if reqErr != nil {
			t.Fatalf("unwanted error in errChan: %s", reqErr)
		} else if req != nilRequest {
			t.Fatalf("mismatching wanted and got request objects")
		}
	})
}
