package client

import (
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"io"
)

type Request struct {
	Method  method.Method
	Path    string
	Query   Query
	Proto   proto.Proto
	Headers *headers.Headers
	Body    io.Reader
}

func NewRequest() *Request {
	return &Request{
		Query:   make(Query),
		Headers: headers.NewHeaders(),
	}
}

func (r *Request) WithMethod(m method.Method) *Request {
	r.Method = m
	return r
}

func (r *Request) WithPath() *Request {

	return r
}
