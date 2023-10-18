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

type Coding interface {
	Token() string
	Encode(input []byte) (output []byte, err error)
	Decode(input []byte) (output []byte, err error)
}

type Constructor func(buff []byte) Coding

type Manager struct {
	codings  map[Token]Coding
	buffSize int64
}

func NewManager(buffSize int64) Manager {
	return Manager{
		codings:  make(map[Token]Coding),
		buffSize: buffSize,
	}
}

// AddCoding adds a coding to the list of available
func (m Manager) AddCoding(constructor Constructor) {
	coding := constructor(m.newBuffer())
	m.codings[coding.Token()] = coding

	// this exists in backward-capability purposes. Some old clients may use x-gzip or
	// x-compress instead of regular gzip or compress tokens respectively.
	// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding#directives
	switch coding.Token() {
	case "gzip":
		m.codings["x-gzip"] = coding
	case "compress":
		m.codings["x-compress"] = coding
	}
}

func (m Manager) Encode(token Token, input []byte) (output []byte, err error) {
	if token == identity || len(input) == 0 {
		return input, nil
	}

	coding, found := m.codings[token]
	if !found {
		return nil, ErrUnknownToken
	}

	return coding.Encode(input)
}

func (m Manager) Decode(token Token, input []byte) (output []byte, err error) {
	if token == identity || len(input) == 0 {
		return input, nil
	}

	coding, found := m.codings[token]
	if !found {
		return nil, ErrUnknownToken
	}

	return coding.Decode(input)
}

func (m Manager) newBuffer() []byte {
	return make([]byte, m.buffSize)
}
