package headers

import "indigo/settings"

type (
	Headers       map[string][]byte
	ValueAppender func(b []byte) int
)

type Manager struct {
	Headers    Headers
	Values     []byte
	valueBegin int
	maxHeaders uint8
}

func NewManager(headers settings.Headers) Manager {
	defaultValuesBuffSize := uint16(headers.Number.Default) * headers.ValueLength.Default

	return Manager{
		// TODO: update Default value
		Headers:    make(map[string][]byte, headers.Number.Default),
		Values:     make([]byte, 0, defaultValuesBuffSize),
		maxHeaders: headers.Number.Maximal,
	}
}

func (m *Manager) BeginValue() (oversize bool) {
	m.valueBegin = len(m.Values)

	return uint8(len(m.Headers)) >= m.maxHeaders
}

func (m *Manager) FinalizeValue(key string) (finalValue []byte) {
	finalValue = m.Values[m.valueBegin:]
	m.Headers[key] = finalValue

	return finalValue
}

func (m *Manager) Reset() {
	m.Values = m.Values[:0]
}
