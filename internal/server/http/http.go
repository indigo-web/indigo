package http

import (
	"bytes"
	"fmt"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/utils/uf"
)

type Server struct {
	router         router.Router
	upgradePreResp *http.Response
	onDisconnect   config.OnDisconnectCallback
}

func NewServer(router router.Router, onDisconnect config.OnDisconnectCallback) *Server {
	return &Server{
		router:         router,
		upgradePreResp: http.NewResponse(),
		onDisconnect:   onDisconnect,
	}
}

func (s *Server) Run(client tcp.Client, req *http.Request, trans transport.Transport) {
	for s.HandleRequest(client, req, trans) {
	}

	if s.onDisconnect != nil {
		_ = trans.Write(req.Proto, req, s.onDisconnect(req), client)
	}

	_ = client.Close()
}

func (s *Server) HandleRequest(client tcp.Client, req *http.Request, trans transport.Transport) (ok bool) {
	data, err := client.Read()
	if err != nil {
		_ = trans.Write(req.Proto, req, s.onError(req, err), client)
		return false
	}

	state, extra, err := trans.Parse(data)
	switch state {
	case transport.Pending:
	case transport.HeadersCompleted:
		protocol := req.Proto

		if req.Upgrade != proto.Unknown && proto.HTTP1&req.Upgrade == req.Upgrade {
			protoToken := uf.B2S(bytes.TrimSpace(proto.ToBytes(req.Upgrade)))
			trans.PreWrite(
				req.Proto,
				s.upgradePreResp.
					Code(status.SwitchingProtocols).
					Header("Connection", "upgrade").
					Header("Upgrade", protoToken),
			)
			protocol = req.Upgrade
		}

		client.Unread(extra)
		req.Body.Init(req)

		if req.WasHijacked() {
			return false
		}

		if err = trans.Write(protocol, req, s.onRequest(req), client); err != nil {
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
	case transport.Error:
		// as fatal error already happened and connection will anyway be closed, we don't
		// care about any socket errors anymore
		_ = trans.Write(req.Proto, req, s.onError(req, err), client)
		return false
	default:
		panic(fmt.Sprintf("BUG: got unexpected parser state"))
	}

	return true
}

func (s *Server) onError(req *http.Request, err error) *http.Response {
	return notNil(req, s.router.OnError(req, err))
}

func (s *Server) onRequest(req *http.Request) *http.Response {
	return notNil(req, s.router.OnRequest(req))
}

func notNil(req *http.Request, resp *http.Response) *http.Response {
	if resp != nil {
		return resp
	}

	return http.Respond(req)
}
