package http

import (
	"context"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/crypt"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/transport"
	"net"
)

var zeroContext = context.Background()

type Params = *keyvalue.Storage

// Request represents HTTP request
type Request struct {
	// Method represents the request's method
	Method method.Method
	// Path represents decoded request URI
	Path Path
	// Query are request's URI parameters
	Query query.Query
	// Params are dynamic path's wildcards
	Params Params
	// Proto is the protocol, which was used to make the request
	Proto proto.Proto
	// Headers are request headers. They are stored non-normalized, however lookup is
	// case-insensitive
	Headers headers.Headers
	commonHeaders
	// Remote represents remote net.Addr.
	// WARNING: in order to use the value to represent a user, MAKE SURE there are no proxies
	// in the middle
	Remote net.Addr
	// Ctx is a request context. It may be filled with arbitrary data across middlewares
	// and handler by itself
	Ctx context.Context
	// Env is a set of fixed variables passed by core. They are passed separately from Request.Ctx
	// in order to not only distinguish user-defined values in ctx from those from core, but also
	// to gain performance, as accessing the struct is much faster than looking up in context.Context
	Env Environment
	// Body accesses the request's body
	Body     *Body
	client   transport.Client
	hijacked bool
	response *Response
	jar      cookie.Jar
	cfg      *config.Config
}

// NewRequest returns a new instance of request object and body gateway
// Must not be used externally, this function is for internal purposes only
// HTTP/1.1 as a protocol by default is set because if first request from user
// is invalid, we need to render a response using request method, but appears
// that default method is a null-value (proto.Unknown)
func NewRequest(
	cfg *config.Config, hdrs headers.Headers, query query.Query, response *Response,
	client transport.Client, params Params,
) *Request {
	return &Request{
		Query:    query,
		Params:   params,
		Proto:    proto.HTTP11,
		Headers:  hdrs,
		Remote:   client.Remote(),
		Ctx:      zeroContext,
		client:   client,
		response: response,
		cfg:      cfg,
	}
}

// Cookies returns a cookie jar with parsed cookies key-value pairs, and an error
// if the syntax is malformed. The returned jar should be re-used, as this method
// doesn't cache the parsed result across calls and may be pretty expensive
func (r *Request) Cookies() (cookie.Jar, error) {
	if r.jar == nil {
		r.jar = cookie.NewJarPreAlloc(r.cfg.Headers.CookiesPreAllocate)
	}

	r.jar.Clear()

	// in RFC 6265, 5.4 cookies are explicitly prohibited from being split into
	// list, yet in HTTP/2 it's allowed. I have concerns of some user-agents may
	// despite sending them as a list, even via HTTP/1.1
	for _, value := range r.Headers.Values("cookie") {
		if err := cookie.Parse(r.jar, value); err != nil {
			return nil, err
		}
	}

	return r.jar, nil
}

// Respond returns Response object.
//
// WARNING: this method clears the response builder under the hood. As it is passed
// by reference, it'll be cleared EVERYWHERE along a handler
func (r *Request) Respond() *Response {
	return r.response.Clear()
}

// Hijack the connection. Request body will be implicitly read (so if you need it you
// should read it before) to the end. After handler exits, the connection will
// be closed, so the connection can be hijacked at most once
func (r *Request) Hijack() (transport.Client, error) {
	if err := r.Body.Discard(); err != nil {
		return nil, err
	}

	r.hijacked = true

	return r.client, nil
}

// Hijacked tells whether the connection was hijacked or not
func (r *Request) Hijacked() bool {
	return r.hijacked
}

// Reset clears request headers and reads body into nowhere until completed.
// It is implemented to clear the request object between requests
func (r *Request) Reset() (err error) {
	r.Query.Reset()
	r.Params.Clear()
	r.Headers.Clear()
	r.commonHeaders = commonHeaders{}
	r.Ctx = zeroContext
	r.Env = Environment{}

	return r.Body.Reset()
}

type Environment struct {
	// Error contains an error, if occurred
	Error error
	// AllowedMethods is used to pass a string containing all the allowed methods for a
	// specific endpoint. Has non-zero-value only when 405 Method Not Allowed raises
	AllowedMethods string
	// Encryption is a token that corresponds to the used encryption method. May be
	// extended by custom values
	Encryption crypt.Encryption
	// AliasFrom contains the original request path, in case it was replaced via alias
	// aka implicit redirect
	AliasFrom string
}

type commonHeaders struct {
	// Encoding holds an information about encoding, that was used to make the request
	Encoding Encodings
	// ContentLength obtains the value from Content-Length header. It holds the value of 0
	// if isn't presented.
	//
	// NOTE: you shouldn't rely on this value, as it may be anything (mostly 0) if any
	// Transfer-Encoding were applied.
	ContentLength int
	// ContentType obtains Content-Type header value
	ContentType string
	// Connection holds the Connection header value. It isn't normalized, so can be anything
	// and in any case. So in order to compare it, highly recommended to do it case-insensibly
	Connection string
	// Upgrade is the protocol token, which is set by default to proto.Unknown. In
	// case it is anything else, then Upgrade header was received
	Upgrade proto.Proto
}

type Encodings struct {
	// Transfer contains all applied Transfer-Encoding codings in their original order, except
	// the chunked. Chunked Transfer Encoding has its own boolean flag.
	Transfer []string
	// Content contains all applied Content-Encoding codings in their original order.
	Content []string
	// Chunked doesn't belong to any of encodings, as it is still must be processed individually
	Chunked, HasTrailer bool
}
