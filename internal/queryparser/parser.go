package queryparser

import (
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/utils/uf"
)

func Parse(data []byte, query *headers.Headers) (err error) {
	var (
		offset int
		key    string
	)

	state := eKey

	if len(data) == 0 {
		return nil
	}

	for i := range data {
		switch state {
		case eKey:
			switch data[i] {
			case '=':
				key = uf.B2S(data[offset:i])
				if len(key) == 0 {
					return status.ErrBadQuery
				}

				offset = i + 1
				state = eValue
			case '+':
				data[i] = ' '
			case '%':
				if len(data) <= i+2 {
					return status.ErrBadQuery
				}

				// TODO: apply url decoding here, too
				return status.ErrBadQuery
			}
		case eValue:
			switch data[i] {
			case '&':
				query.Add(key, uf.B2S(data[offset:i]))
				offset = i + 1
				state = eKey
			case '+':
				data[i] = ' '
			case '%':
				if len(data) <= i+2 {
					return status.ErrBadQuery
				}

				// TODO: apply url decoding here, too
				return status.ErrBadQuery
			}
		}
	}

	if state == eKey {
		return status.ErrBadQuery
	}

	query.Add(key, uf.B2S(data[offset:]))

	return nil
}
