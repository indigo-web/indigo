package codec

import (
	"bytes"
	"fmt"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/require"
	"io"
	"strconv"
	"strings"
	"testing"
)

func gzipped(text string) []byte {
	buff := bytes.NewBuffer(nil)
	c := gzip.NewWriter(buff)
	_, err := c.Write([]byte(text))
	if err != nil {
		panic("unexpected error during gzipping")
	}
	if c.Close() != nil {
		panic("unexpected error during closing gzip writer")
	}

	return buff.Bytes()
}

func gunzip(gzipped ...[]byte) (string, error) {
	dc := NewGZIP().New()
	err := dc.ResetDecompressor(dummy.NewClient(gzipped...).Once())
	if err != nil {
		return "", err
	}

	return fetchAll(dc)
}

func fetchAll(source http.Fetcher) (string, error) {
	builder := strings.Builder{}

	for {
		data, err := source.Fetch()
		builder.Write(data)
		switch err {
		case nil:
		case io.EOF:
			return builder.String(), nil
		default:
			return "", err
		}
	}
}

func TestGZIP(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		text, err := gunzip(gzipped("Hello, world!"))
		require.NoError(t, err)
		require.Equal(t, "Hello, world!", text)
	})

	t.Run("scattered", func(t *testing.T) {
		text := strings.Repeat("Hello, world! Lorem ipsum! ", 100)
		scattered := scatter(gzipped(text), 2)
		result, err := gunzip(scattered...)
		require.NoError(t, err)
		require.Equal(t, text, result)
	})

	res, err := gunzip([]byte("\x1f\x8b\b\x00\x00\x00\x00\x00\x00\x03\xf3H\xcd\xc9\xc9\xd7Q(\xcf/\xcaI\xe1\x02\x00\xa6?ZG\r\x00\x00\x00"))
	fmt.Println("result:", err, strconv.Quote(res))
}

func scatter(b []byte, step int) (pieces [][]byte) {
	for i := 0; i < len(b); i += step {
		pieces = append(pieces, b[i:min(i+step, len(b))])
	}

	return pieces
}
