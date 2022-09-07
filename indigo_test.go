package indigo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	stdhttp "net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fakefloordiv/indigo/settings"

	"github.com/fakefloordiv/indigo/http"

	methods "github.com/fakefloordiv/indigo/http/method"
	"github.com/fakefloordiv/indigo/http/proto"
	"github.com/fakefloordiv/indigo/router"
	"github.com/fakefloordiv/indigo/router/inbuilt"
	"github.com/fakefloordiv/indigo/types"

	"github.com/stretchr/testify/require"
)

const (
	addr = "localhost:16100"
	URL  = "http://" + addr
)

const (
	testHeaderKey   = "hello"
	testHeaderValue = "World!"

	testQueryKey   = "hel lo"
	testQueryValue = "wor ld"

	testRequestBody = "Hello, world!"

	testFilename       = "tests/index.html"
	testFileIfNotFound = "404 not found"
)

func instantlyDisconnect() {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}

	_ = conn.Close()
}

func readN(conn net.Conn, n int) ([]byte, error) {
	receivedBuff := make([]byte, 0, n)
	buff := make([]byte, n)

	for {
		recvd, err := conn.Read(buff)
		if err != nil {
			return nil, err
		}

		receivedBuff = append(receivedBuff, buff[:recvd]...)

		if len(receivedBuff) == n {
			return receivedBuff, nil
		} else if len(receivedBuff) > n {
			return nil, errors.New("received too much data")
		}
	}
}

func getRouter(t *testing.T) router.Router {
	r := inbuilt.NewRouter()

	r.Get("/simple-get", func(request *types.Request) types.Response {
		require.Equal(t, methods.GET, request.Method)
		_, err := request.Query.Get("some non-existing query key")
		require.Error(t, err)
		require.Empty(t, request.Fragment)
		require.Equal(t, proto.HTTP11, request.Proto)

		return types.WithResponse
	})

	r.Get("/get-resp-body", func(_ *types.Request) types.Response {
		return types.WithResponse.WithBody(testRequestBody)
	})

	r.Get("/get-read-body", func(request *types.Request) types.Response {
		require.Contains(t, request.Headers, testHeaderKey)
		requestHeader := strings.Join(request.Headers[testHeaderKey], ",")
		require.Equal(t, testHeaderValue, requestHeader)

		body, err := request.Body()
		require.NoError(t, err)
		require.Empty(t, body)

		return types.WithResponse
	})

	with := r.Group("/with-")

	with.Get("query", func(request *types.Request) types.Response {
		value, err := request.Query.Get(testQueryKey)
		require.NoError(t, err)
		require.Equal(t, testQueryValue, string(value))

		return types.WithResponse
	})

	// request.OnBody() is not tested because request.Body() (wrapper for OnBody)
	// is tested, that is enough to make sure that requestbody works correct

	r.Post("/read-body", func(request *types.Request) types.Response {
		body, err := request.Body()
		require.NoError(t, err)
		require.Equal(t, testRequestBody, string(body))

		return types.WithResponse
	})

	r.Post("/do-not-read-body", func(request *types.Request) types.Response {
		return types.WithResponse
	})

	r.Get("/hijack-conn-no-body-read", func(request *types.Request) types.Response {
		conn, err := request.Hijack()
		require.NoError(t, err)

		// just to notify client that we are ready for receiving something
		_, _ = conn.Write([]byte("a"))

		data, err := readN(conn, len(testRequestBody))
		require.NoError(t, err)
		require.Equal(t, testRequestBody, string(data))

		_ = conn.Close()

		return types.WithResponse
	})

	r.Get("/hijack-conn-with-body-read", func(request *types.Request) types.Response {
		_, _ = request.Body()

		conn, err := request.Hijack()
		require.NoError(t, err)

		_, _ = conn.Write([]byte("a"))

		data, err := readN(conn, len(testRequestBody))
		require.NoError(t, err)
		require.Equal(t, testRequestBody, string(data))

		_ = conn.Close()

		return types.WithResponse
	})

	r.Get("/with-file", func(request *types.Request) types.Response {
		return types.WithResponse.WithFile(testFilename, func(err error) types.Response {
			t.Fail() // this callback must never be called

			return types.WithResponse
		})
	})

	r.Get("/with-file-notfound", func(request *types.Request) types.Response {
		return types.WithResponse.WithFile(testFilename+"notfound", func(err error) types.Response {
			return types.WithResponse.WithBody(testFileIfNotFound)
		})
	})

	return r
}

func TestAllCases(t *testing.T) {
	// testing everything in one function only because we do not wanna
	// run server multiple times (each time takes a bit time, in sum it
	// is a lot of time)

	r := getRouter(t)
	s := settings.Default()
	s.TCPServer.IDLEConnLifetime = 2
	app := NewApp(addr)

	shutdown := make(chan bool)

	go func() {
		require.Equal(t, http.ErrShutdown, app.Serve(r, s))
		shutdown <- true
	}()

	defer func() {
		go app.Shutdown()
		instantlyDisconnect()

		select {
		case <-shutdown:
		case <-time.After(5 * time.Second):
			require.Fail(t, "server is not shutting down for too long")
		}
	}()

	// wait a bit to guarantee that server is running
	time.Sleep(200 * time.Millisecond)

	t.Run("/simple-get", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(URL + "/simple-get")
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		require.Equal(t, "200 "+stdhttp.StatusText(stdhttp.StatusOK), resp.Status)
		require.Equal(t, "HTTP/1.1", resp.Proto)
		require.Contains(t, resp.Header, "Server")
		require.Equal(t, 1, len(resp.Header["Server"]))
		require.Equal(t, "indigo", resp.Header["Server"][0])
	})

	t.Run("/get-resp-body", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(URL + "/get-resp-body")
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, testRequestBody, string(body))
	})

	t.Run("/head", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Head(URL + "/get-resp-body")
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		require.Equal(t, "200 "+stdhttp.StatusText(stdhttp.StatusOK), resp.Status)
		require.Equal(t, "HTTP/1.1", resp.Proto)
		require.Contains(t, resp.Header, "Server")
		require.Equal(t, 1, len(resp.Header["Server"]))
		require.Equal(t, "indigo", resp.Header["Server"][0])

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, string(body))
	})

	t.Run("/get-read-body", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodGet,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/get-read-body",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: stdhttp.Header{
				"Hello": {"World!"},
			},
			Host:       addr,
			RemoteAddr: addr,
		}
		resp, err := stdhttp.DefaultClient.Do(request)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	})

	t.Run("/with-query", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(URL + "/with-query?hel+lo=wor+ld")
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	})

	t.Run("/read-body", func(t *testing.T) {
		body := new(bytes.Buffer)
		body.Write([]byte(testRequestBody))
		resp, err := stdhttp.DefaultClient.Post(URL+"/read-body", "", body)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	})

	t.Run("/do-not-read-body", func(t *testing.T) {
		body := new(bytes.Buffer)
		body.Write([]byte(testRequestBody))
		resp, err := stdhttp.DefaultClient.Post(URL+"/read-body", "", body)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	})

	t.Run("/hijack-conn-no-body-read", func(t *testing.T) {
		sendSimpleRequest(t, "/hijack-conn-no-body-read")
	})

	t.Run("/hijack-conn-with-body-read", func(t *testing.T) {
		sendSimpleRequest(t, "/hijack-conn-with-body-read")
	})

	t.Run("/with-file", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(URL + "/with-file")
		require.NoError(t, err)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		actualContent, err := os.ReadFile(testFilename)
		require.Equal(t, string(actualContent), string(data))
	})

	t.Run("/with-file-notfound", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(URL + "/with-file-notfound")
		require.NoError(t, err)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, testFileIfNotFound, string(data))
	})

	t.Run("/test-idle-disconnect", func(t *testing.T) {
		conn, err := net.Dial("tcp4", addr)
		require.NoError(t, err)

		buff := make([]byte, 10)
		n, err := conn.Read(buff)
		require.Zero(t, n)
		require.EqualError(t, err, io.EOF.Error())
	})

	t.Run("/trace", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodTrace,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/trace",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: stdhttp.Header{
				"Hello": {"World!"},
			},
			Host:       addr,
			RemoteAddr: addr,
		}
		resp, err := stdhttp.DefaultClient.Do(request)
		require.NoError(t, err)

		require.Contains(t, resp.Header, "Content-Type")
		require.Equal(t, 1, len(resp.Header["Content-Type"]), "too many content-type values")
		require.Equal(t, "message/http", resp.Header["Content-Type"][0])

		dataBytes, err := io.ReadAll(resp.Body)
		data := string(dataBytes)
		require.NoError(t, err)

		wantRequestLine := "TRACE /trace HTTP/1.1\r\n"
		require.True(t, len(data) > len(wantRequestLine))
		require.True(t, data[:len(wantRequestLine)] == wantRequestLine)

		headerLines := strings.Split(data[len(wantRequestLine):], "\r\n")
		// request is terminated with \r\n\r\n, so 2 last values in headerLines
		// are empty strings. Remove them
		headerLines = headerLines[:len(headerLines)-2]
		wantHeaderLines := []string{
			"hello: World!",
			"host: " + addr,
			"user-agent: Go-http-client/1.1",
			"accept-encoding: gzip",
		}

		require.Equal(t, len(wantHeaderLines), len(headerLines))

		for _, line := range headerLines {
			require.True(t, contains(wantHeaderLines, line), "unwanted header line: "+line)
		}
	})

	stdhttp.DefaultClient.CloseIdleConnections()

	// at this point server is supposed to be closed, but who knows, who knows
}

func sendSimpleRequest(t *testing.T, path string) {
	conn, err := net.Dial("tcp", addr)
	require.NoError(t, err)

	request := fmt.Sprintf("GET %s HTTP/1.1\r\n\r\n", path)

	_, err = conn.Write([]byte(request))
	require.NoError(t, err)
	buff := make([]byte, 1)
	_, err = conn.Read(buff)
	require.NoError(t, err)
	require.Equal(t, buff, []byte{'a'})
	_, err = conn.Write([]byte(testRequestBody))
	require.NoError(t, err)

	_ = conn.Close()
}

func contains(strs []string, substr string) bool {
	for _, str := range strs {
		if str == substr {
			return true
		}
	}

	return false
}
