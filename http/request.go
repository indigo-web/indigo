package http

import (
	"context"
	"io"
	"net"

	"github.com/fakefloordiv/indigo/http/headers"
	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/http/url"
)

type (
	onBodyCallback     func([]byte) error
	onCompleteCallback func() error
	BodyReader         interface {
		Init(*Request)
		Read() ([]byte, error)
	}
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

	ContentLength int
	ChunkedTE     bool

	body     BodyReader
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
	hdrs headers.Headers, query url.Query, response Response, remote net.Addr, body BodyReader,
) *Request {
	request := &Request{
		Query:    query,
		Proto:    proto.HTTP11,
		Headers:  hdrs,
		Remote:   remote,
		body:     body,
		Ctx:      context.Background(),
		response: response,
	}

	return request
}

// OnBody is a low-level interface accessing a request body. It takes onBody callback that is
// being called every time a piece of body is read (note: even a single byte can be passed).
// In case error returned, it'll be returned from OnBody method. In case onBody never did return
// an error, onComplete will be called when the body will be finished. This callback also can
// return an error that'll be returned from OnBody method - for example, in case body's hash sum
// is invalid
func (r *Request) OnBody(onBody onBodyCallback, onComplete onCompleteCallback) error {
	for {
		piece, err := r.body.Read()
		switch err {
		case nil:
		case io.EOF:
			return onComplete()
		default:
			return err
		}

		if err = onBody(piece); err != nil {
			return err
		}
	}
}

// Body is a high-level function that wraps OnBody, and the only it does is reading
// pieces of body into the buffer that is a nil by default, but may grow and will stay
// as big as it grew until the disconnect
func (r *Request) Body() ([]byte, error) {
	if !r.HasBody() {
		return nil, nil
	}

	if r.bodyBuff == nil {
		r.bodyBuff = make([]byte, r.ContentLength)
	}

	r.bodyBuff = r.bodyBuff[:0]

	err := r.OnBody(func(b []byte) error {
		r.bodyBuff = append(r.bodyBuff, b...)
		return nil
	}, func() error {
		return nil
	})

	return r.bodyBuff, err
}

// Reader returns io.Reader for request body. This method may be called multiple times,
// but reading from multiple readers leads to Undefined Behaviour
func (r *Request) Reader() io.Reader {
	return newBodyIOReader(r.body)
}

func (r Request) HasBody() bool {
	return r.ContentLength > 0 || r.ChunkedTE
}

// Reset resets request headers and reads body into nowhere until completed.
// It is implemented to clear the request object between requests
func (r *Request) Reset() (err error) {
	r.Fragment = ""
	r.Query.Set(nil)
	r.Ctx = context.Background()
	r.response = r.response.Reset()

	if err = r.resetBody(); err != nil {
		return err
	}

	r.ContentLength = 0
	r.ChunkedTE = false

	return nil
}

// resetBody just reads the body until its end
func (r *Request) resetBody() error {
	return r.OnBody(func([]byte) error {
		return nil
	}, func() error {
		return nil
	})
}

func RespondTo(request *Request) Response {
	return request.response
}
