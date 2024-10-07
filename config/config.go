package config

import (
	"math"
	"time"

	"github.com/indigo-web/utils/constraint"
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

	URLBufferSize struct {
		Default, Maximal int
	}

	Query struct {
		PreAlloc int
	}
)

type (
	URL struct {
		// MaxLength is a size for buffer that'll be allocated once and will be kept
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
		// to response with 431 Header Fields Too Large
		ValueSpace HeadersValuesSpace
		// MaxEncodingTokens is a limit of how many encodings can be applied at the body
		// for a single request
		MaxEncodingTokens int
		// Default headers are those, which will be rendered on each response unless they were
		// not overridden by user
		Default map[string]string
		// CookiesPreAllocate defines the size of keyvalue.Storage, which is used to store cookies
		// once on demand. Therefore, it's going to be allocated only if used
		CookiesPreAllocate int
	}

	Body struct {
		// MaxSize describes the maximal size of a body, that can be processed. 0 will discard
		// any request with body (each call to request's body will result in status.ErrBodyTooLarge)
		MaxSize uint
		// MaxChunkSize is responsible for a maximal size of a single chunk being transferred
		// via chunked TE
		MaxChunkSize int64
		// DecodingBufferSize is a size of a buffer, used to store decoded request's body
		DecodingBufferSize int64
		// BufferPrealloc defines the initial length of the buffer when the whole body at once
		// is requested (normally via String() or Bytes() methods)
		BufferPrealloc uint64
	}

	HTTP struct {
		// ResponseBuffSize is responsible for a response buffer that is being allocated when
		// client connects and is used for rendering the response into it
		ResponseBuffSize int
		// FileBuffSize defines the size of the read buffer when reading a file
		FileBuffSize int
	}

	TCP struct {
		// ReadBufferSize is a size of buffer in bytes which will be used to read from
		// socket
		ReadBufferSize int
		// ReadTimeout is a duration after which client will be automatically disconnected
		ReadTimeout time.Duration
	}
)

type Config struct {
	URL     URL
	Headers Headers
	Body    Body
	HTTP    HTTP
	TCP     TCP
}

// Default returns default config. Those are initially well-balanced, however maximal defaults
// are pretty permitting
func Default() *Config {
	return &Config{
		URL: URL{
			BufferSize: URLBufferSize{
				// allocate 2kb buffer by default for storing URI (including query and protocol)
				Default: 2 * 1024,
				// allow at most 64kb of URI, including query and protocol
				Maximal: math.MaxUint16,
				// NOTE: setting the maximal value too small (e.g. smaller than 10-15 bytes) may cause
				// strange and ambiguous HTTP errors
			},
			Query: Query{
				PreAlloc: 10,
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
			MaxEncodingTokens:  15,
			Default:            DefaultHeaders,
			CookiesPreAllocate: 5,
		},
		Body: Body{
			MaxSize:      512 * 1024 * 1024, // 512 megabytes
			MaxChunkSize: 128 * 1024,        // 128 kilobytes
			// 8 kilobytes is by default twice more than TCPs read buffer, so must
			// be enough to avoid multiple reads per single TCP chunk
			DecodingBufferSize: 8 * 1024,
			BufferPrealloc:     1024,
		},
		HTTP: HTTP{
			ResponseBuffSize: 1024,
			FileBuffSize:     64 * 1024, // 64kb read buffer for files is pretty much sufficient
		},
		TCP: TCP{
			ReadBufferSize: 4 * 1024,
			ReadTimeout:    90 * time.Second,
		},
	}
}

// Fill fills zero-values from partially-filled Config instance with default ones
func Fill(src *Config) (new *Config) {
	defaults := Default()

	return &Config{
		URL: URL{
			BufferSize: URLBufferSize{
				Default: numOr(src.URL.BufferSize.Default, defaults.URL.BufferSize.Default),
				Maximal: numOr(src.URL.BufferSize.Maximal, defaults.URL.BufferSize.Maximal),
			},
			Query: Query{
				PreAlloc: numOr(src.URL.Query.PreAlloc, defaults.URL.Query.PreAlloc),
			},
		},
		Headers: Headers{
			Number: HeadersNumber{
				Default: numOr(src.Headers.Number.Default, defaults.Headers.Number.Default),
				Maximal: numOr(src.Headers.Number.Maximal, defaults.Headers.Number.Maximal),
			},
			MaxKeyLength:   numOr(src.Headers.MaxKeyLength, defaults.Headers.MaxKeyLength),
			MaxValueLength: numOr(src.Headers.MaxValueLength, defaults.Headers.MaxValueLength),
			KeySpace: HeadersKeysSpace{
				Default: numOr(src.Headers.KeySpace.Default, defaults.Headers.KeySpace.Default),
				Maximal: numOr(src.Headers.KeySpace.Maximal, defaults.Headers.KeySpace.Maximal),
			},
			ValueSpace: HeadersValuesSpace{
				Default: numOr(src.Headers.ValueSpace.Default, defaults.Headers.ValueSpace.Default),
				Maximal: numOr(src.Headers.ValueSpace.Maximal, defaults.Headers.ValueSpace.Maximal),
			},
			MaxEncodingTokens: numOr(src.Headers.MaxEncodingTokens, defaults.Headers.MaxEncodingTokens),
			Default:           mapOr(src.Headers.Default, defaults.Headers.Default),
		},
		Body: Body{
			MaxSize:            numOr(src.Body.MaxSize, defaults.Body.MaxSize),
			MaxChunkSize:       numOr(src.Body.MaxChunkSize, defaults.Body.MaxChunkSize),
			DecodingBufferSize: numOr(src.Body.DecodingBufferSize, defaults.Body.DecodingBufferSize),
		},
		HTTP: HTTP{
			ResponseBuffSize: numOr(src.HTTP.ResponseBuffSize, defaults.HTTP.ResponseBuffSize),
			FileBuffSize:     numOr(src.HTTP.FileBuffSize, defaults.HTTP.FileBuffSize),
		},
		TCP: TCP{
			ReadBufferSize: numOr(src.TCP.ReadBufferSize, defaults.TCP.ReadBufferSize),
			ReadTimeout:    numOr(src.TCP.ReadTimeout, defaults.TCP.ReadTimeout),
		},
	}
}

func numOr[T constraint.Number](custom, defaultVal T) T {
	if custom == 0 {
		return defaultVal
	}

	return custom
}

func mapOr[K comparable, V any](custom, defaultVal map[K]V) map[K]V {
	if custom == nil {
		return defaultVal
	}

	return custom
}
