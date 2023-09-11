package parser

import (
	"github.com/indigo-web/indigo/client"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
)

type Parser interface {
	Init(headers *headers.Headers, body *http.Body)
	Parse(data []byte) (headersParsed bool, rest []byte, err error)
	Response() client.Response
}
