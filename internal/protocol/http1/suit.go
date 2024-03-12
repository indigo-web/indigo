package http1

import (
	"bytes"
	"fmt"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/internal/tcp"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/uf"
)

type Suit struct {
	*Parser
	*Serializer
	router         router.Router
	client         tcp.Client
	upgradePreResp *http.Response
	onDisconnect   func(req *http.Request) *http.Response
}

func New(
	cfg config.Config,
	r router.Router,
	request *http.Request,
	client tcp.Client,
	keyBuff, valBuff, startLineBuff *buffer.Buffer,
	respBuff []byte,
	respFileBuffSize int,
	defaultHeaders map[string]string,
) *Suit {
	return &Suit{
		Parser:         NewParser(request, keyBuff, valBuff, startLineBuff, cfg.Headers),
		Serializer:     NewSerializer(respBuff, respFileBuffSize, defaultHeaders, request, client),
		router:         r,
		client:         client,
		upgradePreResp: http.NewResponse(),
		onDisconnect:   cfg.HTTP.OnDisconnect,
	}
}

// Initialize is the same constructor as just New, but consumes fewer arguments.
func Initialize(cfg config.Config, r router.Router, client tcp.Client, req *http.Request) *Suit {
	keyBuff, valBuff, startLineBuff := construct.Buffers(cfg)
	respBuff := make([]byte, 0, cfg.HTTP.ResponseBuffSize)

	return New(
		cfg, r, req, client, keyBuff, valBuff, startLineBuff,
		respBuff, cfg.HTTP.ResponseBuffSize, cfg.Headers.Default,
	)
}

func (s *Suit) ServeOnce() bool {
	return s.serve(true)
}

func (s *Suit) Serve() {
	s.serve(false)
	if s.onDisconnect != nil {
		_ = s.Write(s.Parser.request.Proto, s.onDisconnect(s.Parser.request))
	}
}

func (s *Suit) serve(once bool) (ok bool) {
	for {
		req := s.Parser.request
		client := s.client

		data, err := client.Read()
		if err != nil {
			// read-error most probably means deadline exceeding. Just notify the user in
			// this case and return
			s.onError(req, status.ErrCloseConnection)
			return false
		}

		state, extra, err := s.Parse(data)
		switch state {
		case Pending:
		case HeadersCompleted:
			version := req.Proto

			if req.Upgrade != proto.Unknown && proto.HTTP1&req.Upgrade == req.Upgrade {
				protoToken := uf.B2S(bytes.TrimSpace(proto.ToBytes(req.Upgrade)))
				s.PreWrite(
					req.Proto,
					s.upgradePreResp.
						Code(status.SwitchingProtocols).
						Header("Connection", "upgrade").
						Header("Upgrade", protoToken),
				)
				version = req.Upgrade
			}

			client.Unread(extra)
			req.Body.Init(req)

			if req.WasHijacked() {
				return false
			}

			if err = s.Write(version, s.onRequest(req)); err != nil {
				// if error happened during writing the response, it makes no sense to try
				// to write anything again
				s.onError(req, status.ErrCloseConnection)
				return false
			}

			if err = req.Clear(); err != nil {
				// abusing the fact that req.Clear() can fail only due to read error
				s.onError(req, status.ErrCloseConnection)
				return false
			}
		case Error:
			// as fatal error already happened and connection will anyway be closed, we don't
			// care about any socket errors anymore
			_ = s.Write(req.Proto, s.onError(req, err))
		default:
			panic(fmt.Sprintf("BUG: got unexpected parser state"))
		}

		if once {
			return true
		}
	}
}

func (s *Suit) onError(req *http.Request, err error) *http.Response {
	return notNil(req, s.router.OnError(req, err))
}

func (s *Suit) onRequest(req *http.Request) *http.Response {
	return notNil(req, s.router.OnRequest(req))
}

func notNil(req *http.Request, resp *http.Response) *http.Response {
	if resp != nil {
		return resp
	}

	return http.Respond(req)
}
