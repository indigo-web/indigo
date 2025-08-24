package codec

import (
	"io"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/transport/dummy"
	"github.com/stretchr/testify/require"
)

func compress(inst Instance, text string) []byte {
	loopback := dummy.NewMockClient().Journaling()
	inst.ResetCompressor(loopback)

	if _, err := inst.Write([]byte(text)); err != nil {
		panic(err)
	}

	if err := inst.Close(); err != nil {
		panic(err)
	}

	return loopback.Written()
}

func decompress(inst Instance, data ...[]byte) (string, error) {
	if err := inst.ResetDecompressor(dummy.NewMockClient(data...)); err != nil {
		return "", err
	}

	return fetchAll(inst)
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

func testCodec(t *testing.T, inst Instance) {
	t.Run("identity", func(t *testing.T) {
		result, err := decompress(inst, compress(inst, "Hello, world!"))
		require.NoError(t, err)
		require.Equal(t, "Hello, world!", result)
	})

	t.Run("stream", func(t *testing.T) {
		text := strings.Repeat("Hello, world! Lorem ipsum! ", 100)
		scattered := scatter(compress(inst, text), 2)
		result, err := decompress(inst, scattered...)
		require.NoError(t, err)
		require.Equal(t, text, result)
	})
}

func TestGZIP(t *testing.T) {
	testCodec(t, NewGZIP().New())
}
