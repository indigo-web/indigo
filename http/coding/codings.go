package coding

import (
	"errors"
)

var (
	ErrUnknownToken = errors.New("coding token is not recognized")
)

type Token = string

// identity stands for "no encoding", according to RFC
const identity Token = "identity"

type Encoder interface {
	Encode(input []byte) (output []byte, err error)
}

type Decoder interface {
	Decode(input []byte) (output []byte, err error)
}

type Coding interface {
	Encoder
	Decoder
}

type constructor[T any] func(inBuff []byte) T

type Manager struct {
	encoders map[Token]Encoder
	decoders map[Token]Decoder
	buffSize int
}

func NewManager(buffSize int) Manager {
	return Manager{
		encoders: make(map[Token]Encoder),
		decoders: make(map[Token]Decoder),
		buffSize: buffSize,
	}
}

// AddEncoder adds a new encoder to the list of available
func (m Manager) AddEncoder(token Token, encoder constructor[Encoder]) {
	addCoding(token, encoder(m.newBuffer()), m.encoders)
}

// AddDecoder adds a new decoder to the list of available
func (m Manager) AddDecoder(token Token, decoder constructor[Decoder]) {
	addCoding(token, decoder(m.newBuffer()), m.decoders)
}

// AddCoding adds both Encoder and Decoder at the same time
func (m Manager) AddCoding(token Token, codingConstructor constructor[Coding]) {
	coding := codingConstructor(m.newBuffer())
	addCoding[Encoder](token, coding, m.encoders)
	addCoding[Decoder](token, coding, m.decoders)
}

func (m Manager) Encode(token Token, input []byte) (output []byte, err error) {
	if token == identity || len(input) == 0 {
		return input, nil
	}

	encoder, found := m.encoders[token]
	if !found {
		return nil, ErrUnknownToken
	}

	return encoder.Encode(input)
}

func (m Manager) Decode(token Token, input []byte) (output []byte, err error) {
	if token == identity || len(input) == 0 {
		return input, nil
	}

	decoder, found := m.decoders[token]
	if !found {
		return nil, ErrUnknownToken
	}

	return decoder.Decode(input)
}

func (m Manager) newBuffer() []byte {
	return make([]byte, m.buffSize)
}

func addCoding[V any](token Token, value V, into map[Token]V) {
	into[token] = value

	// this exists in backward-capability purposes. Some old clients may use x-gzip or
	// x-compress instead of regular gzip or compress tokens respectively.
	// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding#directives
	switch token {
	case "gzip":
		into["x-gzip"] = value
	case "compress":
		into["x-compress"] = value
	}
}
