package mime

import (
	"path/filepath"

	"github.com/indigo-web/indigo/internal/strutil"
)

// Unset explicitly tells to omit the Charset or MIME. It's working for both, exploiting the fact
// that both are type aliases to string.
const Unset = ""

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
	JAVASCRIPT     MIME = "text/javascript"
	WASM           MIME = "application/wasm"
	SQL            MIME = "application/sql"
	TZIF           MIME = "application/tzif"
	XFDF           MIME = "application/xfdf"
	HTTP           MIME = "message/http"
)

// Complies returns whether two MIMEs are compatible. Empty MIME is considered
// compatible with any other MIME
func Complies(mime MIME, with string) bool {
	// get rid of parameters if any
	with, _ = strutil.CutHeader(with)
	return len(with) == 0 || with == mime
}

// Guess tries to guess the MIME type based on the file path, i.e. purely on the file extension.
// If no defaultMime is passed, an empty MIME is returned. Otherwise, the first one will be picked.
func Guess(path string, defaultMime ...MIME) MIME {
	ext := Extension[filepath.Ext(path)]
	if len(ext) == 0 && len(defaultMime) > 0 {
		return defaultMime[0]
	}

	return ext
}
