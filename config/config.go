package config

import (
	"github.com/indigo-web/indigo/http/mime"
	"time"
)

var DefaultHeaders = map[string]string{
	"Accept-Encodings": "identity",
}

type (
	HeadersNumber struct {
		Default, Maximal int
	}

	HeadersKeysSpace struct {
		Default, Maximal int
	}

	HeadersValuesSpace struct {
		Default, Maximal int
	}

	BodyForm struct {
		// EntriesPrealloc is the number of preallocated seats for form.Form in body entity.
		EntriesPrealloc uint64
		// BufferPrealloc defines the initial length of the buffer when the whole body at once
		// is requested (normally via String() or Bytes() methods.)
		BufferPrealloc uint64
		// DefaultCoding sets the default content encoding unless one is explicitly set.
		DefaultCoding string
		// DefaultContentType sets the default form body MIME (as for multipart) unless one is
		// explicitly set.
		DefaultContentType mime.MIME
	}

	HTTPResponseBuffer struct {
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
		// MaxKeyLength is responsible for maximal header key length restriction.
		MaxKeyLength int
		// MaxValueLength is responsible for maximal header value length restriction.
		MaxValueLength int
		// KeySpace is responsible for limitation of how much space can headers' keys
		// consume. Default value is how many bytes to pre-allocate, and maximal is
		// how many bytes can be stored maximally
		KeySpace HeadersKeysSpace
		// HeadersValuesSpace is responsible for a maximal space in bytes available for
		// keeping header values in memory.
		// Default value is initial space allocated when client connects.
		// Maximal value is a hard limit, reaching which one client triggers server
		// to response with 431 Header Fields Too Large.
		ValueSpace HeadersValuesSpace
		// MaxEncodingTokens is a limit of how many encodings can be applied at the body
		// in a single request.
		MaxEncodingTokens int
		// Default headers are those, which will be rendered on each response unless overridden explicitly.
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
		//// DecodingBufferSize is a size of a buffer, used to store decoded request's body
		//DecodingBufferSize int
		Form BodyForm
	}

	HTTP struct {
		// ResponseBuffer is used to store the byte-representation of a response, ready to be sent
		// over the network.
		//
		// Response buffer growth rules:
		//  1) If the stream is sized (1) and its size overflows current buffer length (2),
		//   grow it to contain the whole stream at once, but limit the size to at most
		//   `HTTP.ResponseBuffer.Maximal`
		//  2) If the stream is unsized (1) and the previous write used more than ~98.44% of its total
		//   capacity (2), the capacity doubles.
		ResponseBuffer HTTPResponseBuffer
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
	}
)

// Config holds settings used across various parts of indigo, mainly restrictions, limitations
// and pre-allocations.
//
// Please note: ALWAYS modify defaults (returned via Default()) and NEVER try to initialize the
// config manually, as this will result in highly ambiguous errors.
type Config struct {
	URI     URI
	Headers Headers
	Body    Body
	HTTP    HTTP
	NET     NET
}

// Default returns default config. Those are initially well-balanced, however maximal defaults
// are pretty permitting.
func Default() *Config {
	return &Config{
		URI: URI{
			RequestLineSize: URIRequestLineSize{
				Default: 2 * 1024,
				// allow at most 32kb. This limit is pretty much tolerant as most web entities
				// are limiting it to 4-8kb, however we do also store path parameters here.
				Maximal: 32 * 1024,
			},
			ParamsPrealloc: 5,
		},
		Headers: Headers{
			Number: HeadersNumber{
				Default: 10,
				Maximal: 50,
			},
			MaxKeyLength:   100,      // basically 100 bytes
			MaxValueLength: 8 * 1024, // 8 kilobytes (just like nginx)
			KeySpace: HeadersKeysSpace{
				Default: 1 * 1024,
				Maximal: 4 * 1024,
			},
			ValueSpace: HeadersValuesSpace{
				// for simple requests without many heavy-weighted headers must be enough
				// to avoid a relatively big amount of re-allocations
				// this may be an issue, if there are more custom encodings than this. However,
				// such cases are considered to be too rare
				Default: 1 * 1024, // allocate 1kb buffer by default
				Maximal: 8 * 1024, // however allow at most 8kb of headers
			},
			MaxEncodingTokens: 15,
			Default:           DefaultHeaders,
			CookiesPrealloc:   5,
		},
		Body: Body{
			MaxSize: 512 * 1024 * 1024, // 512 megabytes
			Form: BodyForm{
				EntriesPrealloc: 8,
				// 1kb is intended for primarily x-www-form-urlencoded, as multipart
				// needs of memory are fairly low
				BufferPrealloc:     1024,
				DefaultCoding:      "utf8",
				DefaultContentType: mime.Plain,
			},
		},
		HTTP: HTTP{
			ResponseBuffer: HTTPResponseBuffer{
				Default: 1024,
				Maximal: 64 * 1024,
			},
		},
		NET: NET{
			ReadBufferSize:            4 * 1024, // 4kb is more than enough for ordinary requests.
			ReadTimeout:               90 * time.Second,
			AcceptLoopInterruptPeriod: 5 * time.Second,
		},
	}
}
