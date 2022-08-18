package server

import (
	"indigo/errors"
	"indigo/http/parser"
	"indigo/router"
	"indigo/types"
)

// HTTPServer provides 2 methods:
// - Run: starts requests processor, or what they need I don't know.
//        Method is supposed to be blocking, so in a separated goroutine
//        expected to be started
// - OnData: main thing here. It parses request, and sends a signal into
//           the gateway to notify requests processor goroutine, or what
//           they need I don't know
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

// Run starts request processor in blocking mode
func (h httpServer) Run() {
	h.requestProcessor()
}

// OnData is a core-core function here, because does all the main stuff
// core must do. It parses a data provided by tcp server, and according
// to the parser state returned, decides what to do
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
			case errors.ErrBadRequest, errors.ErrURIDecoding:
				h.notifier <- badRequest
			case errors.ErrTooManyHeaders, errors.ErrTooLarge:
				h.notifier <- requestEntityTooLarge
			case errors.ErrURITooLong:
				h.notifier <- requestURITooLong
			case errors.ErrHeaderFieldsTooLarge:
				h.notifier <- requestHeaderFieldsTooLarge
			case errors.ErrUnsupportedProtocol:
				h.notifier <- unsupportedProtocol
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

// requestProcessor is a top function in the whole userspace (requests processing
// space), it receives a signal from notifier chan and decides what to do starting
// from the actual signal. Also, when called, calls router OnStart() method
func (h httpServer) requestProcessor() {
	h.router.OnStart()

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
		case requestHeaderFieldsTooLarge:
			h.router.OnError(h.request, h.respWriter, errors.ErrHeaderFieldsTooLarge)
			h.notifier <- processed
			return
		case requestURITooLong:
			h.router.OnError(h.request, h.respWriter, errors.ErrURITooLong)
			h.notifier <- processed
			return
		case unsupportedProtocol:
			h.router.OnError(h.request, h.respWriter, errors.ErrUnsupportedProtocol)
			h.notifier <- processed
			return
		}
	}
}
