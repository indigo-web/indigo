package queryparser

import (
	"indigo/errors"
	"indigo/internal"
)

const (
	// yes, this is a bad design. I don't know how to pass settings here
	// so, TODO: queries must be initialized in indi.go
	defaultQueriesLength = 5
)

func Parse(data []byte) (queries map[string][]byte, err error) {
	var (
		offset int
		key    string
	)

	state := eKey
	queries = make(map[string][]byte, defaultQueriesLength)

	for i := range data {
		switch state {
		case eKey:
			if data[i] == '=' {
				key = internal.B2S(data[offset:i])
				offset = i + 1
				state = eValue
			}
		case eValue:
			if data[i] == '&' {
				queries[key] = data[offset:i]
				offset = i + 1
				state = eKey
			}
		}
	}

	if state == eKey {
		return nil, errors.ErrBadQuery
	}

	queries[key] = data[offset:]

	return queries, nil
}
