package http

import (
	"context"
	"github.com/indigo-web/indigo/internal/unreader"
	"io"
	"net"

	"github.com/indigo-web/indigo/http/headers"
	// I don't know why, but otherwise GoLand cries about unused import, even if it's used
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
)

type (
	Params = map[string]string

	Path struct {
		String   string
		Params   Params
		Query    query.Query
		Fragment Fragment
	}

	Fragment = string
)

// Request struct represents http request
// About headers manager see at http/headers/headers.go:Manager
// Headers attribute references at that one that lays in manager
type Request struct {
	body             *Body
	conn             net.Conn
	Remote           net.Addr
	Ctx              context.Context
	Headers          *headers.Headers
	response         Response
	Path             Path
	ContentLength    int
	TransferEncoding headers.TransferEncoding
	Method           method.Method
	Upgrade          proto.Proto
	Proto            proto.Proto
	wasHijacked      bool
	clearParamsMap   bool
}

// NewRequest returns a new instance of request object and body gateway
// Must not be used externally, this function is for internal purposes only
// HTTP/1.1 as a protocol by default is set because if first request from user
// is invalid, we need to render a response using request method, but appears
// that default method is a null-value (proto.Unknown)
func NewRequest(
	hdrs *headers.Headers, query query.Query, response Response, conn net.Conn, body *Body,
	paramsMap Params, disableParamsMapClearing bool,
) *Request {
	request := &Request{
		Path: Path{
			Params: paramsMap,
			Query:  query,
		},
		Proto:          proto.HTTP11,
		Headers:        hdrs,
		Remote:         conn.RemoteAddr(),
		conn:           conn,
		body:           body,
		Ctx:            context.Background(),
		response:       response,
		clearParamsMap: !disableParamsMapClearing,
	}

	return request
}

// Body returns an entity representing a request's body
func (r *Request) Body() *Body {
	return r.body
}

// Hijack the connection. Request body will be implicitly read (so if you need it you
// should read it before) all the body left. After handler exits, the connection will
// be closed, so the connection can be hijacked only once
func (r *Request) Hijack() (net.Conn, error) {
	if err := r.body.Reset(); err != nil {
		return nil, err
	}

	r.wasHijacked = true

	return r.conn, nil
}

// WasHijacked returns true or false, depending on whether was a connection hijacked
func (r *Request) WasHijacked() bool {
	return r.wasHijacked
}

// Clear resets request headers and reads body into nowhere until completed.
// It is implemented to clear the request object between requests
func (r *Request) Clear() (err error) {
	r.Path.Fragment = ""
	r.Path.Query.Set(nil)
	r.Ctx = context.Background()
	r.response = r.response.Clear()

	if err = r.body.Reset(); err != nil {
		return err
	}

	r.ContentLength = 0
	r.TransferEncoding = headers.TransferEncoding{}
	r.Upgrade = proto.Unknown

	if r.clearParamsMap && len(r.Path.Params) > 0 {
		for k := range r.Path.Params {
			delete(r.Path.Params, k)
		}
	}

	return nil
}

// RespondTo returns a response object of request
func RespondTo(request *Request) Response {
	return request.response
}

// bodyIOReader is an implementation of io.Reader for request body
type bodyIOReader struct {
	unreader *unreader.Unreader
	reader   BodyReader
}

func newBodyIOReader() bodyIOReader {
	return bodyIOReader{
		unreader: new(unreader.Unreader),
	}
}

func (b bodyIOReader) Read(buff []byte) (n int, err error) {
	data, err := b.unreader.PendingOr(b.reader.Read)
	copy(buff, data)
	n = len(data)

	if len(buff) < len(data) {
		b.unreader.Unread(data[len(buff):])
		n = len(buff)
	}

	return n, err
}

func (b bodyIOReader) WriteTo(w io.Writer) (n int64, err error) {
	for {
		data, err := b.reader.Read()
		switch err {
		case nil:
		case io.EOF:
			return n, nil
		default:
			return 0, err
		}

		n1, err := w.Write(data)
		n += int64(n1)
	}
}

func (b bodyIOReader) Reset() {
	b.unreader.Unread(nil)
}
