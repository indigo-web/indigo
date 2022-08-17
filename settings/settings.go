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
	HeadersNumber       Setting[uint8]
	HeaderKeyBuffSize   Setting[uint8]
	HeaderValueBuffSize Setting[uint16]
	URLBuffSize         Setting[uint16]
	SockReadBufferSize  Setting[uint16]
	BodyLength          Setting[uint32]
	BodyBuff            Setting[uint32]
	BodyChunkSize       Setting[uint32]
)

type Settings struct {
	HeadersNumber       HeadersNumber
	HeaderKeyBuffSize   HeaderKeyBuffSize
	HeaderValueBuffSize HeaderValueBuffSize
	URLBuffSize         URLBuffSize
	SockReadBufferSize  SockReadBufferSize
	BodyLength          BodyLength
	BodyBuff            BodyBuff
	BodyChunkSize       BodyChunkSize
}

func Default() Settings {
	// Usually, Default field stands for size of pre-allocated something
	// and Maximal stands for maximal size of something

	return Settings{
		HeadersNumber: HeadersNumber{
			Default: math.MaxUint8 / 4,
			Maximal: math.MaxUint8,
		},
		HeaderKeyBuffSize: HeaderKeyBuffSize{
			// I heard Apache has the same
			Default: 100,
			Maximal: 100,
		},
		HeaderValueBuffSize: HeaderValueBuffSize{
			Default: math.MaxUint16 / 8,
			Maximal: math.MaxUint16,
		},
		URLBuffSize: URLBuffSize{
			// math.MaxUint16 / 32 == 1024
			Default: math.MaxUint16 / 32,
			Maximal: math.MaxUint16,
		},
		SockReadBufferSize: SockReadBufferSize{
			// in case of SockReadBufferSize, we don't have an option of growth,
			// so only one of them is used
			Default: 2048,
			Maximal: 2048,
		},
		BodyLength: BodyLength{
			Default: math.MaxUint32,
			Maximal: math.MaxUint32,
		},
		BodyBuff: BodyBuff{
			Default: 0,
			Maximal: math.MaxUint32,
		},
		BodyChunkSize: BodyChunkSize{
			// in case of BodyChunkSize, we don't have an option of growth,
			// too
			Default: math.MaxUint32,
			Maximal: math.MaxUint32,
		},
	}
}

// Fill takes some settings and fills it with default values
// everywhere where it is not filled
func Fill(original Settings) (modified Settings) {
	defaultSettings := Default()

	original.HeadersNumber.Default = customOrDefault(
		original.HeadersNumber.Default, defaultSettings.HeadersNumber.Default,
	)
	original.HeadersNumber.Maximal = customOrDefault(
		original.HeadersNumber.Maximal, defaultSettings.HeadersNumber.Maximal,
	)
	original.HeaderKeyBuffSize.Default = customOrDefault(
		original.HeaderKeyBuffSize.Default, defaultSettings.HeaderKeyBuffSize.Default,
	)
	original.HeaderKeyBuffSize.Maximal = customOrDefault(
		original.HeaderKeyBuffSize.Maximal, defaultSettings.HeaderKeyBuffSize.Maximal,
	)
	original.HeaderValueBuffSize.Default = customOrDefault(
		original.HeaderValueBuffSize.Default, defaultSettings.HeaderValueBuffSize.Default,
	)
	original.HeaderValueBuffSize.Maximal = customOrDefault(
		original.HeaderValueBuffSize.Maximal, defaultSettings.HeaderValueBuffSize.Maximal,
	)
	original.URLBuffSize.Default = customOrDefault(
		original.URLBuffSize.Default, defaultSettings.URLBuffSize.Default,
	)
	original.URLBuffSize.Maximal = customOrDefault(
		original.URLBuffSize.Maximal, defaultSettings.URLBuffSize.Maximal,
	)
	original.SockReadBufferSize.Default = customOrDefault(
		original.SockReadBufferSize.Default, defaultSettings.SockReadBufferSize.Default,
	)
	original.SockReadBufferSize.Maximal = customOrDefault(
		original.SockReadBufferSize.Maximal, defaultSettings.SockReadBufferSize.Maximal,
	)
	original.BodyLength.Default = customOrDefault(
		original.BodyLength.Default, defaultSettings.BodyLength.Default,
	)
	original.BodyLength.Maximal = customOrDefault(
		original.BodyLength.Maximal, defaultSettings.BodyLength.Maximal,
	)
	original.BodyBuff.Default = customOrDefault(
		original.BodyBuff.Default, defaultSettings.BodyBuff.Default,
	)
	original.BodyBuff.Maximal = customOrDefault(
		original.BodyBuff.Maximal, defaultSettings.BodyBuff.Maximal,
	)
	original.BodyChunkSize.Default = customOrDefault(
		original.BodyChunkSize.Default, defaultSettings.BodyChunkSize.Default,
	)
	original.BodyChunkSize.Maximal = customOrDefault(
		original.BodyChunkSize.Maximal, defaultSettings.BodyChunkSize.Maximal,
	)

	return original
}

func customOrDefault[T number](custom, defaultVal T) T {
	if custom == 0 {
		return defaultVal
	}

	return custom
}
