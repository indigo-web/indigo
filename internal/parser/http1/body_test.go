package http1

import (
	"io"
	"strconv"
	"testing"

	"github.com/indigo-web/indigo/internal/server/tcp/dummy"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/functools"
	"github.com/indigo-web/indigo/settings"
	"github.com/stretchr/testify/require"
)

func getRequestWithReader(chunked bool, body ...[]byte) (*http.Request, http.BodyReader) {
	client := dummy.NewCircularClient(body...)
	reader := NewBodyReader(client, settings.Default().Body)

	var (
		contentLength int
		hdrs          headers.Headers
	)

	if chunked {
		hdrs = headers.NewHeaders(map[string][]string{
			"Transfer-Encoding": {"chunked"},
		})
	} else {
		// unfortunately, len() cannot be passed as an ordinary function
		length := func(b []byte) int {
			return len(b)
		}
		contentLength = functools.Sum(functools.Map(length, body))
		hdrs = headers.NewHeaders(map[string][]string{
			"Content-Length": {strconv.Itoa(contentLength)},
		})
	}

	request := http.NewRequest(
		hdrs, query.Query{}, http.NewResponse(), dummy.NewNopConn(), reader, false,
	)
	request.ContentLength = contentLength
	request.IsChunked = chunked

	return request, reader
}

func readFullBody(reader http.BodyReader) (body []byte, err error) {
	for {
		piece, err := reader.Read()
		switch err {
		case nil:
			body = append(body, piece...)
		case io.EOF:
			return body, nil
		default:
			return nil, err
		}
	}
}

func TestBodyReader_Plain(t *testing.T) {
	t.Run("CallOnce", func(t *testing.T) {
		body := []byte("Hello, world!")
		request, reader := getRequestWithReader(false, body)
		reader.Init(request)

		actualBody, err := readFullBody(reader)
		require.NoError(t, err)
		require.Equal(t, string(body), string(actualBody))
	})

	t.Run("CallMultipleTimes", func(t *testing.T) {
		body := [][]byte{
			[]byte("Hel"),
			[]byte("lo, "),
			[]byte("wor"),
			[]byte("ld!"),
		}
		toString := func(b []byte) string {
			return string(b)
		}
		bodyString := functools.Sum(functools.Map(toString, body))

		request, reader := getRequestWithReader(false, body...)
		reader.Init(request)

		actualBody, err := readFullBody(reader)
		require.NoError(t, err)
		require.Equal(t, bodyString, string(actualBody))
	})
}

func TestBodyReader_Chunked(t *testing.T) {
	chunked := []byte("7\r\nMozilla\r\n9\r\nDeveloper\r\n7\r\nNetwork\r\n0\r\n\r\n")
	wantBody := "MozillaDeveloperNetwork"
	request, reader := getRequestWithReader(true, chunked)
	reader.Init(request)

	actualBody, err := readFullBody(reader)
	require.NoError(t, err)
	require.Equal(t, wantBody, string(actualBody))
}
