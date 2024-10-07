package urlencoded

import (
	"bytes"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/hexconv"
)

// Decode replaces all urlencoded sequences by corresponding ASCII characters
func Decode(data []byte) ([]byte, error) {
	for i := bytes.IndexByte(data, '%'); i != -1; i = bytes.IndexByte(data, '%') {
		if i >= len(data)-2 {
			return nil, status.ErrURLDecoding
		}

		a := hexconv.Parse(data[i+1])
		b := hexconv.Parse(data[i+2])
		if a > 0xf || b > 0xf {
			return nil, status.ErrURLDecoding
		}
		data[i] = (a << 4) | b
		copy(data[i+1:], data[i+3:])
		data = data[:len(data)-2]
	}

	return data, nil
}
