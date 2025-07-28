package http1

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/buffer"
	"github.com/indigo-web/indigo/internal/codecutil"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/internal/strutil"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/transport"
	"maps"
)

type Suit struct {
	*Parser
	*body
	*serializer
	router router.Router
	client transport.Client
	codecs codecutil.Cache
}

func newSuit(
	cfg *config.Config,
	r router.Router,
	request *http.Request,
	client transport.Client,
	body *body,
	codecs codecutil.Cache,
	headersBuff, statusBuff *buffer.Buffer,
	respBuff []byte,
) *Suit {
	defHeaders := maps.Clone(cfg.Headers.Default)
	defHeaders["Accept-Encoding"] = codecs.AcceptEncodings()

	return &Suit{
		Parser:     NewParser(cfg, request, headersBuff, statusBuff),
		body:       body,
		serializer: newSerializer(cfg, request, client, codecs, respBuff, defHeaders),
		router:     r,
		client:     client,
		codecs:     codecs,
	}
}

// New instantiates an HTTP/1 protocol suit.
func New(
	cfg *config.Config,
	r router.Router,
	client transport.Client,
	request *http.Request,
	codecs codecutil.Cache,
) *Suit {
	headersBuff, statusBuff := construct.Buffers(cfg)
	respBuff := make([]byte, 0, cfg.HTTP.ResponseBuffer.Default)
	b := newBody(client, cfg.Body)

	return newSuit(cfg, r, request, client, b, codecs, headersBuff, statusBuff, respBuff)
}

func (s *Suit) ServeOnce() (ok bool) {
	return s.serve(true)
}

func (s *Suit) Serve() {
	s.serve(false)
}

func (s *Suit) serve(once bool) (ok bool) {
	client := s.client
	request := s.Parser.request

	for {
		data, err := client.Read()
		if err != nil {
			// read-error most probably means deadline exceeding. Just notify the user in
			// this case and return.
			s.router.OnError(request, status.ErrCloseConnection)
			return false
		}

		done, extra, err := s.Parse(data)
		if err != nil {
			resp := respond(request, s.router.OnError(request, err))
			_ = s.Write(request.Protocol, resp)
			return false
		}

		if !done {
			if once {
				return true
			}

			continue
		}

		client.Pushback(extra)
		request.Body.Reset(request)
		s.body.Reset(request)

		transferEncoding := request.Encoding.Transfer
		if !validateTransferEncodingTokens(transferEncoding) {
			resp := respond(request, s.router.OnError(request, status.ErrUnsupportedEncoding))
			_ = s.Write(request.Protocol, resp)
			return false
		}

		if len(transferEncoding) > 0 {
			// get rid of the trailing chunked encoding as it is already built-in.
			if err = s.applyDecoders(transferEncoding[:len(transferEncoding)-1]); err != nil {
				// even if the connection is going to be upgraded in advance, the error happened with the
				// request prior to upgrade.
				resp := respond(request, s.router.OnError(request, err))
				_ = s.Write(request.Protocol, resp)
				return false
			}
		}

		if err = s.applyDecoders(request.Encoding.Content); err != nil {
			resp := respond(request, s.router.OnError(request, err))
			_ = s.Write(request.Protocol, resp)
			return false
		}

		version := request.Protocol
		if request.Upgrade != proto.Unknown && proto.HTTP1&request.Upgrade != 0 {
			s.Upgrade()
			version = request.Upgrade
		}

		resp := respond(request, s.router.OnRequest(request))

		if request.Hijacked() {
			// in case the connection was hijacked, we must not intrude after, so fail fast
			return false
		}

		if err = s.Write(version, resp); err != nil {
			// considering any write errors could occur due to broken connection, it makes
			// thereby no sense to try to write any error back. Moreover, there could be an
			// already sent data, which would overlay and result in a complete mess at the
			// client side.
			s.router.OnError(request, status.ErrCloseConnection)
			return false
		}

		if err = request.Body.Discard(); err != nil {
			resp = s.router.OnError(request, status.ErrCloseConnection)
			_ = s.Write(request.Protocol, resp)
			return false
		}

		if !isKeepAlive(version, request) {
			s.router.OnError(request, status.ErrCloseConnection)
			return true
		}

		if once {
			return true
		}

		request.Reset()
	}
}

func isKeepAlive(protocol proto.Protocol, req *http.Request) bool {
	switch protocol {
	case proto.HTTP10:
		return strutil.CmpFoldSafe(req.Connection, "keep-alive")
	case proto.HTTP11:
		// in case of HTTP/1.1, keep-alive may be only disabled
		return !strutil.CmpFoldSafe(req.Connection, "close")
	default:
		// as the protocol is unknown and the code was probably caused by some sort
		// of bug, consider closing it
		return false
	}
}

func validateTransferEncodingTokens(tokens []string) bool {
	if len(tokens) == 0 {
		return true
	}

	for _, token := range tokens[:len(tokens)-1] {
		if token == "chunked" {
			return false
		}
	}

	return tokens[len(tokens)-1] == "chunked"
}

func (s *Suit) applyDecoders(tokens []string) error {
	request := s.Parser.request

	for i := len(tokens); i > 0; i-- {
		c := s.codecs.Get(tokens[i-1])
		if c == nil {
			return status.ErrUnsupportedEncoding
		}

		if err := c.ResetDecompressor(request.Body.Fetcher); err != nil {
			return status.ErrInternalServerError
		}

		request.Body.Fetcher = c
	}

	return nil
}

// respond ensures the passed resp is not nil, otherwise http.Respond(req) is returned
func respond(req *http.Request, resp *http.Response) *http.Response {
	if resp != nil {
		return resp
	}

	return http.Respond(req)
}
