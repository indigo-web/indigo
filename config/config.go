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

	URLBufferSize struct {
		Default, Maximal int
	}

	Query struct {
		// ParamsPrealloc sets the initial capacity of Params (aka keyvalue.Storage) storage.
		ParamsPrealloc int
		// BufferPrealloc sets the initial capacity of decoding buffer for queries. It is used
		// to store decoded keys and values, as they are a subject of urlencoding, too.
		BufferPrealloc int
		// DefaultFlagValue sets the default value for all the flags. A flag is a query key without
		// a value (can either be set empty or absent at all.) It is prohibited by HTTP RFC, however
		// is used in the wild web. Unfortunate that there's a need to tolerate in general.
		DefaultFlagValue string
	}
)

type (
	URL struct {
		// BufferSize is a size for buffer that'll be allocated once and will be kept
		// until client disconnect
		BufferSize URLBufferSize
		Query      Query
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
		Default map[string]string
		// CookiesPrealloc defines the initial keyvalue.Storage capacity, used to store the cookies
		// itself.
		CookiesPrealloc int
	}

	Body struct {
		// MaxSize describes the maximal size of a body, that can be processed. 0 will discard
		// any request with body (each call to request's body will result in status.ErrBodyTooLarge)
		MaxSize uint
		// MaxChunkSize is responsible for a maximal size of a single chunk being transferred
		// via chunked TE
		MaxChunkSize int
		// DecodingBufferSize is a size of a buffer, used to store decoded request's body
		DecodingBufferSize int
		Form               BodyForm
	}

	HTTP struct {
		// ResponseBuffSize is responsible for a response buffer that is being allocated when
		// client connects and is used for rendering the response into it
		ResponseBuffSize int
		// FileBuffSize defines the size of the read buffer when reading a file
		FileBuffSize int
		//Codecs       []codec.Codec
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
	URL     URL
	Headers Headers
	Body    Body
	HTTP    HTTP
	NET     NET
}

// Default returns default config. Those are initially well-balanced, however maximal defaults
// are pretty permitting.
func Default() *Config {
	return &Config{
		URL: URL{
			BufferSize: URLBufferSize{
				// allocate 2kb buffer by default for storing URI (including query and protocol)
				Default: 2 * 1024,
				// allow at most 16kb of URI, including query and protocol. The limit is pretty much
				// tolerant, as most web entities are limiting it to 4-8kb.
				Maximal: 16 * 1024,
			},
			Query: Query{
				ParamsPrealloc: 10,
				// considering queries are generally not that huge, this must be fairly enough
				// on average.
				BufferPrealloc:   256,
				DefaultFlagValue: "1",
			},
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
			MaxSize:      512 * 1024 * 1024, // 512 megabytes
			MaxChunkSize: 128 * 1024,        // 128 kilobytes
			// 8 kilobytes is by default twice more than NETs read buffer, so must
			// be enough to avoid multiple reads per single NET chunk
			DecodingBufferSize: 8 * 1024,
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
			ResponseBuffSize: 1024,
			FileBuffSize:     64 * 1024, // 64kb read buffer for files is pretty much sufficient
		},
		NET: NET{
			ReadBufferSize:            4 * 1024, // 4kb is more than enough for ordinary requests.
			ReadTimeout:               90 * time.Second,
			AcceptLoopInterruptPeriod: 5 * time.Second,
		},
	}
}
