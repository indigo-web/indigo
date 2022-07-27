package tests

import (
	"errors"
	"indigo/types"
)

type RouterRetVal struct {
	Err error
}

type RouterMock struct {
	Actions []RouterRetVal

	onRequestCalls, onErrorCalls int
	err                          error
}

func (r *RouterMock) OnRequest(_ *types.Request, _ types.ResponseWriter) error {
	if r.onErrorCalls >= len(r.Actions) {
		r.err = errors.New("too much calls")
		return r.Actions[len(r.Actions)-1].Err
	}

	action := r.Actions[r.onRequestCalls]
	r.onRequestCalls++

	return action.Err
}

func (r *RouterMock) OnError(_ error) {
	r.onErrorCalls++
}

func (r RouterMock) GetError() error {
	return r.err
}
