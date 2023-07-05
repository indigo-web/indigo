package http1

import (
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/decode"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/indigo/settings"
	"github.com/indigo-web/utils/ft"
	"github.com/stretchr/testify/require"
)

func getRequestWithReader(chunked bool, body ...[]byte) (*http.Request, http.BodyReader) {
	client := dummy.NewCircularClient(body...)
	reader := NewBodyReader(client, settings.Default().Body)
	reqBody := http.NewBody(reader, decode.NewDecoder())

	var (
		contentLength int
		hdrs          *headers.Headers
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
		contentLength = ft.Sum(ft.Map(length, body))
		hdrs = headers.NewHeaders(map[string][]string{
			"Content-Length": {strconv.Itoa(contentLength)},
		})
	}

	request := http.NewRequest(
		hdrs, query.Query{}, http.NewResponse(), dummy.NewNopConn(), reqBody, nil, false,
	)
	request.ContentLength = contentLength
	request.TransferEncoding.Chunked = chunked

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
		bodyString := ft.Sum(ft.Map(toString, body))

		request, reader := getRequestWithReader(false, body...)
		reader.Init(request)

		actualBody, err := readFullBody(reader)
		require.NoError(t, err)
		require.Equal(t, bodyString, string(actualBody))
	})

	t.Run("ALotOfData", func(t *testing.T) {
		const buffSize = 10
		var (
			first  = strings.Repeat("a", buffSize)
			second = strings.Repeat("b", buffSize)
		)

		client := dummy.NewCircularClient([]byte(first + second))

		hdrs := headers.NewHeaders(nil)
		request := http.NewRequest(
			hdrs, query.Query{}, http.NewResponse(), dummy.NewNopConn(), nil, nil, false,
		)
		request.ContentLength = buffSize
		reader := NewBodyReader(client, settings.Default().Body)
		reader.Init(request)

		data, err := reader.Read()
		require.NoError(t, err)
		require.Equal(t, first, string(data))

		data, err = reader.Read()
		require.EqualError(t, err, io.EOF.Error())
		require.Empty(t, data)

		data, err = client.Read()
		require.NoError(t, err)
		require.Equal(t, second, string(data))
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
