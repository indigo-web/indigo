package http1

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/transport/dummy"
	"io"
	"strconv"
	"strings"
	"testing"

	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/utils/ft"
	"github.com/stretchr/testify/require"
)

func getRequestWithBody(chunked bool, body ...[]byte) (*http.Request, *Body) {
	client := dummy.NewCircularClient(body...).OneTime()
	chunkedParser := chunkedbody.NewParser(chunkedbody.DefaultSettings())
	reqBody := NewBody(client, chunkedParser, config.Default().Body)

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

	request := construct.Request(config.Default(), dummy.NewNopClient(), reqBody)
	request.Headers = hdrs
	request.ContentLength = contentLength
	request.Encoding.Chunked = chunked
	reqBody.Reset(request)

	return request, reqBody
}

func readall(body *Body) ([]byte, error) {
	var buff []byte

	for {
		data, err := body.Retrieve()
		buff = append(buff, data...)
		switch err {
		case nil:
		case io.EOF:
			return buff, nil
		default:
			return buff, err
		}
	}
}

func TestBodyReader_Plain(t *testing.T) {
	t.Run("all at once", func(t *testing.T) {
		sample := []byte("Hello, world!")
		_, body := getRequestWithBody(false, sample)
		actualBody, err := body.Retrieve()
		require.EqualError(t, err, io.EOF.Error())
		require.Equal(t, string(sample), string(actualBody))
	})

	t.Run("consecutive data pieces", func(t *testing.T) {
		sample := [][]byte{
			[]byte("Hel"),
			[]byte("lo, "),
			[]byte("wor"),
			[]byte("ld!"),
		}
		bodyString := "Hello, world!"

		_, body := getRequestWithBody(false, sample...)
		actualBody, err := readall(body)
		require.NoError(t, err)
		require.Equal(t, bodyString, string(actualBody))
	})

	t.Run("distinction", func(t *testing.T) {
		const buffSize = 10
		var (
			first  = strings.Repeat("a", buffSize)
			second = strings.Repeat("b", buffSize)
		)

		client := dummy.NewCircularClient([]byte(first + second))
		request := construct.Request(config.Default(), dummy.NewNopClient(), nil)
		request.ContentLength = buffSize
		chunkedParser := chunkedbody.NewParser(chunkedbody.DefaultSettings())
		body := NewBody(client, chunkedParser, config.Default().Body)
		body.Reset(request)

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

	t.Run("too big plain body", func(t *testing.T) {
		data := strings.Repeat("a", 10)
		request, _ := getRequestWithBody(false, []byte(data))
		client := dummy.NewCircularClient([]byte(data))
		chunkedParser := chunkedbody.NewParser(chunkedbody.DefaultSettings())
		s := config.Default().Body
		s.MaxSize = 9
		body := NewBody(client, chunkedParser, s)
		body.Reset(request)

		_, err := readall(body)
		require.EqualError(t, err, status.ErrBodyTooLarge.Error())
	})
}

func TestBodyReader_Chunked(t *testing.T) {
	chunked := []byte("7\r\nMozilla\r\n9\r\nDeveloper\r\n7\r\nNetwork\r\n0\r\n\r\n")
	wantBody := "MozillaDeveloperNetwork"
	request, body := getRequestWithBody(true, chunked)
	body.Reset(request)

	actualBody, err := readall(body)
	require.NoError(t, err)
	require.Equal(t, wantBody, string(actualBody))
}

func TestBodyReader_TillEOF(t *testing.T) {
	request, body := getRequestWithBody(false, []byte("Hello, "), []byte("world!"))
	request.Connection = "close"
	body.Reset(request)

	actualBody, err := readall(body)
	require.NoError(t, err)
	require.Equal(t, "Hello, world!", string(actualBody))
}
