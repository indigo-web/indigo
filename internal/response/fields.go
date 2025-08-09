package response

import (
	"io"

	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/kv"
)

type Fields struct {
	Code            status.Code
	Status          status.Status
	ContentEncoding string
	Charset         mime.Charset
	Stream          io.Reader
	StreamSize      int64
	Buffer          []byte
	Headers         []kv.Pair
	Cookies         []cookie.Cookie
}

func (f *Fields) Clear() {
	*f = Fields{
		Code:    status.OK,
		Buffer:  f.Buffer[:0],
		Headers: f.Headers[:0],
		Cookies: f.Cookies[:0],
	}
}
