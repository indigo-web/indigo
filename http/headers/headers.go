package headers

import "indigo/settings"

type (
	Headers       map[string][]byte
	ValueAppender func(b []byte) int
)

type Manager struct {
	Headers       Headers
	Values        []byte
	valueBegin    int
	headersNumber settings.HeadersNumber
}

func NewManager(headers Headers, settings settings.Headers) Manager {
	defaultValuesBuffSize := uint16(settings.Number.Default) * settings.ValueLength.Default

	return Manager{
		// TODO: update Default value
		Headers:       headers,
		Values:        make([]byte, 0, defaultValuesBuffSize),
		headersNumber: settings.Number,
	}
}

func (m *Manager) BeginValue() (oversize bool) {
	m.valueBegin = len(m.Values)

	return uint8(len(m.Headers)) >= m.headersNumber.Maximal
}

func (m Manager) FinalizeValue(key string) (finalValue []byte) {
	finalValue = m.Values[m.valueBegin:]
	m.Headers[key] = finalValue

	return finalValue
}

func (m *Manager) Reset() {
	m.Values = m.Values[:0]
	m.Headers = make(map[string][]byte, m.headersNumber.Default)
}
