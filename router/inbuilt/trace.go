package inbuilt

import (
	"github.com/flrdv/uf"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/kv"
)

func traceHandler(request *http.Request) *http.Response {
	resp := request.Respond()
	// exploit the fact that Write method never returns an error
	_, _ = resp.Write(uf.S2B(request.Method.String()))
	_, _ = resp.Write([]byte(" "))
	// we probably should've escaped the path back, as otherwise this might lead to
	// some unwanted situations, however... Well, this doesn't seem to be effort-worthy.
	_, _ = resp.Write(uf.S2B(request.Path))
	if !request.Params.Empty() {
		pairs := request.Params.Expose()
		_, _ = resp.Write([]byte("?"))
		writeParam(resp, pairs[0])

		for _, pair := range pairs[1:] {
			// can avoid if len(pair.Key) == 0 { continue } (to filter out deleted entries),
			// because the TRACE handler is supposed to be executed if no other handler ran,
			// thereby having a completely virgin request, which was never touched by dirty
			// user's hands ever before. And won't after as well. Awesome.
			_, _ = resp.Write([]byte("&"))
			writeParam(resp, pair)
		}
	}

	_, _ = resp.Write([]byte(" "))
	_, _ = resp.Write(uf.S2B(request.Protocol.String()))
	_, _ = resp.Write([]byte("\r\n"))

	for key, value := range request.Headers.Pairs() {
		_, _ = resp.Write(uf.S2B(key))
		_, _ = resp.Write([]byte(": "))
		_, _ = resp.Write(uf.S2B(value))
		_, _ = resp.Write([]byte("\r\n"))
	}

	_, _ = resp.Write([]byte("\r\n"))

	return resp.ContentType(mime.HTTP)
}

func writeParam(resp *http.Response, param kv.Pair) {
	_, _ = resp.Write(uf.S2B(param.Key))
	_, _ = resp.Write([]byte("="))
	_, _ = resp.Write(uf.S2B(param.Value))
}
