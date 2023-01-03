package http

import (
	"github.com/fakefloordiv/indigo/http/status"
	"net"

	"github.com/fakefloordiv/indigo/internal/render"

	"github.com/fakefloordiv/indigo/http"
	"github.com/fakefloordiv/indigo/internal/parser"
	"github.com/fakefloordiv/indigo/router"
)

// HTTPServer provides 3 methods:
// - Run: starts requests processor, or what they need I don't know.
//        Method is supposed to be blocking, so in a separated goroutine
//        expected to be started
// - OnData: main thing here. It parses request, and sends a signal into
//           the gateway to notify requests processor goroutine, or what
//           they need I don't know
// - HijackConn: connection hijacker, of course
type HTTPServer interface {
	Run()
	OnData(b []byte) error
	HijackConn() net.Conn
}

type httpServer struct {
	request    *http.Request
	respWriter http.ResponseWriter
	router     router.Router
	parser     parser.HTTPRequestsParser
	conn       net.Conn
	renderer   *render.Renderer

	notifier chan serverState
	err      error
}

func NewHTTPServer(
	req *http.Request, router router.Router, parser parser.HTTPRequestsParser,
	conn net.Conn, renderer *render.Renderer,
) HTTPServer {
	server := &httpServer{
		request: req,
		respWriter: func(b []byte) error {
			_, err := conn.Write(b)
			return err
		},
		router:   router,
		parser:   parser,
		conn:     conn,
		renderer: renderer,
		notifier: make(chan serverState),
	}

	req.Hijack = http.Hijacker(req, server.HijackConn)

	return server
}

// Run first prepares request by setting up hijacker, then starts
// requests processor in blocking mode
func (h *httpServer) Run() {
	h.requestProcessor()
}

// OnData is a core-core function here, because does all the main stuff
// core must do. It parses a data provided by tcp server, and according
// to the parser state returned, decides what to do
func (h *httpServer) OnData(data []byte) (err error) {
	if len(data) == 0 {
		h.err = status.ErrConnectionTimeout
		h.notifier <- eError
		<-h.notifier

		return nil
	}

	var state parser.RequestState

	for len(data) > 0 {
		state, data, err = h.parser.Parse(data)

		switch state {
		case parser.Pending:
		case parser.HeadersCompleted:
			h.notifier <- eHeadersCompleted
		case parser.RequestCompleted:
			h.notifier <- eHeadersCompleted
			fallthrough
		case parser.BodyCompleted:
			switch <-h.notifier {
			case eProcessed:
				h.parser.Release()
			case eConnHijack:
				return status.ErrHijackConn
			default:
				return status.ErrCloseConnection
			}
		case parser.ConnectionClose:
			h.err = status.ErrCloseConnection
			h.notifier <- eError
			<-h.notifier

			return nil
		case parser.Error:
			if err == status.ErrURIDecoding {
				err = status.ErrBadRequest
			}

			h.err = err
			h.notifier <- eError

			// wait for processor to handle the error before connection will be closed
			// for example, respond client with error
			<-h.notifier

			return err
		default:
			panic("BUG: http/server/http.go:OnData(): received unknown state")
		}
	}

	return nil
}

// requestProcessor is a top function in the whole userspace (requests processing
// space), it receives a signal from notifier chan and decides what to do starting
// from the actual signal. Also, when called, calls router OnStart() method
func (h *httpServer) requestProcessor() {
	// implicitly dereference a method to avoid dereferences on every response,
	// that is actually not that cheap
	respRenderer := h.renderer.Response

	renderer := func(response http.Response) error {
		return respRenderer(h.request, response, h.respWriter)
	}

	for {
		switch <-h.notifier {
		case eHeadersCompleted:
			// in case connection was hijacked, router does not know about it,
			// so he tries to write a response as usual. But he fails, because
			// connection is (supposed to be) already closed. He returns an error, but
			// request processor... Also doesn't know about hijacking! That's why
			// here we are checking a notifier chan whether it's nil (it may be nil
			// ONLY here and ONLY because of hijacking)
			if renderer(h.router.OnRequest(h.request)) != nil {
				// request object must be reset in any way because otherwise
				// deadlock will happen here
				_ = h.request.Reset()

				if h.notifier != nil {
					h.notifier <- eError
				}

				return
			}

			if err := h.request.Reset(); err != nil {
				// we already sent a response that did not errored, so no way here
				// to send one more response with error. Just ignore it
				_ = h.router.OnError(h.request, err)
				h.notifier <- eError
				return
			}

			h.notifier <- eProcessed
		case eError:
			_ = renderer(h.router.OnError(h.request, h.err))
			h.notifier <- eProcessed
			return
		default:
			panic("BUG: http/server/http.go:requestProcessor(): received unknown state")
		}
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
	h.notifier <- eConnHijack
	h.notifier = nil

	return h.conn
}
