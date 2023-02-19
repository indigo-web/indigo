package types

type (
	DefaultHeader struct {
		Value string
		Seen  bool
	}

	HeadersMap map[string]*DefaultHeader
)
