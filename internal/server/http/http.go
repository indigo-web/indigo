package http

import (
	"errors"
	"fmt"
	"os"

	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/render"
	"github.com/indigo-web/indigo/internal/server/tcp"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/parser"
	"github.com/indigo-web/indigo/router"
)

type Server interface {
	Run(tcp.Client, *http.Request, http.BodyReader, render.Renderer, parser.HTTPRequestsParser)
}

type BenchmarkServer interface {
	RunOnce(tcp.Client, *http.Request, http.BodyReader, render.Renderer, parser.HTTPRequestsParser)
}

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
	renderer render.Renderer, p parser.HTTPRequestsParser,
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
	renderer render.Renderer, p parser.HTTPRequestsParser,
) bool {
	data, err := client.Read()
	if err != nil {
		if errors.Is(err, os.ErrDeadlineExceeded) {
			err = status.ErrConnectionTimeout
		} else {
			err = status.ErrCloseConnection
		}

		_ = renderer.Render(req, h.router.OnError(req, err), client.Write)
		return false
	}

	state, extra, err := p.Parse(data)
	switch state {
	case parser.Pending:
	case parser.Error:
		h.router.OnError(req, err)
		p.Release()
		return false
	case parser.HeadersCompleted, parser.RequestCompleted:
		client.Unread(extra)
		reader.Init(req)
		response := h.router.OnRequest(req)

		if req.WasHijacked() {
			_ = client.Close()
			return false
		}

		if err = renderer.Render(req, response, client.Write); err != nil {
			h.router.OnError(req, status.ErrCloseConnection)
			return false
		}

		p.Release()
		if err = req.Reset(); err != nil {
			h.router.OnError(req, status.ErrCloseConnection)
			return false
		}
	default:
		panic(fmt.Sprintf("BUG: got unexpected parser state: %d", state))
	}

	return true
}
