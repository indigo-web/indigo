package http

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/utils/uf"
	"os"
)

type Server struct {
	router         router.Router
	upgradePreResp *http.Response
}

func NewServer(router router.Router) *Server {
	return &Server{
		router:         router,
		upgradePreResp: http.NewResponse(),
	}
}

func (h *Server) Run(
	client tcp.Client, req *http.Request, trans transport.Transport,
) {
	for {
		if !h.HandleRequest(client, req, trans) {
			break
		}
	}

	_ = client.Close()
}

func (h *Server) HandleRequest(
	client tcp.Client, req *http.Request, trans transport.Transport,
) (continue_ bool) {
	data, err := client.Read()
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			err = status.ErrConnectionTimeout
		} else {
			err = status.ErrCloseConnection
		}

		_ = trans.Write(req.Proto, req, h.router.OnError(req, err), client)
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
				h.upgradePreResp.
					Code(status.SwitchingProtocols).
					Header("Connection", "upgrade").
					Header("Upgrade", protoToken),
			)
			protocol = req.Upgrade
		}

		client.Unread(extra)
		req.Body.Init(req)
		response := h.router.OnRequest(req)
		if response == nil {
			response = req.Respond()
		}

		if req.WasHijacked() {
			return false
		}

		if err = trans.Write(protocol, req, response, client); err != nil {
			// in case we failed to render the response, just close the connection silently.
			// This may affect cases, when the error occurred during rendering an attachment,
			// but server anyway cannot recognize them, so the only thing will be done here
			// is notifying the router about disconnection
			h.router.OnError(req, status.ErrCloseConnection)
			return false
		}

		if err = req.Clear(); err != nil {
			// abusing the fact, that req.Clear() will return an error ONLY if socket error
			// occurred while reading.
			// TODO: what's if the error lays in decoding? This should somehow be processed
			h.router.OnError(req, status.ErrCloseConnection)
			return false
		}
	case transport.Error:
		// as fatal error already happened and connection will anyway be closed, we don't
		// care about any socket errors anymore
		_ = trans.Write(req.Proto, req, h.router.OnError(req, err), client)
		return false
	default:
		panic(fmt.Sprintf("BUG: got unexpected parser state"))
	}

	return true
}
