package response

import (
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/types"
)

const DefaultContentType = "text/html"

type Fields struct {
	Attachment  types.Attachment
	Headers     []headers.Header
	Body        []byte
	Cookies     []cookie.Cookie
	Status      status.Status
	ContentType mime.MIME
	// TODO: add corresponding Content-Encoding field
	// TODO: automatically apply the encoding on a body when specified
	TransferEncoding string
	Code             status.Code
}

func (f *Fields) Clear() {
	f.Code = status.OK
	f.Status = ""
	f.ContentType = DefaultContentType
	f.TransferEncoding = ""
	f.Headers = f.Headers[:0]
	f.Body = nil
	f.Cookies = f.Cookies[:0]
	f.Attachment = types.Attachment{}
}
