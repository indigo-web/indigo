package settings

import "math"

type number interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64
}

type Setting[T number] struct {
	Default T // soft limit
	Maximal T // hard limit
}

type Settings struct {
	HeadersNumber       Setting[uint8]
	HeaderKeyBuffSize   Setting[uint8]
	HeaderValueBuffSize Setting[uint16]
	URLBuffSize         Setting[uint16]
	SockReadBufferSize  Setting[uint16]
	BodyLength          Setting[uint32]
	BodyBuff            Setting[uint32]
	BodyChunkSize       Setting[uint32]
}

func Default() Settings {
	// Usually, Default field stands for size of pre-allocated something
	// and Maximal stands for maximal size of something

	return Settings{
		HeadersNumber: Setting[uint8]{
			Default: math.MaxUint8 / 4,
			Maximal: math.MaxUint8,
		},
		HeaderKeyBuffSize: Setting[uint8]{
			// I heard Apache has the same
			Default: 100,
			Maximal: 100,
		},
		HeaderValueBuffSize: Setting[uint16]{
			Default: math.MaxUint16 / 8,
			Maximal: math.MaxUint16,
		},
		URLBuffSize: Setting[uint16]{
			// math.MaxUint16 / 32 == 1024
			Default: math.MaxUint16 / 32,
			Maximal: math.MaxUint16,
		},
		SockReadBufferSize: Setting[uint16]{
			// in case of SockReadBufferSize, we don't have an option of growth,
			// so only one of them is used
			Default: 2048,
			Maximal: 2048,
		},
		BodyLength: Setting[uint32]{
			Default: math.MaxUint32,
			Maximal: math.MaxUint32,
		},
		BodyBuff: Setting[uint32]{
			Default: 0,
			Maximal: math.MaxUint32,
		},
		BodyChunkSize: Setting[uint32]{
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

	if original.HeadersNumber.Default == 0 {
		original.HeadersNumber.Default = defaultSettings.HeadersNumber.Default
	}
	if original.HeadersNumber.Maximal == 0 {
		original.HeadersNumber.Maximal = defaultSettings.HeadersNumber.Maximal
	}
	if original.HeaderKeyBuffSize.Default == 0 {
		original.HeaderKeyBuffSize.Default = defaultSettings.HeaderKeyBuffSize.Default
	}
	if original.HeaderKeyBuffSize.Maximal == 0 {
		original.HeaderKeyBuffSize.Maximal = defaultSettings.HeaderKeyBuffSize.Maximal
	}
	if original.HeaderValueBuffSize.Default == 0 {
		original.HeaderValueBuffSize.Default = defaultSettings.HeaderValueBuffSize.Default
	}
	if original.HeaderValueBuffSize.Maximal == 0 {
		original.HeaderValueBuffSize.Maximal = defaultSettings.HeaderValueBuffSize.Maximal
	}
	if original.URLBuffSize.Default == 0 {
		original.URLBuffSize.Default = defaultSettings.URLBuffSize.Default
	}
	if original.URLBuffSize.Maximal == 0 {
		original.URLBuffSize.Maximal = defaultSettings.URLBuffSize.Maximal
	}
	if original.SockReadBufferSize.Default == 0 {
		original.SockReadBufferSize.Default = defaultSettings.SockReadBufferSize.Default
	}
	if original.SockReadBufferSize.Maximal == 0 {
		original.SockReadBufferSize.Maximal = defaultSettings.SockReadBufferSize.Maximal
	}
	if original.BodyLength.Default == 0 {
		original.BodyLength.Default = defaultSettings.BodyLength.Default
	}
	if original.BodyLength.Maximal == 0 {
		original.BodyLength.Maximal = defaultSettings.BodyLength.Maximal
	}
	if original.BodyBuff.Default == 0 {
		original.BodyBuff.Default = defaultSettings.BodyBuff.Default
	}
	if original.BodyBuff.Maximal == 0 {
		original.BodyBuff.Maximal = defaultSettings.BodyBuff.Maximal
	}
	if original.BodyChunkSize.Default == 0 {
		original.BodyChunkSize.Default = defaultSettings.BodyChunkSize.Default
	}
	if original.BodyChunkSize.Maximal == 0 {
		original.BodyChunkSize.Maximal = defaultSettings.BodyChunkSize.Maximal
	}

	return original
}
