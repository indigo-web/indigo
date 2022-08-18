package types

import (
	"indigo/http/headers"
	methods "indigo/http/method"
	"indigo/http/proto"
	"indigo/http/url"
	"indigo/internal"
)

type Request struct {
	Method   methods.Method
	Path     url.Path
	Query    url.Query
	Fragment url.Fragment
	Proto    proto.Proto

	Headers        headers.Headers
	headersManager *headers.Manager

	body     requestBody
	bodyBuff []byte
}

func NewRequest(manager *headers.Manager) (*Request, *internal.BodyGateway) {
	requestBodyStruct, gateway := newRequestBody()
	request := &Request{
		Query:          url.NewQuery(nil),
		Proto:          proto.HTTP11,
		Headers:        manager.Headers,
		headersManager: manager,
		body:           requestBodyStruct,
	}

	return request, gateway
}

func (r *Request) OnBody(onBody onBodyCallback, onComplete onCompleteCallback) error {
	return r.body.Read(onBody, onComplete)
}

func (r *Request) Body() ([]byte, error) {
	r.bodyBuff = r.bodyBuff[:0]
	err := r.body.Read(func(b []byte) error {
		r.bodyBuff = append(r.bodyBuff, b...)
		return nil
	}, func(err error) {
	})

	return r.bodyBuff, err
}

func (r *Request) Reset() error {
	r.headersManager.Reset()
	r.Headers = r.headersManager.Headers

	return r.body.Reset()
}
