package http

import (
	"context"
	"net"

	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/transport"
)

var zeroContext = context.Background()

type (
	Headers = *kv.Storage
	Header  = kv.Pair
	Params  = *kv.Storage
	Vars    = *kv.Storage
)

// Request is the HTTP request representation.
type Request struct {
	// Method is an enum representing the request method.
	Method method.Method
	// Path is a decoded and validated string, guaranteed to hold ASCII-printable characters only.
	Path string
	// Params are request URI parameters.
	Params Params
	// Vars are dynamic routing segments.
	Vars Vars
	// Proto is the enum of a protocol used for the request. Can be changed (mostly through upgrade).
	Protocol proto.Protocol
	// Headers holds non-normalized header pairs, even though lookup is case-insensitive. Header keys
	// and values aren't validated, therefore may contain ASCII-nonprintable and/or Unicode characters.
	Headers Headers
	commonHeaders
	// Remote holds the remote address. Please note that this is generally not a good parameter to identify
	// a user, because there might be proxies in the middle.
	Remote net.Addr
	// Ctx is user-managed context which lives as long as the connection does and is never automatically
	// cleared.
	Ctx context.Context
	// Env contains a fixed set of contextual values which are useful in specific cases. They aren't
	// passed via the Ctx due to performance considerations.
	Env Environment
	// Body is a dedicated entity providing access to the message body.
	Body     *Body
	client   transport.Client
	hijacked bool
	response *Response
	jar      cookie.Jar
	cfg      *config.Config
}

func NewRequest(
	cfg *config.Config,
	response *Response,
	client transport.Client,
	headers, params, vars *kv.Storage,
) *Request {
	return &Request{
		Protocol: proto.HTTP11,
		Params:   params,
		Vars:     vars,
		Headers:  headers,
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
		r.jar = cookie.NewJarPreAlloc(r.cfg.Headers.CookiesPrealloc)
	}

	r.jar.Clear()

	// even though RFC 6265, 5.4 prohibits the Cookie header from being split into a list,
	// some user-agents still might do so in order to fit each value into the 8K limit.
	for value := range r.Headers.Values("cookie") {
		if err := cookie.Parse(r.jar, value); err != nil {
			return nil, err
		}
	}

	return r.jar, nil
}

// Respond returns Response object.
//
// WARNING: the Response is cleared before being returned. Considering it is stored by pointer,
// this action might affect your application if it was stored anywhere before.
func (r *Request) Respond() *Response {
	return r.response.Clear()
}

// Hijack hijacks an underlying connection. The request body is implicitly discarded before
// exposing the transport. After the handler function terminates, the connection is closed automatically.
func (r *Request) Hijack() (transport.Client, error) {
	if err := r.Body.Discard(); err != nil {
		return nil, err
	}

	r.hijacked = true

	return r.client, nil
}

// Hijacked tells whether the connection was hijacked.
func (r *Request) Hijacked() bool {
	return r.hijacked
}

// Reset clears all the request fields to its zero values. This also includes resetting the context.
func (r *Request) Reset() {
	r.Params.Clear()
	r.Vars.Clear()
	r.Headers.Clear()
	r.commonHeaders = commonHeaders{}
	r.Ctx = zeroContext
	r.Env = Environment{}
}

type Environment struct {
	// Error contains an error, if occurred
	Error error
	// AllowedMethods is used to pass a string containing all the allowed methods for a
	// specific endpoint. Has non-zero-value only when 405 Method Not Allowed raises
	AllowedMethods string
	// Encryption represents the cryptographic protocol on top of the connection. They're
	// comparable against the tls.Version... enums. Zero value means no encryption.
	Encryption uint16
	// AliasFrom contains the original request path, in case it was replaced via alias
	// aka implicit redirect
	AliasFrom string
}

type commonHeaders struct {
	// Encoding holds an information about encoding, that was used to make the request
	Encoding Encodings
	// ContentLength holds the Content-Length header value. It isn't recommended to rely solely on this
	// value, as it can be whatever (but most likely zero) if Request.Encoding.Chunked is true.
	ContentLength int
	// ContentType holds the Content-Type header value.
	ContentType string
	// Connection holds the Connection header value. It isn't normalized, so can be anything
	// and in any case. So in order to compare it, highly recommended to do it case-insensibly
	Connection string
	// Upgrade holds the `Upgrade` header value. It's set to anything but proto.Unknown only when
	// an upgrade request is received.
	Upgrade proto.Protocol
}

type Encodings struct {
	// Accept is the list of accepted by client tokens.
	Accept []string
	// Transfer is the list of tokens used for this request from `Transfer-Encoding` header value.
	Transfer []string
	// Content is the list of tokens used for this request from `Content-Encoding` header value.
	Content []string
	// Chunked describes whether the Transfer attribute is not empty and ends with the `chunked`
	// encoding.
	Chunked bool
}
