package mime

import (
	"github.com/indigo-web/indigo/internal/strutil"
)

type MIME = string

const (
	OctetStream    MIME = "application/octet-stream"
	Plain          MIME = "text/plain"
	HTML           MIME = "text/html"
	XML            MIME = "text/xml"
	JSON           MIME = "application/json"
	YAML           MIME = "application/yaml"
	PDF            MIME = "application/pdf"
	FormUrlencoded MIME = "application/x-www-form-urlencoded"
	Multipart      MIME = "multipart/form-data"
	ZIP            MIME = "application/zip"
	GZIP           MIME = "application/gzip"
	ZLIB           MIME = "application/zlib"
	ZSTD           MIME = "application/zstd"
	AVIF           MIME = "image/avif"
	CSS            MIME = "text/css"
	GIF            MIME = "image/gif"
	JPEG           MIME = "image/jpeg"
	PNG            MIME = "image/png"
	SVG            MIME = "image/svg+xml"
	ICO            MIME = "image/vnd.microsoft.icon"
	WEBP           MIME = "image/webp"
	JS             MIME = "text/javascript"
	WASM           MIME = "application/wasm"
)

// Complies returns whether two MIMEs are compatible. Empty MIME is
// considered compatible with any other MIME
func Complies(mime MIME, with string) bool {
	// get rid of parameters if any
	with, _ = strutil.CutHeader(with)
	return len(with) == 0 || with == mime
}
