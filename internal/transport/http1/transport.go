package http1

import (
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/buffer"
	"github.com/indigo-web/utils/pool"
)

var _ transport.Transport = new(Transport)

type Transport struct {
	*Parser
	*Serializer
}

func New(
	request *http.Request,
	keyBuff, valBuff, startLineBuff *buffer.Buffer,
	valuesPool *pool.ObjectPool[[]string],
	headersSettings settings.Headers,
	respBuff []byte,
	respFileBuffSize int,
	defaultHeaders map[string]string,
) *Transport {
	return &Transport{
		Parser:     NewParser(request, keyBuff, valBuff, startLineBuff, valuesPool, headersSettings),
		Serializer: NewSerializer(respBuff, respFileBuffSize, defaultHeaders),
	}
}
