package tests

import (
	"errors"
	"indigo/http/parser"
)

type ParserRetVal struct {
	State parser.RequestState
	Extra []byte
	Err   error
}

/*
HTTPParserMock is pretty simple. It just waits until CallsExpected is exceeded,
after always returns SetDone, SetExtra and ThrowErr (if not nil then SetDone
will be ignored and always set to true)
*/
type HTTPParserMock struct {
	Actions      []ParserRetVal
	callsCounter int

	err error
}

func (h *HTTPParserMock) Parse(_ []byte) (state parser.RequestState, extra []byte, err error) {
	if h.callsCounter >= len(h.Actions) {
		h.err = errors.New("too much calls")
		return parser.Error, nil, h.err
	}

	action := h.Actions[h.callsCounter]
	h.callsCounter++

	return action.State, action.Extra, action.Err
}

func (h *HTTPParserMock) Clear() {
	h.callsCounter = 0
}

func (h HTTPParserMock) CallsCount() int {
	return h.callsCounter
}

func (h HTTPParserMock) GetError() error {
	return h.err
}
