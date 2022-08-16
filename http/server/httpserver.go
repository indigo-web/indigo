package server

import (
	"indigo/errors"
	"indigo/http/parser"
	"indigo/router"
	"indigo/types"
)

type HTTPServer interface {
	Run()
	OnData(b []byte) error
}

type httpServer struct {
	request    *types.Request
	respWriter types.ResponseWriter
	router     router.Router
	parser     parser.HTTPRequestsParser

	notifier chan serverState
}

func NewHTTPServer(
	req *types.Request, respWriter types.ResponseWriter,
	router router.Router, parser parser.HTTPRequestsParser,
) HTTPServer {
	return httpServer{
		request:    req,
		respWriter: respWriter,
		router:     router,
		parser:     parser,
		notifier:   make(chan serverState),
	}
}

func (h httpServer) Run() {
	h.requestProcessor()
}

func (h httpServer) OnData(data []byte) (err error) {
	var state parser.RequestState

	for len(data) > 0 {
		state, data, err = h.parser.Parse(data)

		switch state {
		case parser.Pending:
		case parser.HeadersCompleted:
			h.notifier <- headersCompleted
		case parser.BodyCompleted:
			if <-h.notifier != processed {
				return errors.ErrCloseConnection
			}
		case parser.RequestCompleted:
			h.notifier <- headersCompleted
			// the reason why we have manually finalize body is that
			// parser has not notified about request completion yet, so
			// handler is not called, but parser already has to write
			// something to a blocking chan. This causes a deadlock, so
			// we choose a bit hacky solution
			h.parser.FinalizeBody()
			if <-h.notifier != processed {
				return errors.ErrCloseConnection
			}
		case parser.ConnectionClose:
			h.notifier <- closeConnection

			return nil
		case parser.Error:
			switch err {
			case errors.ErrBadRequest, errors.ErrURLDecoding:
				h.notifier <- badRequest
			case errors.ErrTooLarge, errors.ErrURLTooLong, errors.ErrTooManyHeaders:
				h.notifier <- requestEntityTooLarge
			default:
				h.notifier <- badRequest
			}

			// wait for processor to handle the error before connection will be closed
			// for example, respond client with error
			<-h.notifier

			return err
		}
	}

	return nil
}

func (h httpServer) requestProcessor() {
	for {
		switch <-h.notifier {
		case headersCompleted:
			if h.router.OnRequest(h.request, h.respWriter) != nil {
				h.notifier <- closeConnection
				return
			}
			if err := h.request.Reset(); err != nil {
				h.router.OnError(h.request, h.respWriter, err)
				h.notifier <- closeConnection
				return
			}

			h.notifier <- processed
		case closeConnection:
			h.router.OnError(h.request, h.respWriter, errors.ErrCloseConnection)
			h.notifier <- processed
			return
		case badRequest:
			h.router.OnError(h.request, h.respWriter, errors.ErrBadRequest)
			h.notifier <- processed
			return
		case requestEntityTooLarge:
			h.router.OnError(h.request, h.respWriter, errors.ErrTooLarge)
			h.notifier <- processed
			return
		}
	}
}
