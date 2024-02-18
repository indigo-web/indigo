package http1

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/transport"
	"github.com/indigo-web/utils/buffer"
)

var _ transport.Transport = new(Transport)

type Transport struct {
	*Parser
	*Serializer
}

func New(
	request *http.Request,
	keyBuff, valBuff, startLineBuff *buffer.Buffer,
	headersCfg config.Headers,
	respBuff []byte,
	respFileBuffSize int,
	defaultHeaders map[string]string,
) *Transport {
	return &Transport{
		Parser:     NewParser(request, keyBuff, valBuff, startLineBuff, headersCfg),
		Serializer: NewSerializer(respBuff, respFileBuffSize, defaultHeaders),
	}
}
