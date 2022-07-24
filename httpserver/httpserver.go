package httpserver

import (
	"indigo/errors"
	"indigo/httpparser"
	"indigo/router"
	"indigo/types"
)

/*
HTTP server is a second and core layer. It receives data from tcp server,
delegates it to parser, parser fills parsed data directly into the request
structure, and then Session struct finally delegates request to a Router
(if request has been received completely)
*/

type poller struct {
	router        router.Router
	writeResponse types.ResponseWriter

	requestsChan chan *types.Request
	errChan      chan error
}

func (p *poller) Poll() {
	for {
		select {
		case request := <-p.requestsChan:
			if err := p.router.OnRequest(request, p.writeResponse); err != nil {
				p.router.OnError(err)
				p.errChan <- err
				return
			}

			// signalize a completion of request handling
			p.errChan <- nil
		case err := <-p.errChan:
			p.router.OnError(err)
			return
		}
	}
}

type HTTPHandler interface {
	OnData(b []byte) error
	Poll()
}

type HTTPHandlerArgs struct {
	Router     router.Router
	Request    *types.Request
	Parser     httpparser.HTTPRequestsParser
	RespWriter types.ResponseWriter
}

type httpHandler struct {
	request *types.Request
	parser  httpparser.HTTPRequestsParser
	poller  poller

	requestsChan chan *types.Request
	errChan      chan error
}

func NewHTTPHandler(args HTTPHandlerArgs) HTTPHandler {
	requestsChan, errChan := make(chan *types.Request), make(chan error)

	return &httpHandler{
		request: args.Request,
		parser:  args.Parser,
		poller: poller{
			router:        args.Router,
			writeResponse: args.RespWriter,
			requestsChan:  requestsChan,
			errChan:       errChan,
		},

		requestsChan: requestsChan,
		errChan:      errChan,
	}
}

func (c *httpHandler) Poll() {
	c.poller.Poll()
}

func (c *httpHandler) OnData(data []byte) (err error) {
	var done bool

	for len(data) > 0 {
		done, data, err = c.parser.Parse(data)

		if err != nil {
			c.poller.errChan <- errors.ErrParsingRequest
			return err
		}

		if done {
			c.poller.requestsChan <- c.request

			if err = <-c.poller.errChan; err != nil {
				return err
			}
		}
	}

	return nil
}
