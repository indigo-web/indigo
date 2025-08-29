package config

import (
	"time"

	"github.com/indigo-web/indigo/http/mime"
)

type (
	HeadersNumber struct {
		Default, Maximal int
	}

	HeadersSpace struct {
		Default, Maximal int
	}

	BodyForm struct {
		// EntriesPrealloc is the number of preallocated seats for form.Form in body entity.
		EntriesPrealloc uint64
		// BufferPrealloc is the initial length for a buffer storing a whole request body, if its
		// length isn't known in advance (e.g. chunked transfer encoding.)
		BufferPrealloc uint64
		// DefaultCoding sets the default content encoding unless one is explicitly set.
		DefaultCoding mime.Charset
		// DefaultContentType sets the default form body MIME (as for multipart) unless one is
		// explicitly set.
		DefaultContentType mime.MIME
	}

	NETWriteBufferSize struct {
		Default, Maximal int
	}

	URIRequestLineSize struct {
		Default, Maximal int
	}
)

type (
	URI struct {
		// RequestLineSize is a shared buffer storing path and parameters. Also used to store method and
		// protocol in a form of an intermediate storage when they must be saved among calls. Please note
		// that setting the maximal boundary too low might result in very ambiguous errors.
		RequestLineSize URIRequestLineSize
		// ParamsPrealloc for http.Request.Params field.
		ParamsPrealloc int
	}

	Headers struct {
		// Number is responsible for headers map size.
		// Default value is an initial size of allocated headers map.
		// Maximal value is maximum number of headers allowed to be presented
		Number HeadersNumber
		// Space limits the amount of memory occupied by request headers.
		Space HeadersSpace
		// MaxEncodingTokens is a limit of how many encodings can be applied at the body
		// in a single request.
		MaxEncodingTokens int
		// MaxAcceptEncodingTokens is a limit for the buffer storing accepted by a client encodings.
		MaxAcceptEncodingTokens int
		// Default headers are headers to be included into every response implicitly, unless
		// explicitly overridden.
		Default map[string]string `test:"nullable"`
		// CookiesPrealloc defines the initial kv.Storage capacity, used to store the cookies
		// itself.
		CookiesPrealloc int
	}

	Body struct {
		// MaxSize describes the maximal size of a body, that can be processed. 0 will discard
		// any request with body (each call to request's body will result in status.ErrBodyTooLarge).
		// In order to disable the setting, use the math.MaxUInt64 value.
		MaxSize uint64
		// Form is either application/x-www-form-urlencoded or multipart/form-data. Due to their common
		// nature, they are easy to be generalized.
		Form BodyForm
	}

	NET struct {
		// ReadBufferSize is a size of buffer in bytes which will be used to read from
		// socket
		ReadBufferSize int
		// ReadTimeout controls the maximal lifetime of IDLE connections. If no data was
		// received in this period of time, it'll be closed.
		ReadTimeout time.Duration
		// AcceptLoopInterruptPeriod controls how often will the Accept() call be interrupted
		// in order to check whether it's time to stop. Defaults to 5 seconds.
		AcceptLoopInterruptPeriod time.Duration
		// WriteBufferSize stores the HTTP response, which is going to be transmitted.
		//
		// The buffer growth rules are:
		//  1) If a stream is sized (1) and its size overflows current buffer size (2),
		//   the buffer grows to contain the whole stream at once, but limited by the
		//   `HTTP.ResponseBuffer.Maximal`
		//  2) If a stream is unsized (1) and the previous write used more than ~98.44% of its
		//   capacity (2), the capacity doubles.
		WriteBufferSize NETWriteBufferSize
		// SmallBody limits how big must a response body be in order to be compressed, if the
		// auto compression option is enabled. This setting doesn't affect enforced compression
		// options and unsized streams.
		SmallBody int64
	}
)

// Config holds settings used across various parts of indigo, mainly restrictions, limitations
// and pre-allocations.
//
// You must ALWAYS modify defaults (returned via Default()) and NEVER try to initialize the
// config manually, because most likely this will result in ambiguous errors.
type Config struct {
	URI     URI
	Headers Headers
	Body    Body
	NET     NET
}

// Default returns default config. Those are initially well-balanced, however maximal defaults
// are pretty permitting.
func Default() *Config {
	return &Config{
		URI: URI{
			RequestLineSize: URIRequestLineSize{
				Default: 2 * 1024,
				// allow at most 16kb of request line, which is effectively pretty much tolerant,
				// considering most web-entities limit it to 4-8kb. However, we do also store path
				// parameters here, too.
				Maximal: 16 * 1024,
			},
			ParamsPrealloc: 5,
		},
		Headers: Headers{
			Number: HeadersNumber{
				Default: 10,
				Maximal: 50,
			},
			Space: HeadersSpace{
				Default: 1 * 1024,  // 1kb for headers must be fairly enough in most cases.
				Maximal: 16 * 1024, // However, there also might be extremely long cookies.
			},
			MaxEncodingTokens:       4,  // 1 for chunked, leaving at most 3 compressors to be composed
			MaxAcceptEncodingTokens: 20, // that must be a way too advanced client
			Default:                 make(map[string]string),
			CookiesPrealloc:         5,
		},
		Body: Body{
			MaxSize: 512 * 1024 * 1024, // 512 megabytes
			Form: BodyForm{
				EntriesPrealloc: 8,
				// 1kb is intended for primarily x-www-form-urlencoded, as multipart
				// needs of memory are assumingly fairly low
				BufferPrealloc:     1024,
				DefaultCoding:      mime.UTF8,
				DefaultContentType: mime.Plain,
			},
		},
		NET: NET{
			ReadBufferSize:            2 * 1024, // 4kb is more than enough for ordinary requests.
			ReadTimeout:               90 * time.Second,
			AcceptLoopInterruptPeriod: 5 * time.Second,
			WriteBufferSize: NETWriteBufferSize{
				Default: 2 * 1024,
				Maximal: 64 * 1024,
			},
			SmallBody: 4 * 1024,
		},
	}
}
