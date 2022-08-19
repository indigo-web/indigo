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
	// HeadersNumber is responsible for headers map size
	// Default value is an initial size of allocated headers map
	// Maximal value is maximum number of headers allowed to be presented
	HeadersNumber Setting[uint8]

	// HeadersKeyLength is responsible for header key length
	// Default value is an initial size of header key buffer allocated in parser
	// Maximal value is a maximal length of header key
	HeadersKeyLength Setting[uint8]

	// HeadersValueLength is responsible for header value length
	// Default value is an initial size for every header value
	// Maximal value is a maximal possible length for header
	HeadersValueLength Setting[uint16]

	// URLLength is responsible for URL buffer
	// Default value is an initial size of URL buffer
	// Maximal value is a maximal length of URL (protocol and method are
	//         included, so real limit will be a bit less than specified one,
	//         depends on method and protocol)
	URLLength Setting[uint16]

	// TCPServerRead is responsible for tcp server reading buffer settings
	// Default value is a size of buffer for reading from socket, also
	//         we can call this setting as a "how many bytes are read from
	//         socket at most"
	TCPServerRead Setting[uint16]

	// BodyLength is responsible for body length parameters
	// Default value stands for nothing, it's unused
	// Maximal value is a maximal length of body
	BodyLength Setting[uint32]

	// BodyChunkSize is responsible for chunks in chunked transfer encoding mode
	// Default value also stands for nothing because of peculiar properties of
	//         chunked body parser
	// Maximal value is a maximal length of chunk
	BodyChunkSize Setting[uint32]
)

type (
	Headers struct {
		Number      HeadersNumber
		KeyLength   HeadersKeyLength
		ValueLength HeadersValueLength
	}

	URL struct {
		Length URLLength
	}

	TCPServer struct {
		Read TCPServerRead
	}

	Body struct {
		Length    BodyLength
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
			ValueLength: HeadersValueLength{
				Default: 4096,
				Maximal: 8192,
			},
		},
		URL: URL{
			Length: URLLength{
				Default: 8192,
				Maximal: math.MaxUint16,
			},
		},
		TCPServer: TCPServer{
			Read: TCPServerRead{
				Default: 2048,
			},
		},
		Body: Body{
			Length: BodyLength{
				Default: 1024,
				Maximal: math.MaxUint32,
			},
			ChunkSize: BodyChunkSize{
				Default: 4096,
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
		original.Headers.Number.Default, defaultSettings.Headers.Number.Default,
	)
	original.Headers.Number.Maximal = customOrDefault(
		original.Headers.Number.Maximal, defaultSettings.Headers.Number.Maximal,
	)
	original.Headers.KeyLength.Default = customOrDefault(
		original.Headers.KeyLength.Default, defaultSettings.Headers.KeyLength.Default,
	)
	original.Headers.KeyLength.Maximal = customOrDefault(
		original.Headers.KeyLength.Maximal, defaultSettings.Headers.KeyLength.Maximal,
	)
	original.Headers.ValueLength.Default = customOrDefault(
		original.Headers.ValueLength.Default, defaultSettings.Headers.ValueLength.Default,
	)
	original.Headers.ValueLength.Maximal = customOrDefault(
		original.Headers.ValueLength.Maximal, defaultSettings.Headers.ValueLength.Maximal,
	)
	original.URL.Length.Default = customOrDefault(
		original.URL.Length.Default, defaultSettings.URL.Length.Default,
	)
	original.URL.Length.Maximal = customOrDefault(
		original.URL.Length.Maximal, defaultSettings.URL.Length.Maximal,
	)
	original.TCPServer.Read.Default = customOrDefault(
		original.TCPServer.Read.Default, defaultSettings.TCPServer.Read.Default,
	)
	original.TCPServer.Read.Maximal = customOrDefault(
		original.TCPServer.Read.Maximal, defaultSettings.TCPServer.Read.Maximal,
	)
	original.Body.Length.Default = customOrDefault(
		original.Body.Length.Default, defaultSettings.Body.Length.Default,
	)
	original.Body.Length.Maximal = customOrDefault(
		original.Body.Length.Maximal, defaultSettings.Body.Length.Maximal,
	)
	original.Body.ChunkSize.Default = customOrDefault(
		original.Body.ChunkSize.Default, defaultSettings.Body.ChunkSize.Default,
	)
	original.Body.ChunkSize.Maximal = customOrDefault(
		original.Body.ChunkSize.Maximal, defaultSettings.Body.ChunkSize.Maximal,
	)

	return original
}

func customOrDefault[T number](custom, defaultVal T) T {
	if custom == 0 {
		return defaultVal
	}

	return custom
}
