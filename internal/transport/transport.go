package transport

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/proto"
)

type Parser interface {
	Parse(b []byte) (state RequestState, extra []byte, err error)
}

// RequestState represents the state of the request's parsing
type RequestState uint8

const (
	Pending RequestState = iota + 1
	HeadersCompleted
	Error
)

type Writer interface {
	Write([]byte) error
}

// Serializer converts an HTTP response builder into bytes and writes it
type Serializer interface {
	PreWrite(target proto.Proto, response *http.Response)
	Write(target proto.Proto, request *http.Request, response *http.Response, writer Writer) error
}

// Transport is a general pair of a parser and a dumper. Usually consists of both belonging
// to a same protocol major version
type Transport interface {
	Parser
	Serializer
}
