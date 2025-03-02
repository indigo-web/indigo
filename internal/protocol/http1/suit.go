package http1

import (
	"fmt"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/transport"
	"github.com/indigo-web/utils/buffer"
)

type Suit struct {
	*parser
	*serializer
	upgradePreResp *http.Response
	body           *Body
	router         router.Router
	client         transport.Client
}

func New(
	cfg *config.Config,
	r router.Router,
	request *http.Request,
	client transport.Client,
	body *Body,
	keyBuff, valBuff, startLineBuff *buffer.Buffer,
	respBuff []byte,
	respFileBuffSize int,
) *Suit {
	return &Suit{
		parser:         newParser(request, keyBuff, valBuff, startLineBuff, cfg.Headers),
		serializer:     newSerializer(respBuff, respFileBuffSize, cfg.Headers.Default, request, client),
		upgradePreResp: http.NewResponse(),
		body:           body,
		router:         r,
		client:         client,
	}
}

// Initialize is the same constructor as just New, but consumes fewer arguments.
func Initialize(cfg *config.Config, r router.Router, client transport.Client, req *http.Request, body *Body) *Suit {
	keyBuff, valBuff, startLineBuff := construct.Buffers(cfg)
	respBuff := make([]byte, 0, cfg.HTTP.ResponseBuffSize)

	return New(
		cfg, r, req, client, body, keyBuff, valBuff, startLineBuff,
		respBuff, cfg.HTTP.ResponseBuffSize,
	)
}

func (s *Suit) ServeOnce() bool {
	return s.serve(true)
}

func (s *Suit) Serve() {
	s.serve(false)
}

func (s *Suit) serve(once bool) (ok bool) {
	req := s.parser.request
	client := s.client

	for {
		data, err := client.Read()
		if err != nil {
			// read-error most probably means deadline exceeding. Just notify the user in
			// this case and return
			s.router.OnError(req, status.ErrCloseConnection)
			return false
		}

		state, extra, err := s.Parse(data)
		switch state {
		case Pending:
		case HeadersCompleted:
			client.Unread(extra)
			s.body.Reset(req)

			version := req.Proto
			if req.Upgrade != proto.Unknown && proto.HTTP1&req.Upgrade == req.Upgrade {
				// TODO: replace this with a method "WriteUpgrade" or similar
				s.PreWrite(
					req.Proto,
					s.upgradePreResp.
						Code(status.SwitchingProtocols).
						Header("Connection", "upgrade").
						Header("Upgrade", trimLast(req.Proto.String())),
				)
				version = req.Upgrade
			}

			resp := respond(req, s.router.OnRequest(req))

			if req.Hijacked() {
				// in case the connection was hijacked, we must not intrude after, so fail fast
				return false
			}

			if err = s.Write(version, resp); err != nil {
				// if error happened during writing the response, it makes no sense to try
				// to write anything again
				s.router.OnError(req, status.ErrCloseConnection)
				return false
			}

			if err = req.Reset(); err != nil {
				// abusing the fact that req.Clear() can fail only due to read error
				s.router.OnError(req, status.ErrCloseConnection)
				return false
			}
		case Error:
			// as fatal error already happened and connection will anyway be closed, we don't
			// care about any socket errors anymore
			resp := respond(req, s.router.OnError(req, err))
			_ = s.Write(req.Proto, resp)
			return false
		default:
			panic(fmt.Sprintf("BUG: got unexpected parser state"))
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

func trimLast(s string) string {
	if len(s) == 0 {
		return s
	}

	return s[:len(s)-1]
}
