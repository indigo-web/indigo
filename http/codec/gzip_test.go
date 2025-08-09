package codec

import (
	"io"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
)

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
}

func gzipped(text string) []byte {
	c := NewGZIP().New()
	sinkhole := dummy.NewMockClient().Journaling()
	c.ResetCompressor(sinkhole)

	if _, err := c.Write([]byte(text)); err != nil {
		panic(err)
	}

	if err := c.Close(); err != nil {
		panic(err)
	}

	return sinkhole.Written()
}

func gunzip(gzipped ...[]byte) (string, error) {
	dc := NewGZIP().New()
	err := dc.ResetDecompressor(dummy.NewMockClient(gzipped...))
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

func scatter(b []byte, step int) (pieces [][]byte) {
	for i := 0; i < len(b); i += step {
		pieces = append(pieces, b[i:min(i+step, len(b))])
	}

	return pieces
}
