package types

import (
	"indigo/http"
	"indigo/internal"
)

type Params map[string][]byte

type Request struct {
	Method   http.Method
	Path     []byte
	Params   Params
	Protocol http.Protocol
	Headers  http.Headers

	body         requestBody
	bodyBuff     []byte
	bodyBuffSize uint32
}

func NewRequest(
	pathBuffer []byte, headers http.Headers, params Params,
	bodyBuffSize uint32) (Request, *internal.Pipe) {
	// pipe is sized chan because parser can write an error even before
	// handler will be called
	pipe := internal.NewChanSizedPipe(0, 1)

	return Request{
		Path:     pathBuffer,
		Params:   params,
		Protocol: http.Protocol{},
		Headers:  headers,
		body: requestBody{
			body: pipe,
		},
		bodyBuffSize: bodyBuffSize,
	}, pipe
}

func (r *Request) GetBody(bodyCb onBodyCallback, completeCb onBodyCompleteCallback) error {
	return r.body.Read(bodyCb, completeCb)
}

func (r *Request) GetFullBody() (body []byte, err error) {
	if uint32(cap(r.bodyBuff)) < r.bodyBuffSize {
		r.bodyBuff = make([]byte, 0, r.bodyBuffSize)
	} else {
		r.bodyBuff = r.bodyBuff[:0]
	}

	err = r.GetBody(
		func(b []byte) error {
			r.bodyBuff = append(r.bodyBuff, b...)
			return nil
		},
		func(bodyErr error) {
			return
		},
	)

	return r.bodyBuff, err
}

func (r *Request) Reset() {
	// TODO: headers must also be reset
	r.body.Reset()
}
