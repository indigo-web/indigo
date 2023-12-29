package http1

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/pool"
)

type Transport struct {
	parser *Parser
	dumper *Dumper
}

func New(
	request *http.Request,
	keyBuff, valBuff, startLineBuff buffer.Buffer,
	valuesPool pool.ObjectPool[[]string],
	headersSettings settings.Headers,
	respBuff []byte,
	respFileBuffSize int,
	defaultHeaders map[string]string,
) *Transport {
	return &Transport{
		parser: NewParser(request, keyBuff, valBuff, startLineBuff, valuesPool, headersSettings),
		dumper: NewDumper(respBuff, respFileBuffSize, defaultHeaders),
	}
}

func (t *Transport) Parse(b []byte) (state transport.RequestState, extra []byte, err error) {
	return t.parser.Parse(b)
}

func (t *Transport) PreDump(target proto.Proto, response *http.Response) {
	t.dumper.PreDump(target, response)
}

func (t *Transport) Dump(target proto.Proto, request *http.Request, response *http.Response, writer transport.Writer) error {
	return t.dumper.Dump(target, request, response, writer)
}
