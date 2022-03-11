package httpserver

import (
	"indigo/webserver"
	"io"
)

/*
I need a NewHTTPProtocol() constructor that will receive a pool of request objects (yes, sync.Pool,
how originally), where request object is already cleared (so there will be no difference between
new instance and already used), settings struct, handler, and... Fuck it, I need some Client{} struct
that will contain all these values, but also contain some callbacks to send a response. For example:
type Client struct {
	RequestObjectsPool sync.Pool
	Settings           webserver.Settings
	OnMessageComplete
}
*/

/*
InitialHeadersBufferSize is an initial capacity value when a new buffer for headers is allocated
This is made to avoid some useless allocations in the beginning when buffer doubles on first and second
header (1 -> 2, 2 -> 4, etc.), as usually we have at least 2-3 headers. So it's necessary to do that doubles
on such a small values
*/
const InitialHeadersBufferSize = 5

type httpProtocol struct {
	dispatcher webserver.Dispatcher
	requestCtx webserver.RequestCtx
	settings   webserver.Settings
	// TODO: maybe, there is a better way to keep a writer for each request?
	bodyWriter *io.PipeWriter
}

func (h httpProtocol) OnMessageBegin() error {
	// currently, I don't need to do anything here
	// maybe, I'd better remove this callback from parser? Calls are expensive enough
	return nil
}

func (h *httpProtocol) OnMethod(method []byte) error {
	h.request.Method = method

	return nil
}

func (h *httpProtocol) OnPath(path []byte) error {
	h.request.Path = path

	return nil
}

func (h *httpProtocol) OnProtocol(proto []byte) error {
	h.request.Protocol = proto

	return nil
}

func (h httpProtocol) OnHeadersBegin() error { return nil }

func (h *httpProtocol) OnHeader(key, value []byte) error {
	// TODO: add a sync.Pool here
	header := webserver.Header{
		Key:   key,
		Value: value,
	}

	if h.request.Headers.AppendAssertDuplicate(header) {
		// who is that shitbag who sent us a duplicated header?
		return webserver.ErrDuplicatedHeader
	}

	return nil
}

func (h *httpProtocol) OnHeadersComplete() error {
	reader, writer := io.Pipe()
	h.request.Body = reader
	h.bodyWriter = writer

}

func (h *httpProtocol) OnBody(piece []byte) error {
	// TODO: I need an io.Pipe() to set a reader to request
	//       maybe, create io.Pipe() directly in server, and pass it to the protocol?
	//       sounds like a good idea. I don't wanna do a lot of optimization stuff here, but in server

	panic("implement me")
}

func (h httpProtocol) OnMessageComplete() error {
	//TODO implement me
	panic("implement me")
}
