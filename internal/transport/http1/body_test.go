package http1

import (
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/settings"
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

func getRequestWithBody(chunked bool, body ...[]byte) (*http.Request, *Body) {
	client := dummy.NewCircularClient(body...)
	chunkedParser := chunkedbody.NewParser(chunkedbody.DefaultSettings())
	reqBody := NewBody(client, chunkedParser, settings.Default().Body)

	var (
		contentLength int
		hdrs          headers.Headers
	)

	if chunked {
		hdrs = headers.NewFromMap(map[string][]string{
			"Transfer-Encoding": {"chunked"},
		})
	} else {
		// unfortunately, len() cannot be passed as an ordinary function
		length := func(b []byte) int {
			return len(b)
		}
		contentLength = ft.Sum(ft.Map(length, body))
		hdrs = headers.NewFromMap(map[string][]string{
			"Content-Length": {strconv.Itoa(contentLength)},
		})
	}

	request := http.NewRequest(
		hdrs, new(query.Query), http.NewResponse(), dummy.NewNopConn(), reqBody, nil,
	)
	request.ContentLength = contentLength
	request.Encoding.Chunked = chunked
	reqBody.Init(request)

	return request, reqBody
}

func TestBodyReader_Plain(t *testing.T) {
	t.Run("call once", func(t *testing.T) {
		sample := []byte("Hello, world!")
		_, body := getRequestWithBody(false, sample)
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

		_, body := getRequestWithBody(false, sample...)
		actualBody, err := body.String()
		require.NoError(t, err)
		require.Equal(t, bodyString, actualBody)

		actualBody, err = body.String()
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

		hdrs := headers.New()
		request := http.NewRequest(
			hdrs, new(query.Query), http.NewResponse(), dummy.NewNopConn(), nil, nil,
		)
		request.ContentLength = buffSize
		chunkedParser := chunkedbody.NewParser(chunkedbody.DefaultSettings())
		body := NewBody(client, chunkedParser, settings.Default().Body)
		body.Init(request)

		data, err := body.Retrieve()
		require.Equal(t, first, string(data))
		require.EqualError(t, err, io.EOF.Error())

		data, err = body.Retrieve()
		require.Empty(t, data)
		require.EqualError(t, err, io.EOF.Error())

		data, err = client.Read()
		require.NoError(t, err)
		require.Equal(t, second, string(data))
	})

	t.Run("reader", func(t *testing.T) {
		data := "qwertyuiopasdfghjklzxcvbnm"
		_, body := getRequestWithBody(false, []byte(data))
		result := make([]byte, 0, len(data))
		buff := make([]byte, 1)

		for {
			n, err := body.Read(buff)
			result = append(result, buff[:n]...)
			if err == io.EOF {
				break
			}

			require.NoError(t, err)
		}

		require.Equal(t, data, string(result))
	})

	t.Run("too big plain body", func(t *testing.T) {
		data := strings.Repeat("a", 10)
		request, _ := getRequestWithBody(false, []byte(data))
		client := dummy.NewCircularClient([]byte(data))
		chunkedParser := chunkedbody.NewParser(chunkedbody.DefaultSettings())
		s := settings.Default().Body
		s.MaxSize = 9
		body := NewBody(client, chunkedParser, s)
		body.Init(request)

		_, err := body.Bytes()
		require.EqualError(t, err, status.ErrBodyTooLarge.Error())
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
