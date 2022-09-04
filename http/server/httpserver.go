package server

import (
	"github.com/fakefloordiv/indigo/http"
	"net"

	"github.com/fakefloordiv/indigo/http/parser"
	"github.com/fakefloordiv/indigo/router"
	"github.com/fakefloordiv/indigo/types"
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
	HijackConn() net.Conn
}

type httpServer struct {
	request    *types.Request
	respWriter types.ResponseWriter
	router     router.Router
	parser     parser.HTTPRequestsParser
	conn       net.Conn

	notifier chan serverState
}

func NewHTTPServer(
	req *types.Request, respWriter types.ResponseWriter, router router.Router,
	parser parser.HTTPRequestsParser, conn net.Conn,
) HTTPServer {
	return &httpServer{
		request:    req,
		respWriter: respWriter,
		router:     router,
		parser:     parser,
		conn:       conn,
		notifier:   make(chan serverState),
	}
}

// Run first prepares request by setting up hijacker, then starts
// requests processor in blocking mode
func (h httpServer) Run() {
	h.request.Hijack = types.Hijacker(h.request, h.HijackConn)

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
			switch <-h.notifier {
			case processed:
			case connHijack:
				return errors.ErrHijackConn
			default:
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

			switch <-h.notifier {
			case processed:
			case connHijack:
				return errors.ErrHijackConn
			default:
				return errors.ErrCloseConnection
			}
		case parser.ConnectionClose:
			h.notifier <- closeConnection

			return nil
		case parser.Error:
			switch err {
			case errors.ErrBadRequest, errors.ErrURIDecoding:
				h.notifier <- badRequest
			case errors.ErrMethodNotImplemented:
				h.notifier <- methodNotImplemented
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
			// in case connection was hijacked, router does not know about it,
			// so he tries to write a response as usual. But he fails, because
			// connection is (must be) already closed. He returns an error, but
			// request processor... Also doesn't know about hijacking! That's why
			// here we are checking a notifier chan whether it's nil (it may be nil
			// ONLY here and ONLY because of hijacking)
			if h.router.OnRequest(h.request, h.respWriter) != nil {
				if h.notifier != nil {
					h.notifier <- closeConnection
				}

				return
			}
			if err := h.request.Reset(); err != nil {
				h.router.OnError(h.request, h.respWriter, err)
				h.notifier <- closeConnection
				return
			}

			h.notifier <- processed
			continue
		case closeConnection:
			h.router.OnError(h.request, h.respWriter, errors.ErrCloseConnection)
		case badRequest:
			h.router.OnError(h.request, h.respWriter, errors.ErrBadRequest)
		case methodNotImplemented:
			h.router.OnError(h.request, h.respWriter, errors.ErrMethodNotImplemented)
		case requestEntityTooLarge:
			h.router.OnError(h.request, h.respWriter, errors.ErrTooLarge)
		case requestHeaderFieldsTooLarge:
			h.router.OnError(h.request, h.respWriter, errors.ErrHeaderFieldsTooLarge)
		case requestURITooLong:
			h.router.OnError(h.request, h.respWriter, errors.ErrURITooLong)
		case unsupportedProtocol:
			h.router.OnError(h.request, h.respWriter, errors.ErrUnsupportedProtocol)
		default:
			panic("BUG: http/server/httpserver.go:requestProcessor(): received unknown state")
		}

		// most cases are doing this. Easier to add continue-statement somewhere
		h.notifier <- processed
		return
	}
}

func (h *httpServer) HijackConn() net.Conn {
	// HijackConn call can be initiated only by user. So in this case, we know
	// that server is in clearly defined state - waiting for body completion,
	// than waiting for a completion signal from requestProcessor. requestProcessor
	// cannot signalize anything because it's busy of waiting for handler completion
	// so in this case, we can just send some signal by ourselves. The idea is to
	// notify the core about hijacking, so it'll die silently, without closing the
	// connection. After successful core notifying, we're deleting our channel, making
	// it nil, so request processor MUST check it to avoid possible panic
	// note: for successful connection hijacking, request body MUST BE read before.
	//       Also, connection must be manually closed, otherwise router will write
	//       a http response there
	h.notifier <- connHijack
	h.notifier = nil

	return h.conn
}
