package decoder

import "github.com/indigo-web/indigo/http/status"

type (
	Decoder interface {
		Decode(input []byte) (output []byte, err error)
	}
	Constructor func(buffer []byte) Decoder
)

// Manager manages all the attached decoders. In case there are no decoders, original data
// will be returned. Manager MUST be per-client, not per-instance
type Manager struct {
	decoders map[string]Decoder
	buffSize int64
}

func NewManager(buffSize int64) *Manager {
	return &Manager{
		decoders: make(map[string]Decoder),
		buffSize: buffSize,
	}
}

func (m *Manager) Add(token string, constructor Constructor) {
	decoder := constructor(make([]byte, m.buffSize))
	m.decoders[token] = decoder

	// this exists in backward-capability purposes. Some old clients may use x-gzip or
	// x-compress instead of regular gzip or compress tokens respectively.
	// see https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Content-Encoding#directives
	switch token {
	case "gzip":
		m.decoders[token] = decoder
	case "compress":
		m.decoders[token] = decoder
	}
}

// identity stands for "no attached decoders"
const identity = "identity"

func (m *Manager) Decode(token string, encoded []byte) (decoded []byte, err error) {
	if len(token) == 0 || token == identity {
		return encoded, nil
	}

	decoder, found := m.decoders[token]
	if !found {
		return nil, status.ErrUnsupportedEncoding
	}

	return decoder.Decode(encoded)
}
