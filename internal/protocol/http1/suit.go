package http1

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/buffer"
	"github.com/indigo-web/indigo/internal/codecutil"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/transport"
)

type Suit struct {
	*parser
	*body
	*serializer
	router router.Router
	client transport.Client
}

func newSuit(
	cfg *config.Config,
	r router.Router,
	request *http.Request,
	client transport.Client,
	body *body,
	keysBuff, valsBuff, requestLineBuff buffer.Buffer,
	respBuff []byte,
) *Suit {
	return &Suit{
		parser:     newParser(cfg, request, keysBuff, valsBuff, requestLineBuff),
		body:       body,
		serializer: newSerializer(cfg, client, respBuff, cfg.Headers.Default, request),
		router:     r,
		client:     client,
	}
}

// New instantiates an HTTP/1 protocol suit.
func New(
	cfg *config.Config, r router.Router, client transport.Client,
	req *http.Request, decoders codecutil.Cache[http.Decompressor],
) *Suit {
	keysBuff, valsBuff, requestLineBuff := construct.Buffers(cfg)
	respBuff := make([]byte, 0, cfg.HTTP.ResponseBuffer.Default)
	body := newBody(client, cfg.Body, decoders)

	return newSuit(
		cfg, r, req, client, body, keysBuff, valsBuff, requestLineBuff, respBuff,
	)
}

func (s *Suit) ServeOnce() bool {
	return s.serve(true)
}

func (s *Suit) Serve() {
	s.serve(false)
}

func (s *Suit) serve(once bool) (ok bool) {
	client := s.client
	request := s.parser.request

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
			continue
		}

		client.Pushback(extra)

		request.Body.Reset(request)
		if err = s.body.Reset(request); err != nil {
			// an error could occur here only if there were applied unrecognized encodings.
			// Unfortunately, as the client made a decision to bravely ignore the list of
			// encodings we do support, we don't want to read all the crap he sent us either.
			s.router.OnError(request, err)
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
			// if error happened during writing the response, it makes no sense to try
			// to write anything again
			s.router.OnError(request, status.ErrCloseConnection)
			return false
		}

		request.Reset()
		if err = request.Body.Discard(); err != nil {
			s.router.OnError(request, status.ErrCloseConnection)
			return false
		}

		if once {
			return true
		}
	}
}

// respond ensures the passed resp is not nil, otherwise http.Respond(req) is returned
func respond(req *http.Request, resp *http.Response) *http.Response {
	if resp != nil {
		return resp
	}

	return http.Respond(req)
}
