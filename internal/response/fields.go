package response

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/types"
)

const DefaultContentType = "text/html"

type Header struct {
	Key, Value string
}

type Fields struct {
	Attachment  types.Attachment
	Status      status.Status
	ContentType string
	// TODO: add corresponding Content-Encoding field
	// TODO: automatically apply the encoding on a body when specified
	TransferEncoding string
	Headers          []Header
	Body             []byte
	Code             status.Code
}

func (f Fields) Clear() Fields {
	f.Code = status.OK
	f.Status = ""
	f.ContentType = DefaultContentType
	f.TransferEncoding = ""
	f.Headers = f.Headers[:0]
	f.Body = nil
	f.Attachment = types.Attachment{}

	return f
}
