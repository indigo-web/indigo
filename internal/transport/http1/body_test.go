package http1

import (
	"github.com/indigo-web/indigo/http/coding"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/query"
	"github.com/indigo-web/indigo/internal/server/tcp/dummy"
	"github.com/indigo-web/utils/ft"
	"github.com/stretchr/testify/require"
)

func getRequestWithBody(chunked bool, body ...[]byte) (*http.Request, http.Body) {
	client := dummy.NewCircularClient(body...)
	chunkedParser := chunkedbody.NewParser(chunkedbody.DefaultSettings())
	reqBody := NewBody(client, chunkedParser, coding.NewManager(0))

	var (
		contentLength int
		hdrs          *headers.Headers
	)

	if chunked {
		hdrs = headers.FromMap(map[string][]string{
			"Transfer-Encoding": {"chunked"},
		})
	} else {
		// unfortunately, len() cannot be passed as an ordinary function
		length := func(b []byte) int {
			return len(b)
		}
		contentLength = ft.Sum(ft.Map(length, body))
		hdrs = headers.FromMap(map[string][]string{
			"Content-Length": {strconv.Itoa(contentLength)},
		})
	}

	request := http.NewRequest(
		hdrs, query.Query{}, http.NewResponse(), dummy.NewNopConn(), reqBody, nil, false,
	)
	request.ContentLength = contentLength
	request.Encoding.Chunked = chunked

	return request, reqBody
}

func TestBodyReader_Plain(t *testing.T) {
	t.Run("call once", func(t *testing.T) {
		sample := []byte("Hello, world!")
		request, body := getRequestWithBody(false, sample)
		body.Init(request)

		actualBody, err := body.String()
		require.NoError(t, err)
		require.Equal(t, string(sample), actualBody)
	})

	t.Run("multiple calls", func(t *testing.T) {
		sample := [][]byte{
			[]byte("Hel"),
			[]byte("lo, "),
			[]byte("wor"),
			[]byte("ld!"),
		}
		toString := func(b []byte) string {
			return string(b)
		}
		bodyString := ft.Sum(ft.Map(toString, sample))

		request, body := getRequestWithBody(false, sample...)
		body.Init(request)

		actualBody, err := body.String()
		require.NoError(t, err)
		require.Equal(t, bodyString, actualBody)
	})

	t.Run("a lot of data", func(t *testing.T) {
		const buffSize = 10
		var (
			first  = strings.Repeat("a", buffSize)
			second = strings.Repeat("b", buffSize)
		)

		client := dummy.NewCircularClient([]byte(first + second))

		hdrs := headers.NewHeaders()
		request := http.NewRequest(
			hdrs, query.Query{}, http.NewResponse(), dummy.NewNopConn(), nil,
			nil, false,
		)
		request.ContentLength = buffSize
		chunkedParser := chunkedbody.NewParser(chunkedbody.DefaultSettings())
		body := NewBody(client, chunkedParser, coding.NewManager(0))
		body.Init(request)

		data, err := body.Retrieve()
		require.NoError(t, err)
		require.Equal(t, first, string(data))

		data, err = body.Retrieve()
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
	request, body := getRequestWithBody(true, chunked)
	body.Init(request)

	actualBody, err := body.String()
	require.NoError(t, err)
	require.Equal(t, wantBody, actualBody)
}
