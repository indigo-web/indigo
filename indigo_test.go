package indigo

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"github.com/indigo-web/indigo/ctx"
	"github.com/indigo-web/indigo/http/coding"
	"github.com/stretchr/testify/require"
	"io"
	"net"
	stdhttp "net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/indigo-web/indigo/http/status"

	"github.com/indigo-web/indigo/settings"

	"github.com/indigo-web/indigo/http"

	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/router"
	"github.com/indigo-web/indigo/router/inbuilt"
)

const (
	host       = "localhost"
	port       = uint16(16100)
	portString = "16100"
	addr       = host + ":" + portString
	URL        = "http://" + addr
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

func getStaticRouter(t *testing.T) router.Router {
	r := inbuilt.New()

	r.Get("/simple-get", func(request *http.Request) *http.Response {
		require.Equal(t, method.GET, request.Method)
		_, err := request.Query.Get("some non-existing query key")
		require.Error(t, err)
		require.Equal(t, proto.HTTP11, request.Proto)

		return request.Respond()
	})

	r.Get("/get-resp-body", func(request *http.Request) *http.Response {
		return request.Respond().WithBody(testRequestBody)
	})

	r.Get("/get-read-body", func(request *http.Request) *http.Response {
		require.True(t, request.Headers.Has(testHeaderKey))
		requestHeader := strings.Join(request.Headers.Values(testHeaderKey), ",")
		require.Equal(t, testHeaderValue, requestHeader)

		body, err := request.Body().Full()
		require.NoError(t, err)
		require.Empty(t, body)

		return request.Respond()
	})

	r.Group("/with-").
		Get("query", func(request *http.Request) *http.Response {
			value, err := request.Query.Get(testQueryKey)
			require.NoError(t, err)
			require.Equal(t, testQueryValue, value)

			return request.Respond()
		}).
		Get("file", func(request *http.Request) *http.Response {
			resp, err := request.Respond().WithFile(testFilename)
			require.NoError(t, err)
			return resp
		}).
		Get("file-notfound", func(request *http.Request) *http.Response {
			_, err := request.Respond().WithFile(testFilename + "non-existing")
			require.Error(t, err)

			return request.Respond().WithBody(testFileIfNotFound)
		})

	// request.OnBody() is not tested because request.Body() (wrapper for OnBody)
	// is tested, that is enough to make sure that requestbody works correct

	r.Post("/read-body", func(request *http.Request) *http.Response {
		body, err := request.Body().Full()
		require.NoError(t, err)
		require.Equal(t, testRequestBody, string(body))

		return request.Respond()
	})

	r.Post("/read-body-gzipped", func(request *http.Request) *http.Response {
		body, err := request.Body().Full()
		require.NoError(t, err)
		require.Equal(t, testRequestBody, string(body))

		return request.Respond()
	})

	r.Post("/body-reader", func(request *http.Request) *http.Response {
		body, err := io.ReadAll(request.Body())
		require.NoError(t, err)
		require.Equal(t, testRequestBody, string(body))

		return request.Respond()
	})

	r.Post("/do-not-read-body", http.Respond)

	r.Get("/hijack-conn-no-body-read", func(request *http.Request) *http.Response {
		conn, err := request.Hijack()
		require.NoError(t, err)

		// just to notify client that we are ready for receiving something
		_, _ = conn.Write([]byte("a"))

		data, err := readN(conn, len(testRequestBody))
		require.NoError(t, err)
		require.Equal(t, testRequestBody, string(data))

		_ = conn.Close()

		return request.Respond()
	})

	r.Get("/hijack-conn-with-body-read", func(request *http.Request) *http.Response {
		_, _ = request.Body().Full()

		conn, err := request.Hijack()
		require.NoError(t, err)

		_, _ = conn.Write([]byte("a"))

		data, err := readN(conn, len(testRequestBody))
		require.NoError(t, err)
		require.Equal(t, testRequestBody, string(data))

		_ = conn.Close()

		return request.Respond()
	})

	r.Get("/ctx-value", func(request *http.Request) *http.Response {
		require.Equal(t, "egg", request.Ctx.Value("easter").(string))

		return request.Respond()
	})

	return r
}

func TestServer_Static(t *testing.T) {
	r := getStaticRouter(t)
	s := settings.Default()
	s.TCP.ReadTimeout = 1 * time.Second
	app := NewApp(host, port)
	app.AddCoding(coding.NewGZIP)

	runningServer := newServer(ctx.WithValue(context.Background(), "easter", "egg"), app)
	go runningServer.Start(t, r, s)

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
		body := strings.NewReader(testRequestBody)
		resp, err := stdhttp.DefaultClient.Post(URL+"/read-body", "text/html", body)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	})

	t.Run("/read-body-gzipped", func(t *testing.T) {
		compressed, err := compressGZIP([]byte(testRequestBody))
		require.NoError(t, err)
		request := fmt.Sprintf(
			"POST /read-body-gzipped HTTP/1.1\r\n"+
				"Transfer-Encoding: gzip\r\n"+
				"Content-Length: %d\r\n"+
				"Connection: close\r\n\r\n%s",
			len(compressed), string(compressed),
		)

		conn, err := net.Dial("tcp4", addr)
		require.NoError(t, err)
		_, err = conn.Write([]byte(request))
		require.NoError(t, err)
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		code, err := checkResponseStatus(buff[:n])
		require.NoError(t, err)
		require.Equal(t, status.OK, code)
	})

	t.Run("/body-reader", func(t *testing.T) {
		body := new(bytes.Buffer)
		body.Write([]byte(testRequestBody))
		resp, err := stdhttp.DefaultClient.Post(URL+"/body-reader", "text/html", body)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	})

	t.Run("/do-not-read-body", func(t *testing.T) {
		body := new(bytes.Buffer)
		body.Write([]byte(testRequestBody))
		resp, err := stdhttp.DefaultClient.Post(URL+"/read-body", "text/html", body)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	})

	t.Run("/hijack-conn-no-body-read", func(t *testing.T) {
		sendSimpleRequest(t, "/hijack-conn-no-body-read", addr)
	})

	t.Run("/hijack-conn-with-body-read", func(t *testing.T) {
		sendSimpleRequest(t, "/hijack-conn-with-body-read", addr)
	})

	t.Run("/with-file", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(URL + "/with-file")
		require.NoError(t, err)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		actualContent, err := os.ReadFile(testFilename)
		require.NoError(t, err)
		require.Equal(t, string(actualContent), string(data))
	})

	t.Run("/with-file-notfound", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(URL + "/with-file-notfound")
		require.NoError(t, err)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, testFileIfNotFound, string(data))
	})

	t.Run("/trace", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodTrace,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/simple-get",
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

		wantRequestLine := "TRACE /simple-get HTTP/1.1\r\n"
		require.Greater(t, len(data), len(wantRequestLine))
		require.Equal(t, wantRequestLine, data[:len(wantRequestLine)])

		headerLines := strings.Split(data[len(wantRequestLine):], "\r\n")
		// request is terminated with \r\n\r\n, so 2 last values in headerLines
		// are empty strings. Remove them
		headerLines = headerLines[:len(headerLines)-2]
		wantHeaderLines := []string{
			"Hello: World!",
			"Host: " + addr,
			"User-Agent: Go-http-client/1.1",
			"Accept-Encoding: gzip",
			"Content-Length: 0",
		}

		for _, line := range headerLines {
			require.True(t, contains(wantHeaderLines, line), "unwanted header line: "+line)
		}
	})

	t.Run("/method-not-allowed-allow-header", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodDelete,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/simple-get",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Host:       addr,
			RemoteAddr: addr,
		}
		resp, err := stdhttp.DefaultClient.Do(request)
		require.NoError(t, err)
		require.Equal(t, int(status.MethodNotAllowed), resp.StatusCode)

		require.Contains(t, resp.Header, "Allow")
		require.Equal(t, "GET", resp.Header["Allow"][0])
		require.Equal(t, 1, len(resp.Header["Allow"]))
	})

	// this test must ALWAYS be on the bottom as it is the longest-duration test
	t.Run("/test-idle-disconnect", func(t *testing.T) {
		conn, err := net.Dial("tcp4", addr)
		require.NoError(t, err)

		body, err := io.ReadAll(conn)
		wantResponseLine := "HTTP/1.1 408 Request Timeout\r\n"
		require.GreaterOrEqual(t, len(body), len(wantResponseLine))
		require.Equal(t, wantResponseLine, string(body[:len(wantResponseLine)]))
		// io.ReadAll() returns nil as an error in case error was io.EOF
		require.NoError(t, err)
	})

	t.Run("/ctx-value", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(URL + "/ctx-value")
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	})

	stdhttp.DefaultClient.CloseIdleConnections()
	runningServer.Wait(t, 3*time.Second)
}

func sendSimpleRequest(t *testing.T, path string, addr string) {
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

type serverWrap struct {
	app      *Application
	shutdown chan struct{}
	ctx      context.Context
}

func newServer(ctx context.Context, app *Application) serverWrap {
	return serverWrap{
		app:      app,
		shutdown: make(chan struct{}),
		ctx:      ctx,
	}
}

func (s serverWrap) Start(t *testing.T, r router.Router, settings ...settings.Settings) {
	s.app.SetContext(s.ctx)
	require.Equal(t, status.ErrShutdown, s.app.Serve(r, settings...))
	s.shutdown <- struct{}{}
}

func (s serverWrap) Wait(t *testing.T, duration time.Duration) {
	s.app.Stop()

	select {
	case <-s.shutdown:
	case <-time.After(duration):
		require.Fail(t, "server is not shutting down for too long")
	}
}

func checkResponseStatus(resp []byte) (status.Code, error) {
	lines := bytes.Split(resp, []byte{'\r', '\n'})
	members := bytes.Split(lines[0], []byte{' '})
	if len(members) != 3 {
		return 0, errors.New("response line contains too many members")
	}

	code, err := strconv.Atoi(string(members[1]))
	if err != nil {
		return 0, errors.New("bad status code")
	}

	return status.Code(code), nil
}

func compressGZIP(data []byte) ([]byte, error) {
	r, w := io.Pipe()
	go func() {
		writer := gzip.NewWriter(w)
		_, _ = writer.Write(data)
		_ = writer.Close()
		_ = w.Close()
	}()

	return io.ReadAll(r)
}
