package urlencoded

import (
	"bytes"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/hexconv"
	"github.com/indigo-web/utils/uf"
)

// Decode replaces all urlencoded sequences by corresponding ASCII characters into itself.
func Decode(data []byte) ([]byte, error) {
	for i := bytes.IndexByte(data, '%'); i != -1; i = bytes.IndexByte(data, '%') {
		if i >= len(data)-2 {
			return nil, status.ErrURLDecoding
		}

		a, b := hexconv.Halfbyte[data[i+1]], hexconv.Halfbyte[data[i+2]]
		if a|b > 0x0f {
			return nil, status.ErrURLDecoding
		}
		data[i] = (a << 4) | b
		copy(data[i+1:], data[i+3:])
		data = data[:len(data)-2]
	}

	return data, nil
}

// LazyDecode decodes data into the buffer on demand
func LazyDecode(data []byte, buff []byte) (decoded []byte, buffer []byte, err error) {
	percent := bytes.IndexByte(data, '%')
	if percent == -1 {
		return data, buff, nil
	}

	for percent != -1 {
		if percent >= len(data)-2 {
			return nil, buff, status.ErrURLDecoding
		}

		buff = append(buff, data[:percent]...)
		a, b := hexconv.Halfbyte[data[percent+1]], hexconv.Halfbyte[data[percent+2]]
		if a|b > 0x0f {
			return nil, buff, status.ErrURLDecoding
		}

		buff = append(buff, (a<<4)|b)
		data = data[percent+3:]
		percent = bytes.IndexByte(data, '%')
	}

	buff = append(buff, data...)
	return buff, buff, nil
}

func LazyDecodeString(data string, buff []byte) (string, []byte, error) {
	d, buff, err := LazyDecode(uf.S2B(data), buff)
	return uf.B2S(d), buff, err
}
