package http

import (
	"context"
	"github.com/fakefloordiv/indigo/internal/body"
	"io"
	"net"

	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/http/url"
)

type (
	// ConnectionHijacker is for user. It returns error because it has to
	// read full request body to stop the server in defined state. And,
	// as we know, reading body may return an error
	ConnectionHijacker func() (net.Conn, error)

	// hijackConn is like an interface of httpServer method that notifies
	// core about hijacking and returns connection object
	hijackConn func() net.Conn

	Path     = string
	Fragment = string
)

// Request struct represents http request
// About headers manager see at http/headers/headers.go:Manager
// Headers attribute references at that one that lays in manager
type Request struct {
	Method   methods.Method
	Path     Path
	Query    url.Query
	Fragment Fragment
	Proto    proto.Proto
	Remote   net.Addr

	Headers headers.Headers

	ContentLength uint
	ChunkedTE     bool

	body     *requestBody
	bodyBuff []byte

	Ctx      context.Context
	response Response
	Hijack   ConnectionHijacker
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
func NewRequest(
	hdrs headers.Headers, query url.Query, remote net.Addr, ctx context.Context,
	response Response,
) (*Request, *body.Gateway) {
	requestBodyStruct, gateway := newRequestBody()
	request := &Request{
		Query:    query,
		Proto:    proto.HTTP11,
		Headers:  hdrs,
		Remote:   remote,
		body:     requestBodyStruct,
		Ctx:      ctx,
		response: response,
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
	if r.ContentLength == 0 && !r.ChunkedTE {
		return r.bodyBuff[:0], nil
	}

	if r.bodyBuff == nil {
		r.bodyBuff = make([]byte, r.ContentLength)
	}

	r.bodyBuff = r.bodyBuff[:0]

	err := r.body.Read(func(b []byte) error {
		r.bodyBuff = append(r.bodyBuff, b...)
		return nil
	}, func(err error) {
		// ignore error here, because it will be anyway returned from r.body.Read call
	})

	return r.bodyBuff, err
}

// Reader returns io.Reader for request body. This method may be called multiple times,
// but reading from multiple readers leads to Undefined Behaviour
func (r *Request) Reader() io.Reader {
	return newBodyReader(r.body)
}

// Reset resets request headers and reads body into nowhere until completed.
// It is implemented to clear the request object between requests
func (r *Request) Reset() (err error) {
	r.Fragment = ""
	r.Query.Set(nil)
	r.Ctx = context.Background()
	r.response.Reset()

	if err = r.resetBody(); err != nil {
		return err
	}

	r.ContentLength = 0
	r.ChunkedTE = false

	return nil
}

func (r *Request) resetBody() error {
	if r.ContentLength == 0 && !r.ChunkedTE {
		// in case request does not contain a body, it makes no sense to wait
		// for the only nil from channel. This avoids some useless goroutines
		// switches
		r.body.Unread()

		return nil
	}

	return r.body.Reset()
}

// Hijacker is a layer between request object and server that guarantees that request's body
// will be completely read before connection is hijacked. Request body must be read to avoid
// non-determination caused by possible receiving the request body instead of actual data
// that is expected to be received.
//
// This function is exported only because another package (internal/server/http.go) requires
// it. It cannot be inlined there because it uses non-exported method of the request object.
// That is why user is supposed to not care about this function, moreover it receives a hijacker
// function that is not used by user anywhere (except Request.Hijack method that IS NOT EXPECTED
// TO BE PASSED HERE)
func Hijacker(request *Request, hijacker hijackConn) func() (net.Conn, error) {
	return func() (net.Conn, error) {
		// we anyway don't need to have a body anymore. Also, without reading
		// the body until complete server will not transfer into the state
		// we need so this step is anyway compulsory
		switch err := request.resetBody(); err {
		case nil, ErrRead:
		default:
			return nil, err
		}

		return hijacker(), nil
	}
}

func Respond(request *Request) Response {
	return request.response
}
