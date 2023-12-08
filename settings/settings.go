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

	URLParams struct {
		// This option allows user to disable the automatic path params map clearing.
		// May be useful in cases where params keys are being accessed directly only,
		// and nothing tries to get all the map values
		DisableMapClear bool
	}

	Query struct {
		// MaxLength is responsible for a limit of the query length
		MaxLength int
		// DefaultMapSize is responsible for an initial capacity of query entries map.
		// There is no up limit because:
		//   Maximal number of entries equals to 65536 (math.MaxUint16) divided by
		//   3 (minimal length of query entry) that equals to 21,845.
		//   Worst case: sizeof(int) == 64 and sizeof(unsafe.Pointer) == 64. Then
		//   slice type takes 16 bytes
		//   In that case, we can calculate how much memory AT MOST will be used.
		//   24 bytes (slice type - cap, len and pointer 8 bytes each) + 1 byte
		//   (an array of a single char in best case) + 16 bytes (string type - len
		//   and pointer) + 1 byte (an array of single char in best case)
		//   42 bytes in total for each pair, 917490 bytes in total, that is 896 kilobytes
		//   that is 0.87 megabytes. IMHO that is not that much to care about. In case it
		//   is - somebody will open an issue, or even better, implement the limit by himself
		//   (hope he is lucky enough to find out how to handle with my hand-made DI)
		DefaultMapSize int
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
		Params     URLParams
	}

	TCP struct {
		// ReadBufferSize is a size of buffer in bytes which will be used to read from
		// socket
		ReadBufferSize int
		// ReadTimeout is a duration after which client will be automatically disconnected
		ReadTimeout time.Duration
	}

	Body struct {
		// MaxSize is responsible for a maximal body size in case it is being transferred
		// using ordinary Content-Length header, otherwise (e.g. chunked TE) this limit,
		// unfortunately, doesn't work
		MaxSize int64
		// MaxChunkSize is responsible for a maximal size of a single chunk being transferred
		// via chunked TE
		MaxChunkSize int64
		// DecodedBufferSize is a size of a buffer, used to store decompressed request body
		DecodedBufferSize int64
	}

	HTTP struct {
		// ResponseBuffSize is responsible for a response buffer that is being allocated when
		// client connects and is used for rendering the response into it
		ResponseBuffSize int64
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
			MaxKeyLength:   100,      // 100 bytes
			MaxValueLength: 8 * 1024, // 8 kilobytes (just like nginx)
			ValueSpace: HeadersValuesSpace{
				// for simple requests without many heavy-weighted headers must be enough
				// to avoid a relatively big amount of re-allocations
				Default: 2 * 1024,  // 2kb
				Maximal: 64 * 1024, // 64kb as a limit of amount of memory for header values storing
			},
			MaxEncodingTokens: 10,
		},
		URL: URL{
			BufferSize: URLBufferSize{
				Default: 4 * 1024, // 4kb
				Maximal: math.MaxUint16,
			},
			Query: Query{
				MaxLength:      math.MaxUint16,
				DefaultMapSize: 20,
			},
			Params: URLParams{
				DisableMapClear: false,
			},
		},
		TCP: TCP{
			ReadBufferSize: 4 * 1024,
			ReadTimeout:    90 * time.Second,
		},
		Body: Body{
			MaxSize:      math.MaxUint32,
			MaxChunkSize: math.MaxUint32,
			// 8 kilobytes is by default twice more than TCPs read buffer, so must
			// be enough to avoid multiple reads per single TCP chunk
			DecodedBufferSize: 8 * 1024,
		},
		HTTP: HTTP{
			ResponseBuffSize: 1024,
		},
		HTTPS: HTTPS{
			Addr: "0.0.0.0:443",
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
				MaxLength:      valueOr(src.URL.Query.MaxLength, defaults.URL.Query.MaxLength),
				DefaultMapSize: valueOr(src.URL.Query.DefaultMapSize, defaults.URL.Query.DefaultMapSize),
			},
			Params: URLParams{
				// as we can't determine, whether set value is a zero-value or set on purpose, just leave
				// it as it is. It's anyway equal to zero-value by default
				DisableMapClear: src.URL.Params.DisableMapClear,
			},
		},
		TCP: TCP{
			ReadBufferSize: valueOr(src.TCP.ReadBufferSize, defaults.TCP.ReadBufferSize),
			ReadTimeout:    valueOr(src.TCP.ReadTimeout, defaults.TCP.ReadTimeout),
		},
		Body: Body{
			MaxSize:           valueOr(src.Body.MaxSize, defaults.Body.MaxSize),
			MaxChunkSize:      valueOr(src.Body.MaxChunkSize, defaults.Body.MaxChunkSize),
			DecodedBufferSize: valueOr(src.Body.DecodedBufferSize, defaults.Body.DecodedBufferSize),
		},
		HTTP: HTTP{
			ResponseBuffSize: valueOr(src.HTTP.ResponseBuffSize, defaults.HTTP.ResponseBuffSize),
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
