package headers

import (
	"github.com/fakefloordiv/indigo/internal"
	"github.com/fakefloordiv/indigo/settings"
)

type (
	Headers       map[string][]string
	ValueAppender func(b []byte) int
)

// Manager encapsulates all the stuff about keys and values of headers
// For values, it's just a big slice with a lot of smaller slices that
// pointing at their section with a value
type Manager struct {
	Headers         Headers
	Values          []byte
	valueBegin      int
	headersSettings settings.Headers
}

// NewManager constructs a new instance of Manager. It takes only settings,
// underlying headers map is being allocated every time for each request
// because it's faster than cleaning it by hands
func NewManager(settings settings.Headers) Manager {
	defaultValuesBuffSize := uint16(settings.Number.Default) * settings.ValueLength.Default

	return Manager{
		Headers:         make(Headers, settings.Number.Default),
		Values:          make([]byte, 0, defaultValuesBuffSize),
		headersSettings: settings,
	}
}

// BeginValue just updates an offset and returns a bool that signalizes
// whether this one value exceeds the limit of maximal number of headers
func (m *Manager) BeginValue() (exceeded bool) {
	m.valueBegin = len(m.Values)

	return uint8(len(m.Headers)) >= m.headersSettings.Number.Maximal
}

// AppendValue appends a char to values slice and returns bool that
// signalizes whether current value exceeds max header value length
// limit
func (m *Manager) AppendValue(char byte) (exceeded bool) {
	m.Values = append(m.Values, char)

	return uint16(len(m.Values)-m.valueBegin) >= m.headersSettings.ValueLength.Maximal
}

// FinalizeValue just marks that we are done with our header value. It
// takes provided key and adds a new entry into the headers map
func (m Manager) FinalizeValue(key string) (finalValue string) {
	finalValue = internal.B2S(m.Values[m.valueBegin:])

	values, found := m.Headers[key]
	if !found {
		m.Headers[key] = []string{finalValue}
	} else {
		m.Headers[key] = append(values, finalValue)
	}

	return finalValue
}

// Reset resets manager. It just nulls a slice with values and makes
// new headers map
func (m *Manager) Reset() {
	m.Values = m.Values[:0]
	m.Headers = make(Headers, m.headersSettings.Number.Default)
}
