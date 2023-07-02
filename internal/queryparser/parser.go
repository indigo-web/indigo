package queryparser

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/utils/uf"
)

func Parse(data []byte, queryMapFactory func() map[string]string) (queries map[string]string, err error) {
	// TODO: make queryMapFactory map[string][]string, just like headers

	var (
		offset int
		key    string
	)

	state := eKey
	queries = queryMapFactory()

	if len(data) == 0 {
		return queries, nil
	}

	for i := range data {
		switch state {
		case eKey:
			if data[i] == '=' {
				key = uf.B2S(data[offset:i])
				if len(key) == 0 {
					return nil, status.ErrBadQuery
				}

				offset = i + 1
				state = eValue
			}
		case eValue:
			if data[i] == '&' {
				queries[key] = uf.B2S(data[offset:i])
				offset = i + 1
				state = eKey
			}
		}
	}

	if state == eKey {
		return nil, status.ErrBadQuery
	}

	queries[key] = uf.B2S(data[offset:])

	return queries, nil
}
