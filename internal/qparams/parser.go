package qparams

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/indigo/internal/urlencoded"
	"github.com/indigo-web/utils/uf"
)

// replace empty value (or so-called parameter without value) with the following string
const defaultEmptyValueContent = "1"

func Into(s *keyvalue.Storage) func(string, string) {
	return func(k string, v string) {
		s.Add(k, v)
	}
}

func Parse(data []byte, cb func(k string, v string)) error {
	var key string

parseKey:
	if len(data) == 0 {
		return nil
	}

	for i := 0; i < len(data); i++ {
		// TODO: must check for illegal characters (e.g. whitespaces, non-printables, unsafe characters)
		switch data[i] {
		case '=':
			decoded, err := urlencoded.Decode(data[:i])
			if err != nil {
				return err
			}
			if len(decoded) == 0 {
				return status.ErrBadRequest
			}

			key = uf.B2S(decoded)
			data = data[i+1:]
			goto parseValue
		case '&':
			decoded, err := urlencoded.Decode(data[:i])
			if err != nil {
				return err
			}
			if len(decoded) == 0 {
				return status.ErrBadRequest
			}

			cb(uf.B2S(decoded), defaultEmptyValueContent)
			data = data[i+1:]
			goto parseKey
		case '+':
			data[i] = ' '
		}
	}

	{
		decoded, err := urlencoded.Decode(data)
		if err != nil {
			return err
		}
		if len(decoded) == 0 {
			return status.ErrBadRequest
		}

		cb(uf.B2S(decoded), defaultEmptyValueContent)
	}

	return nil

parseValue:
	for i := range data {
		switch data[i] {
		case '&':
			decoded, err := urlencoded.Decode(data[:i])
			if err != nil {
				return err
			}
			cb(key, value(decoded))
			data = data[i+1:]
			goto parseKey
		case '+':
			data[i] = ' '
		}
	}

	{
		decoded, err := urlencoded.Decode(data)
		if err != nil {
			return err
		}

		cb(key, value(decoded))
	}

	return nil
}

func value(b []byte) string {
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}

	return uf.B2S(b)
}
