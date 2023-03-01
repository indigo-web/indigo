package settings

import (
	"math"
	"time"

	"github.com/indigo-web/indigo/v2/internal/constraints"
)

type (
	HeadersNumber struct {
		Default, Maximal int
	}
	MaxHeaderKeyLength   = int
	MaxHeaderValueLength = int
	HeadersValuesSpace   struct {
		Default, Maximal int
	}
	HeadersValuesObjectPoolSize = int
	MaxURLLength                int

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

	TCPReadBuffSize int
	TCPReadTimeout  = time.Duration

	MaxBodySize      int
	MaxBodyChunkSize = int

	ResponseBuffSize int
)

type (
	Headers struct {
		// Number is responsible for headers map size.
		// Default value is an initial size of allocated headers map.
		// Maximal value is maximum number of headers allowed to be presented
		Number HeadersNumber
		// MaxKeyLength is responsible for maximal header key length restriction.
		MaxKeyLength MaxHeaderKeyLength
		// MaxValueLength is responsible for maximal header value length restriction.
		MaxValueLength MaxHeaderValueLength
		// HeadersValuesSpace is responsible for a maximal space in bytes available for
		// keeping header values in memory.
		// Default value is initial space allocated when client connects.
		// Maximal value is a hard limit, reaching which one client triggers server
		// to response with 431 Header Fields Too Large
		ValueSpace HeadersValuesSpace
		// MaxValuesObjectPoolSize is responsible for a maximal size of string slices object
		// pool
		MaxValuesObjectPoolSize HeadersValuesObjectPoolSize
	}

	URL struct {
		// MaxLength is a size for buffer that'll be allocated once and will be kept
		// until client disconnect
		MaxLength MaxURLLength
		Query     Query
	}

	TCP struct {
		// ReadBufferSize is a size of buffer in bytes which will be used to read from
		// socket
		ReadBufferSize TCPReadBuffSize
		// ReadTimeout is a duration after which client will be automatically disconnected
		ReadTimeout TCPReadTimeout
	}

	Body struct {
		// MaxSize is responsible for a maximal body size in case it is being transferred
		// using ordinary Content-Length header, otherwise (e.g. chunked TE) this limit,
		// unfortunately, doesn't work
		MaxSize MaxBodySize
		// MaxChunkSize is responsible for a maximal size of a single chunk being transferred
		// via chunked TE
		MaxChunkSize MaxBodyChunkSize
	}

	HTTP struct {
		// ResponseBuffSize is responsible for a response buffer that is being allocated when
		// client connects and is used for rendering the response into it
		ResponseBuffSize ResponseBuffSize
	}
)

type Settings struct {
	Headers Headers
	URL     URL
	TCP     TCP
	Body    Body
	HTTP    HTTP
}

func Default() Settings {
	// Usually, Default field stands for size of pre-allocated something
	// and Maximal stands for maximal size of something

	return Settings{
		Headers: Headers{
			Number: HeadersNumber{
				Default: 10,
				Maximal: 100,
			},
			MaxKeyLength:   100,  // 100 bytes
			MaxValueLength: 8192, // 8 kilobytes (just like nginx)
			ValueSpace: HeadersValuesSpace{
				// for simple requests without many header values this will be enough, I hope
				Default: 1024,
				// 128kb as a limit of amount of memory for header values storing
				Maximal: 128 * 1024,
			},
		},
		URL: URL{
			MaxLength: math.MaxUint16,
			Query: Query{
				MaxLength:      math.MaxUint16,
				DefaultMapSize: 20,
			},
		},
		TCP: TCP{
			ReadBufferSize: 2048,
			ReadTimeout:    90 * time.Second,
		},
		Body: Body{
			MaxSize:      math.MaxUint32,
			MaxChunkSize: math.MaxUint32,
		},
		HTTP: HTTP{
			ResponseBuffSize: 1024,
		},
	}
}

// Fill takes some settings and fills it with default values
// everywhere where it is not filled
func Fill(original Settings) (modified Settings) {
	defaultSettings := Default()

	original.Headers.Number.Default = customOrDefault(
		original.Headers.Number.Default, defaultSettings.Headers.Number.Default)
	original.Headers.Number.Maximal = customOrDefault(
		original.Headers.Number.Maximal, defaultSettings.Headers.Number.Maximal)
	original.Headers.MaxKeyLength = customOrDefault(
		original.Headers.MaxKeyLength, defaultSettings.Headers.MaxKeyLength)
	original.Headers.MaxValueLength = customOrDefault(
		original.Headers.MaxValueLength, defaultSettings.Headers.MaxValueLength)
	original.Headers.ValueSpace.Default = customOrDefault(
		original.Headers.ValueSpace.Default, defaultSettings.Headers.ValueSpace.Default)
	original.Headers.ValueSpace.Maximal = customOrDefault(
		original.Headers.ValueSpace.Maximal, defaultSettings.Headers.ValueSpace.Maximal)
	original.Headers.MaxValuesObjectPoolSize = customOrDefault(
		original.Headers.MaxValuesObjectPoolSize, defaultSettings.Headers.MaxValuesObjectPoolSize)
	original.URL.MaxLength = customOrDefault(
		original.URL.MaxLength, defaultSettings.URL.MaxLength)
	original.URL.Query.MaxLength = customOrDefault(
		original.URL.Query.MaxLength, defaultSettings.URL.Query.MaxLength)
	original.URL.Query.DefaultMapSize = customOrDefault(
		original.URL.Query.DefaultMapSize, defaultSettings.URL.Query.DefaultMapSize)
	original.TCP.ReadBufferSize = customOrDefault(
		original.TCP.ReadBufferSize, defaultSettings.TCP.ReadBufferSize)
	original.TCP.ReadTimeout = customOrDefault(
		original.TCP.ReadTimeout, defaultSettings.TCP.ReadTimeout)
	original.Body.MaxSize = customOrDefault(
		original.Body.MaxSize, defaultSettings.Body.MaxSize)
	original.Body.MaxChunkSize = customOrDefault(
		original.Body.MaxChunkSize, defaultSettings.Body.MaxChunkSize)
	original.HTTP.ResponseBuffSize = customOrDefault(
		original.HTTP.ResponseBuffSize, defaultSettings.HTTP.ResponseBuffSize)

	return original
}

func customOrDefault[T constraints.Number](custom, defaultVal T) T {
	if custom == 0 {
		return defaultVal
	}

	return custom
}
