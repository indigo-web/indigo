package settings

import "math"

type number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

type Setting[T number] struct {
	Default T // soft limit
	Maximal T // hard limit
}

type (
	HeadersNumber      Setting[uint8]
	HeadersKeyLength   Setting[uint8]
	HeadersValuesSpace Setting[uint32]

	URLLength Setting[uint16]

	// Query is responsible for url query settings.
	Query struct {
		// Length is responsible for a maximal length of url query may be
		// received.
		// Default value is unused
		Length QueryLength
		// Number is responsible for an initial capacity of query entries map.
		// Maximal value is unused because:
		//   Maximal number of entries equals to 65535 (math.MaxUint16) divided by
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
		Number QueryNumber
	}

	QueryLength Setting[uint16]
	QueryNumber Setting[uint16]

	TCPServerRead Setting[uint16]

	BodyLength    Setting[uint32]
	BodyChunkSize Setting[uint32]
)

type (
	Headers struct {
		// Number is responsible for headers map size.
		// Default value is an initial size of allocated headers map.
		// Maximal value is maximum number of headers allowed to be presented
		Number HeadersNumber
		// KeyLength is responsible for header key length.
		// Default value is an initial size of header key buffer allocated in parser.
		// Maximal value is a maximal length of header key
		KeyLength HeadersKeyLength
		// HeadersValuesSpace is responsible for a maximal space in bytes available for
		// keeping header values in memory.
		// Default value is initial space allocated when client connects.
		// Maximal value is a hard limit, reaching which one client triggers server
		// to response with 431 Header Fields Too Large
		ValueSpace HeadersValuesSpace
	}

	URL struct {
		// Length is responsible for URL buffer.
		// Default value is an initial size of URL buffer.
		// Maximal value is a maximal length of URL (protocol and method are
		// included, so real limit will be a bit less than specified one,
		// depends on method and protocol)
		Length URLLength
		Query  Query
	}

	TCPServer struct {
		// Read is responsible for tcp server reading buffer settings.
		// Default value is a size of buffer for reading from socket, also
		// we can call this setting as a "how many bytes are read from
		// socket at most"
		Read TCPServerRead
		// IDLEConnLifetime is a timer in seconds, after expiration of which one IDLE
		// connection will be actively closed by server.
		// IDLE conn is a connection that does not send anything
		// -1 as value disables timeouts at all
		IDLEConnLifetime int
	}

	Body struct {
		// Length is responsible for body length parameters.
		// Default value is unused.
		// Maximal value is a maximal length of body
		Length BodyLength
		// BodyChunkSize is responsible for chunks in chunked transfer encoding mode.
		// Default value is unused because chunked body parser calls callback with
		// data taken from input stream.
		// Maximal value is a maximal length of chunk
		ChunkSize BodyChunkSize
	}
)

type Settings struct {
	Headers   Headers
	URL       URL
	TCPServer TCPServer
	Body      Body
}

func Default() Settings {
	// Usually, Default field stands for size of pre-allocated something
	// and Maximal stands for maximal size of something

	return Settings{
		Headers: Headers{
			Number: HeadersNumber{
				Default: 25,
				Maximal: 100,
			},
			KeyLength: HeadersKeyLength{
				Default: 100,
				Maximal: math.MaxUint8,
			},
			ValueSpace: HeadersValuesSpace{
				// for simple requests without many header values this will be enough, I hope
				Default: 1024,
				// 128kb as a limit of amount of memory for header values storing
				Maximal: 128 * 1024,
			},
		},
		URL: URL{
			Length: URLLength{
				Default: 8192,
				Maximal: math.MaxUint16,
			},
			Query: Query{
				Length: QueryLength{
					Maximal: math.MaxUint16,
				},
				Number: QueryNumber{
					// I don't know why 20, but let it be
					Default: 20,
				},
			},
		},
		TCPServer: TCPServer{
			Read: TCPServerRead{
				Default: 2048,
			},
			IDLEConnLifetime: 90,
		},
		Body: Body{
			Length: BodyLength{
				Maximal: math.MaxUint32,
			},
			ChunkSize: BodyChunkSize{
				Maximal: math.MaxUint32,
			},
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
	original.Headers.KeyLength.Default = customOrDefault(
		original.Headers.KeyLength.Default, defaultSettings.Headers.KeyLength.Default)
	original.Headers.KeyLength.Maximal = customOrDefault(
		original.Headers.KeyLength.Maximal, defaultSettings.Headers.KeyLength.Maximal)
	original.Headers.ValueSpace.Default = customOrDefault(
		original.Headers.ValueSpace.Default, defaultSettings.Headers.ValueSpace.Default)
	original.Headers.ValueSpace.Maximal = customOrDefault(
		original.Headers.ValueSpace.Maximal, defaultSettings.Headers.ValueSpace.Maximal)
	original.URL.Length.Default = customOrDefault(
		original.URL.Length.Default, defaultSettings.URL.Length.Default)
	original.URL.Length.Maximal = customOrDefault(
		original.URL.Length.Maximal, defaultSettings.URL.Length.Maximal)
	original.URL.Query.Length.Maximal = customOrDefault(
		original.URL.Query.Length.Maximal, defaultSettings.URL.Query.Length.Maximal)
	original.URL.Query.Number.Default = customOrDefault(
		original.URL.Query.Number.Default, defaultSettings.URL.Query.Number.Default)
	original.TCPServer.Read.Default = customOrDefault(
		original.TCPServer.Read.Default, defaultSettings.TCPServer.Read.Default)
	original.TCPServer.IDLEConnLifetime = customOrDefault(
		original.TCPServer.IDLEConnLifetime, defaultSettings.TCPServer.IDLEConnLifetime)
	original.Body.Length.Default = customOrDefault(
		original.Body.Length.Default, defaultSettings.Body.Length.Default)
	original.Body.Length.Maximal = customOrDefault(
		original.Body.Length.Maximal, defaultSettings.Body.Length.Maximal)
	original.Body.ChunkSize.Maximal = customOrDefault(
		original.Body.ChunkSize.Maximal, defaultSettings.Body.ChunkSize.Maximal)

	return original
}

func customOrDefault[T number](custom, defaultVal T) T {
	if custom == 0 {
		return defaultVal
	}

	return custom
}
