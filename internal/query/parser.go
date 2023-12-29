package query

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/datastruct"
	"github.com/indigo-web/indigo/internal/uridecode"
	"github.com/indigo-web/utils/uf"
)

// replace empty value (or so-called parameter without value) with the following string
const defaultEmptyValueContent = "1"

func Parse(data []byte, params *datastruct.KeyValue) (err error) {
	if len(data) == 0 {
		return nil
	}

	data, err = uridecode.Decode(data, data[:0])
	if err != nil {
		return err
	}

	var key string

parseKey:
	if len(data) == 0 {
		return nil
	}

	for i := range data {
		switch data[i] {
		case '=':
			key = uf.B2S(data[:i])
			if len(key) == 0 {
				return status.ErrBadQuery
			}

			data = data[i+1:]
			goto parseValue
		case '&':
			params.Add(uf.B2S(data[:i]), defaultEmptyValueContent)
			data = data[i+1:]
			goto parseKey
		case '+':
			data[i] = ' '
		}
	}

	params.Add(uf.B2S(data), defaultEmptyValueContent)

	return nil

parseValue:
	for i := range data {
		switch data[i] {
		case '&':
			// ignore the fact that the value may be empty. In case there's no equal sign, we
			// definitely know it's a flag and is supposed to be like that. However, if there's
			// an equal mark, we cannot be sure it wasn't on purpose. So just let a user judge
			params.Add(key, uf.B2S(data[:i]))
			data = data[i+1:]
			goto parseKey
		case '+':
			data[i] = ' '
		}
	}

	params.Add(key, uf.B2S(data))

	return nil
}
