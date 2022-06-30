package httpparser

import "math"

const (
	DefaultPathLength        = 2048
	DefaultHeaderLength      = 100
	DefaultHeaderValueLength = 1024
	DefaultChunkSize         = 8192
	DefaultBodyLength        = math.MaxUint32
	DefaultInfoLineBuffSize  = 10
	DefaultHeaderBuffSize    = 10
)

type Settings struct {
	// MaxPathLength may be at most 65535, but I think it's overkill
	MaxPathLength uint16

	// MaxHeaderLength stands not only for max header length, but
	// also for maximal length of parameter. Yes, for headers and
	// params settings are same
	MaxHeaderLength uint8

	// MaxHeaderValueLength stands for max length of value can be presented
	MaxHeaderValueLength uint16

	// MaxChunkSize stands for maximal size chunk may have in chunked transfer
	// encoding
	MaxChunkSize uint

	// MaxBodyLength stands for max body length is possible to be processed
	MaxBodyLength uint

	// InfoLineBuffer stands for a buffer that keeps method, path, etc.
	InfoLineBuffer []byte

	// HeadersBuffer is the same as InfoLineBuffer but
	HeadersBuffer []byte
}

func PrepareSettings(settings Settings) Settings {
	if settings.MaxPathLength == 0 {
		settings.MaxPathLength = DefaultPathLength
	}
	if settings.MaxHeaderLength == 0 {
		settings.MaxHeaderLength = DefaultHeaderLength
	}
	if settings.MaxHeaderValueLength == 0 {
		settings.MaxHeaderValueLength = DefaultHeaderValueLength
	}
	if settings.MaxChunkSize == 0 {
		settings.MaxChunkSize = DefaultChunkSize
	}
	if settings.MaxBodyLength == 0 {
		settings.MaxBodyLength = DefaultBodyLength
	}
	if settings.InfoLineBuffer == nil {
		settings.InfoLineBuffer = make([]byte, 0, DefaultInfoLineBuffSize)
	}
	if settings.HeadersBuffer == nil {
		settings.HeadersBuffer = make([]byte, 0, DefaultHeaderBuffSize)
	}

	return settings
}
