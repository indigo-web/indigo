package webserver

import (
	"io"
)

type RequestCompleted chan bool

type Dispatcher interface {
	ProcessRequest(client Client, completed RequestCompleted)
}

type WriteResponse func(data []byte) error

type Client struct {
	Request       Request
	Response      Response
	WriteResponse WriteResponse
}

type Request struct {
	Method   []byte
	Path     []byte
	Protocol []byte
	Headers  Headers
	Body     io.Reader
}

type Response struct {
	code     int
	codeDesc []byte
	body     []byte
	headers  []Header
}

func (r *Response) WithCode(code int) {
	r.code = code
}

// TODO: add WithXxx for headers and body
// TODO: for headers, I need something like headers.SetOrUpdate function that will
//       update already-existing header or add new
