package qparams

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"github.com/indigo-web/utils/uf"
)

func Into(s *keyvalue.Storage) func(string, string) {
	return func(k string, v string) {
		s.Add(k, v)
	}
}

type (
	CB      = func(k string, v string)
	Decoder = func(src, dst []byte) (decoded, buffer []byte, err error)
)

func Parse(data, buff []byte, cb CB, decoder Decoder, defFlagValue string) (buffer []byte, err error) {
	// TODO: performance can be improved considerably by decoding manually in the loop
	var key string

parseKey:
	if len(data) == 0 {
		return buff, nil
	}

	var decoded []byte

	for i := 0; i < len(data); i++ {
		c := data[i]
		switch c {
		case '=':
			decoded, buff, err = decoder(data[:i], buff)
			if err != nil {
				return buff, err
			}
			if len(decoded) == 0 {
				return buff, status.ErrBadRequest
			}

			key = uf.B2S(decoded)
			data = data[i+1:]
			goto parseValue
		case '&':
			decoded, buff, err = decoder(data[:i], buff)
			if err != nil {
				return buff, err
			}
			if len(decoded) == 0 {
				return buff, status.ErrBadRequest
			}

			cb(uf.B2S(decoded), defFlagValue)
			data = data[i+1:]
			goto parseKey
		}

		if illegalSymbol(c) {
			// exclude all non-printable characters and whitespaces
			return buff, status.ErrBadRequest
		}
	}

	if containsIllegalSymbol(data) {
		return buff, status.ErrBadRequest
	}

	decoded, buff, err = decoder(data, buff)
	if err != nil {
		return buff, err
	}
	if len(decoded) == 0 {
		return buff, status.ErrBadRequest
	}

	cb(uf.B2S(decoded), defFlagValue)

	return buff, nil

parseValue:
	for i, c := range data {
		if c == '&' {
			decoded, buff, err = decoder(data[:i], buff)
			if err != nil {
				return buff, err
			}

			cb(key, value(decoded))
			data = data[i+1:]
			goto parseKey
		} else if illegalSymbol(c) {
			return buff, status.ErrBadRequest
		}
	}

	if containsIllegalSymbol(data) {
		return buff, status.ErrBadRequest
	}

	decoded, buff, err = decoder(data, buff)
	if err != nil {
		return buff, err
	}

	cb(key, value(decoded))

	return buff, nil
}

func containsIllegalSymbol(data []byte) bool {
	for _, c := range data {
		if illegalSymbol(c) {
			return true
		}
	}

	return false
}

func illegalSymbol(c byte) bool {
	return c < 0x21 || c > 0x7e
}

func value(b []byte) string {
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}

	return uf.B2S(b)
}
