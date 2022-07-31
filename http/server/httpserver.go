package server

import (
	"indigo/errors"
	"indigo/http/parser"
	"indigo/router"
	"indigo/types"
)

/*
HTTP server is a second and core layer. It receives data from tcp server,
delegates it to parser, parser fills parsed data directly into the request
structure, and then Session struct finally delegates request to a Router
(if request has been received completely)
*/

type (
	requestsChan chan *types.Request
	errorsChan   chan error
)

type poller struct {
	router        router.Router
	writeResponse types.ResponseWriter

	reqChan requestsChan
	errChan errorsChan
}

func (p *poller) Poll() {
	for {
		select {
		case request := <-p.reqChan:
			if err := p.router.OnRequest(request, p.writeResponse); err != nil {
				p.router.OnError(err)
				p.errChan <- err
				return
			}
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
	Parser     parser.HTTPRequestsParser
	RespWriter types.ResponseWriter
}

type httpHandler struct {
	request *types.Request
	parser  parser.HTTPRequestsParser
	poller  poller

	reqChan requestsChan
	errChan errorsChan
}

func NewHTTPHandler(args HTTPHandlerArgs) HTTPHandler {
	reqChan, errChan := make(requestsChan), make(errorsChan)

	return newHTTPHandler(args, reqChan, errChan)
}

func newHTTPHandler(args HTTPHandlerArgs, reqChan requestsChan, errChan errorsChan) *httpHandler {
	return &httpHandler{
		request: args.Request,
		parser:  args.Parser,
		poller: poller{
			router:        args.Router,
			writeResponse: args.RespWriter,
			reqChan:       reqChan,
			errChan:       errChan,
		},
		reqChan: reqChan,
		errChan: errChan,
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
			c.poller.reqChan <- c.request
		}
	}

	return nil
}
