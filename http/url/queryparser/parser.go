package queryparser

import (
	"github.com/fakefloordiv/indigo/errors"
	"github.com/fakefloordiv/indigo/internal"
)

func Parse(data []byte, queryMapFactory func() map[string][]byte) (queries map[string][]byte, err error) {
	var (
		offset int
		key    string
	)

	state := eKey
	queries = queryMapFactory()

	for i := range data {
		switch state {
		case eKey:
			if data[i] == '=' {
				key = internal.B2S(data[offset:i])
				if len(key) == 0 {
					return nil, errors.ErrBadQuery
				}

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
