package types

import (
	"indigo/http/headers"
	methods "indigo/http/method"
	"indigo/http/proto"
	"indigo/http/url"
	"indigo/internal"
)

// Request struct represents http request
// About headers manager see at http/headers/headers.go:Manager
// Headers attribute references at that one that lays in
// manager
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

// NewRequest returns a new instance of request object and body gateway
// Must not be used externally, this function is for internal purposes only
// HTTP/1.1 as a protocol by default is set because if first request from user
// is invalid, we need to render a response using request method, but appears
// that default method is a null-value (proto.Unknown)
// Also url.Query is being constructed right here instead of passing from outside
// because it has only optional purposes and buff will be nil anyway
// But maybe it's better to implement DI all the way we go? I don't know, maybe
// someone will contribute and fix this
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

// OnBody is a proxy-function for r.body.Read. This method reads body in streaming
// processing mode by calling onBody on each body piece, and onComplete when body
// is over (onComplete is guaranteed to be called except situation when body is already
// read)
func (r *Request) OnBody(onBody onBodyCallback, onComplete onCompleteCallback) error {
	return r.body.Read(onBody, onComplete)
}

// Body is a high-level function that wraps OnBody, and the only it does is reading
// pieces of body into the buffer that is a nil by default, but may grow and will stay
// as big as it grew until the disconnect
func (r *Request) Body() ([]byte, error) {
	r.bodyBuff = r.bodyBuff[:0]
	err := r.body.Read(func(b []byte) error {
		r.bodyBuff = append(r.bodyBuff, b...)
		return nil
	}, func(err error) {
		// ignore error here, because it will be anyway returned from r.body.Read call
	})

	return r.bodyBuff, err
}

// Reset resets request object. It is made to clear the object between requests
func (r *Request) Reset() error {
	r.headersManager.Reset()
	r.Headers = r.headersManager.Headers

	return r.body.Reset()
}
