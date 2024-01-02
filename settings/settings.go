package settings

import (
	"math"
	"time"

	"github.com/indigo-web/utils/constraint"
)

type (
	HeadersNumber struct {
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
	Headers struct {
		// Number is responsible for headers map size.
		// Default value is an initial size of allocated headers map.
		// Maximal value is maximum number of headers allowed to be presented
		Number HeadersNumber
		// MaxKeyLength is responsible for maximal header key length restriction.
		MaxKeyLength int
		// MaxValueLength is responsible for maximal header value length restriction.
		MaxValueLength int
		// HeadersValuesSpace is responsible for a maximal space in bytes available for
		// keeping header values in memory.
		// Default value is initial space allocated when client connects.
		// Maximal value is a hard limit, reaching which one client triggers server
		// to response with 431 Header Fields Too Large
		ValueSpace HeadersValuesSpace
		// MaxValuesObjectPoolSize is responsible for a maximal size of string slices object
		// pool
		MaxValuesObjectPoolSize int
		// MaxEncodingTokens is a limit of how many encodings can be applied at the body
		// for a single request
		MaxEncodingTokens int
	}

	URL struct {
		// MaxLength is a size for buffer that'll be allocated once and will be kept
		// until client disconnect
		BufferSize URLBufferSize
		Query      Query
	}

	TCP struct {
		// ReadBufferSize is a size of buffer in bytes which will be used to read from
		// socket
		ReadBufferSize int
		// ReadTimeout is a duration after which client will be automatically disconnected
		ReadTimeout time.Duration
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
	}

	HTTP struct {
		// ResponseBuffSize is responsible for a response buffer that is being allocated when
		// client connects and is used for rendering the response into it
		ResponseBuffSize int
		// FileBuffSize defines the size of the read buffer when reading a file
		FileBuffSize int
	}

	HTTPS struct {
		// Addr defines an address for HTTPS. If left empty, HTTPS will be disabled
		Addr string
		// Cert is a path to the .pem file with the actual certificate.
		// When using certbot, it is usually stored at /etc/letsencrypt/live/<domain>/fullchain.pem
		//
		// By default, just looking for fullchain.pem in the current working directory.
		Cert string
		// Key is a path to the .pem file with the actual key.
		// When using certbot, it is usually stored at /etc/letsencrypt/live/<domain>/privkey.pem
		//
		// By default, just looking for privkey.pem in the current working directory.
		Key string
	}
)

type Settings struct {
	Headers Headers
	URL     URL
	TCP     TCP
	Body    Body
	HTTP    HTTP
	HTTPS   HTTPS
}

// Default returns default settings. Those are initially well-balanced, however maximal defaults
// are pretty permitting
func Default() Settings {
	return Settings{
		Headers: Headers{
			Number: HeadersNumber{
				Default: 10,
				Maximal: 50,
			},
			MaxKeyLength:   100,      // basically 100 bytes
			MaxValueLength: 8 * 1024, // 8 kilobytes (just like nginx)
			ValueSpace: HeadersValuesSpace{
				// for simple requests without many heavy-weighted headers must be enough
				// to avoid a relatively big amount of re-allocations
				Default: 1 * 1024, // allocate 1kb buffer by default
				Maximal: 8 * 1024, // however allow at most 8kb of headers
			},
			// this may be an issue, if there are more custom encodings than this. However,
			// such cases are considered to be too rare
			MaxEncodingTokens: 15,
		},
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
		TCP: TCP{
			ReadBufferSize: 4 * 1024,
			ReadTimeout:    90 * time.Second,
		},
		Body: Body{
			MaxSize:      512 * 1024 * 1024, // 512 megabytes
			MaxChunkSize: 128 * 1024,        // 128 kilobytes
			// 8 kilobytes is by default twice more than TCPs read buffer, so must
			// be enough to avoid multiple reads per single TCP chunk
			DecodingBufferSize: 8 * 1024,
		},
		HTTP: HTTP{
			ResponseBuffSize: 1024,
			FileBuffSize:     64 * 1024, // 64kb read buffer for files is pretty much sufficient
		},
		HTTPS: HTTPS{
			Cert: "fullchain.pem",
			Key:  "privkey.pem",
		},
	}
}

// Fill fills zero-values from partially-filled Settings instance with default ones
func Fill(src Settings) (modified Settings) {
	defaults := Default()

	return Settings{
		Headers: Headers{
			Number: HeadersNumber{
				Default: valueOr(src.Headers.Number.Default, defaults.Headers.Number.Default),
				Maximal: valueOr(src.Headers.Number.Maximal, defaults.Headers.Number.Maximal),
			},
			MaxKeyLength:   valueOr(src.Headers.MaxKeyLength, defaults.Headers.MaxKeyLength),
			MaxValueLength: valueOr(src.Headers.MaxValueLength, defaults.Headers.MaxValueLength),
			ValueSpace: HeadersValuesSpace{
				Default: valueOr(src.Headers.ValueSpace.Default, defaults.Headers.ValueSpace.Default),
				Maximal: valueOr(src.Headers.ValueSpace.Maximal, defaults.Headers.ValueSpace.Maximal),
			},
			MaxValuesObjectPoolSize: valueOr(src.Headers.MaxValuesObjectPoolSize, defaults.Headers.MaxValuesObjectPoolSize),
			MaxEncodingTokens:       valueOr(src.Headers.MaxEncodingTokens, defaults.Headers.MaxEncodingTokens),
		},
		URL: URL{
			BufferSize: URLBufferSize{
				Default: valueOr(src.URL.BufferSize.Default, defaults.URL.BufferSize.Default),
				Maximal: valueOr(src.URL.BufferSize.Maximal, defaults.URL.BufferSize.Maximal),
			},
			Query: Query{
				PreAlloc: valueOr(src.URL.Query.PreAlloc, defaults.URL.Query.PreAlloc),
			},
		},
		TCP: TCP{
			ReadBufferSize: valueOr(src.TCP.ReadBufferSize, defaults.TCP.ReadBufferSize),
			ReadTimeout:    valueOr(src.TCP.ReadTimeout, defaults.TCP.ReadTimeout),
		},
		Body: Body{
			MaxSize:            valueOr(src.Body.MaxSize, defaults.Body.MaxSize),
			MaxChunkSize:       valueOr(src.Body.MaxChunkSize, defaults.Body.MaxChunkSize),
			DecodingBufferSize: valueOr(src.Body.DecodingBufferSize, defaults.Body.DecodingBufferSize),
		},
		HTTP: HTTP{
			ResponseBuffSize: valueOr(src.HTTP.ResponseBuffSize, defaults.HTTP.ResponseBuffSize),
			FileBuffSize:     valueOr(src.HTTP.FileBuffSize, defaults.HTTP.FileBuffSize),
		},
		HTTPS: HTTPS{
			// Addr is either defined or not. Default value - empty string - disables HTTPS
			Addr: src.HTTPS.Addr,
			Cert: strValueOr(src.HTTPS.Cert, defaults.HTTPS.Cert),
			Key:  strValueOr(src.HTTPS.Key, defaults.HTTPS.Key),
		},
	}
}

func valueOr[T constraint.Number](custom, defaultVal T) T {
	if custom == 0 {
		return defaultVal
	}

	return custom
}

func strValueOr(custom, defaultVal string) string {
	if len(custom) == 0 {
		return defaultVal
	}

	return custom
}
