package http

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal"
	"github.com/indigo-web/indigo/internal/parser"
	"github.com/indigo-web/indigo/internal/render"
	"github.com/indigo-web/indigo/internal/server/tcp"
	"github.com/indigo-web/indigo/router"
	"os"
)

type Server interface {
	Run(tcp.Client, *http.Request, http.BodyReader, render.Engine, parser.HTTPRequestsParser)
}

type BenchmarkServer interface {
	RunOnce(tcp.Client, *http.Request, http.BodyReader, render.Engine, parser.HTTPRequestsParser)
}

var upgrading = http.NewResponse().
	WithCode(status.SwitchingProtocols).
	WithHeader("Connection", "upgrade")

type httpServer struct {
	router router.Router
}

func NewHTTPServer(router router.Router) Server {
	return &httpServer{
		router: router,
	}
}

func (h *httpServer) Run(
	client tcp.Client, req *http.Request, reader http.BodyReader,
	renderer render.Engine, p parser.HTTPRequestsParser,
) {
	for {
		if !h.RunOnce(client, req, reader, renderer, p) {
			break
		}
	}

	_ = client.Close()
}

func (h *httpServer) RunOnce(
	client tcp.Client, req *http.Request, reader http.BodyReader,
	renderer render.Engine, p parser.HTTPRequestsParser,
) bool {
	data, err := client.Read()
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			err = status.ErrConnectionTimeout
		} else {
			err = status.ErrCloseConnection
		}

		_ = renderer.Write(req.Proto, req, h.router.OnError(req, err), client.Write)
		return false
	}

	state, extra, err := p.Parse(data)
	switch state {
	case parser.Pending:
	case parser.HeadersCompleted:
		protocol := req.Proto

		if req.Upgrade != proto.Unknown && proto.HTTP1&req.Upgrade == req.Upgrade {
			protoToken := internal.B2S(bytes.TrimSpace(proto.ToBytes(req.Upgrade)))
			renderer.PreWrite(req.Proto, upgrading.WithHeader("Upgrade", protoToken))
			protocol = req.Upgrade
		}

		client.Unread(extra)
		reader.Init(req)
		response := h.router.OnRequest(req)

		if req.WasHijacked() {
			_ = client.Close()
			return false
		}

		if err = renderer.Write(protocol, req, response, client.Write); err != nil {
			h.router.OnError(req, status.ErrCloseConnection)
			return false
		}

		p.Release()
		if err = req.Clear(); err != nil {
			h.router.OnError(req, status.ErrCloseConnection)
			return false
		}
	case parser.Error:
		h.router.OnError(req, err)
		p.Release()
		return false
	default:
		panic(fmt.Sprintf("BUG: got unexpected parser state"))
	}

	return true
}
