package indigo

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/indigo-web/indigo/http/encryption"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/httptest"
	"github.com/indigo-web/indigo/router/inbuilt/middleware"
	"github.com/stretchr/testify/assert"
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

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/settings"
)

const (
	addr      = "localhost:16100"
	altPort   = 16800
	httpsPort = 16443
	appURL    = "http://" + addr
)

func getHeaders() headers.Headers {
	return headers.New().
		Add("Host", "localhost:16100").
		Add("User-Agent", "Go-http-client/1.1").
		Add("Accept-Encoding", "gzip")
}

func respond(request *http.Request) *http.Response {
	str, err := httptest.Dump(request)
	if err != nil {
		return http.Error(request, err)
	}

	return request.Respond().String(str)
}

func getInbuiltRouter() *inbuilt.Router {
	ctx := context.WithValue(context.Background(), "easter", "egg")

	r := inbuilt.New().
		Use(middleware.Recover).
		Use(middleware.CustomContext(ctx))

	r.Resource("/").
		Get(respond).
		Post(respond)

	r.Get("/file/{name}", func(request *http.Request) *http.Response {
		return request.Respond().File("tests/" + request.Params.Value("name"))
	})

	r.Post("/body-reader", func(request *http.Request) *http.Response {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return http.Error(request, err)
		}

		return http.Bytes(request, body)
	})

	r.Get("/query", func(request *http.Request) (_ *http.Response) {
		params, err := request.Query.Unwrap()
		if err != nil {
			return http.Error(request, err)
		}

		var buff []byte

		for _, pair := range params.Unwrap() {
			buff = append(buff, pair.Key+":"+pair.Value+"."...)
		}

		return http.Bytes(request, buff)
	})

	r.Get("/hijack", func(request *http.Request) (_ *http.Response) {
		conn, err := request.Hijack()
		defer func(conn net.Conn) {
			_ = conn.Close()
		}(conn)
		if err != nil {
			return nil
		}

		if _, err = conn.Write([]byte("j")); err != nil {
			return nil
		}

		return
	})

	r.Get("/ctx-value", func(request *http.Request) *http.Response {
		return http.String(request, request.Ctx.Value("easter").(string))
	})

	r.Get("/panic", func(request *http.Request) (_ *http.Response) {
		panic("ich kann das nicht mehr ertragen")
	}, middleware.Recover)

	r.Get("/json", func(request *http.Request) (_ *http.Response) {
		fields := request.Headers.Values("fields")

		return http.JSON(request, headersToMap(request.Headers, fields))
	})

	return r
}

func headersToMap(hdrs headers.Headers, keys []string) map[string]string {
	m := make(map[string]string, len(keys))

	for _, key := range keys {
		m[key] = strings.Join(hdrs.Values(key), ", ")
	}

	return m
}

func TestFirstPhase(t *testing.T) {
	ch := make(chan struct{})
	app := New(addr)
	go func(app *App) {
		r := getInbuiltRouter().
			Use(
				middleware.CustomContext(
					context.WithValue(context.Background(), "easter", "egg"),
				),
			)
		s := settings.Default()
		s.TCP.ReadTimeout = 500 * time.Millisecond
		_ = app.
			Tune(s).
			NotifyOnStart(func() {
				ch <- struct{}{}
			}).
			NotifyOnStop(func() {
				ch <- struct{}{}
			}).
			Listen(altPort, encryption.Plain).
			AutoHTTPS(httpsPort).
			Serve(r)
	}(app)

	<-ch

	t.Run("root get", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(appURL + "/")
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		repr, err := parseBody(resp)
		require.NoError(t, err)

		for _, err := range httptest.Compare(repr, httptest.Request{
			Method: method.GET,
			Path:   "/",
			Proto:  "HTTP/1.1",
			Headers: getHeaders().
				Add("Content-Length", "0"),
			Body: "",
		}) {
			assert.NoError(t, err)
		}
	})

	t.Run("root get with body", func(t *testing.T) {
		r := strings.NewReader("Hello, world!")
		req, err := stdhttp.NewRequest(stdhttp.MethodGet, appURL+"/", r)
		require.NoError(t, err)
		resp, err := stdhttp.DefaultClient.Do(req)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		repr, err := parseBody(resp)
		require.NoError(t, err)

		for _, err := range httptest.Compare(repr, httptest.Request{
			Method: method.GET,
			Path:   "/",
			Proto:  "HTTP/1.1",
			Headers: getHeaders().
				Add("Content-Length", "13"),
			Body: "Hello, world!",
		}) {
			assert.NoError(t, err)
		}
	})

	t.Run("root post", func(t *testing.T) {
		r := strings.NewReader("Hello, world!")
		resp, err := stdhttp.DefaultClient.Post(appURL+"/", "text/html", r)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		repr, err := parseBody(resp)
		require.NoError(t, err)

		for _, err := range httptest.Compare(repr, httptest.Request{
			Method: method.POST,
			Path:   "/",
			Proto:  "HTTP/1.1",
			Headers: getHeaders().
				Add("Content-Length", "13").
				Add("Content-Type", "text/html"),
			Body: "Hello, world!",
		}) {
			assert.NoError(t, err)
		}
	})

	t.Run("body reader", func(t *testing.T) {
		r := strings.NewReader("Hello, world!")
		resp, err := stdhttp.DefaultClient.Post(appURL+"/body-reader", "text/html", r)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "Hello, world!", string(body))
	})

	t.Run("root head", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Head(appURL + "/")
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()

		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, body)
	})

	t.Run("with query", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(appURL + "/query?hello=world&%20foo=+bar")
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "hello:world. foo: bar.", string(body))
	})

	t.Run("body reader", func(t *testing.T) {
		r := strings.NewReader("Hello, world!")
		resp, err := stdhttp.DefaultClient.Post(appURL+"/", "text/html", r)
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		repr, err := parseBody(resp)
		require.NoError(t, err)

		for _, err := range httptest.Compare(repr, httptest.Request{
			Method: method.POST,
			Path:   "/",
			Proto:  "HTTP/1.1",
			Headers: getHeaders().
				Add("Content-Length", "13").
				Add("Content-Type", "text/html"),
			Body: "Hello, world!",
		}) {
			assert.NoError(t, err)
		}
	})

	t.Run("hijacking", func(t *testing.T) {
		conn, err := sendSimpleRequest(addr, "/hijack")
		defer func() {
			_ = conn.Close()
		}()

		require.NoError(t, err)
		data, err := io.ReadAll(conn)
		require.NoError(t, err)
		require.Equal(t, "j", string(data))
	})

	t.Run("request existing file", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(appURL + "/file/index.html")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		actualContent, err := os.ReadFile("./tests/index.html")
		require.NoError(t, err)
		require.Equal(t, string(actualContent), string(data))
	})

	t.Run("request non-existing file", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(appURL + "/file/doesntexists.html")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusNotFound, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, string(data))
	})

	t.Run("trace", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodTrace,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/",
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
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		require.Contains(t, resp.Header, "Content-Type")
		require.Equal(t, 1, len(resp.Header["Content-Type"]), "too many content-type values")
		require.Equal(t, "message/http", resp.Header["Content-Type"][0])

		dataBytes, err := io.ReadAll(resp.Body)
		data := string(dataBytes)
		require.NoError(t, err)

		wantRequestLine := "TRACE / HTTP/1.1\r\n"
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

	t.Run("not allowed method", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodDelete,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/",
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
		allow := resp.Header["Allow"][0]
		require.True(t, allow == "GET,POST" || allow == "POST,GET")
		require.Equal(t, 1, len(resp.Header["Allow"]))
	})

	// this test must ALWAYS be on the bottom as it is the longest-duration test
	t.Run("idle disconnect", func(t *testing.T) {
		conn, err := net.Dial("tcp4", addr)
		require.NoError(t, err)
		defer conn.Close()

		response, err := io.ReadAll(conn)
		require.NoError(t, err)
		require.Empty(t, response, "must no data be transmitted")
	})

	testCtxValue := func(t *testing.T, addr string) {
		tr := &stdhttp.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &stdhttp.Client{Transport: tr}
		resp, err := client.Get(addr + "/ctx-value")
		require.NoError(t, err)
		defer func() {
			_ = resp.Body.Close()
		}()
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, "egg", string(body))
	}

	t.Run("ctx value", func(t *testing.T) {
		testCtxValue(t, appURL)
	})

	t.Run("https", func(t *testing.T) {
		testCtxValue(t, "https://localhost:"+strconv.Itoa(httpsPort))
	})

	t.Run("alternative port", func(t *testing.T) {
		testCtxValue(t, "http://localhost:"+strconv.Itoa(altPort))
	})

	requireField := func(t *testing.T, m map[string]any, key, value string) {
		actual, found := m[key]
		require.Truef(t, found, "json doesn't contain the key %s", key)
		require.Equal(t, value, actual.(string))
	}

	t.Run("json", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodGet,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/json",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: stdhttp.Header{
				"Fields": {"hello", "foo"},
				"Hello":  {"world"},
				"Foo":    {"bar", "spam"},
			},
			Host:       addr,
			RemoteAddr: addr,
		}
		resp, err := stdhttp.DefaultClient.Do(request)
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		result := make(map[string]any)
		decoder := json.NewDecoder(resp.Body)
		err = decoder.Decode(&result)
		require.NoError(t, err)
		requireField(t, result, "hello", "world")
		requireField(t, result, "foo", "bar, spam")
	})

	t.Run("HTTP/1.0 no explicit keep-alive", func(t *testing.T) {
		raw := "GET /ctx-value HTTP/1.0\r\n\r\n"
		conn, err := net.Dial("tcp", addr)
		require.NoError(t, err)
		defer conn.Close()
		_, err = conn.Write([]byte(raw))
		require.NoError(t, err)
		require.NoError(t, conn.SetReadDeadline(time.Now().Add(time.Second)))
		_, err = io.ReadAll(conn)
		require.NoError(t, err)
	})

	t.Run("HTTP/1.0 with keep-alive", func(t *testing.T) {
		raw := "GET /ctx-value HTTP/1.0\r\nConnection: keep-alive\r\n\r\n"
		const pipelinedRequests = 10
		requests := strings.Repeat(raw, pipelinedRequests)
		conn, err := net.Dial("tcp", addr)
		require.NoError(t, err)
		defer conn.Close()
		_, err = conn.Write([]byte(requests))
		require.NoError(t, err)
		require.NoError(t, conn.SetReadDeadline(time.Now().Add(time.Second)))
		data, err := io.ReadAll(conn)
		require.NoError(t, err)
		n := bytes.Count(data, []byte("HTTP/1.0 200 OK\r\n"))
		require.Equal(t, pipelinedRequests, n, "got less successful responses as expected")
	})

	t.Run("chunked body", func(t *testing.T) {
		request := "POST /body-reader HTTP/1.1\r\n" +
			"Connection: close\r\n" +
			"Transfer-Encoding: chunked\r\n" +
			"\r\n" +
			"7\r\nMozilla\r\n1\r\n \r\n11\r\nDeveloper Network\r\n0\r\n\r\n"
		resp, err := send(addr, []byte(request))
		require.NoError(t, err)
		repr, err := httptest.Parse(string(resp))
		require.NoError(t, err)
		require.Equal(t, "Mozilla Developer Network", repr.Body)
	})

	t.Run("forced stop", func(t *testing.T) {
		app.Stop()
		_, ok := chanRead(ch, 5*time.Second)
		require.True(t, ok, "server did not shut down correctly")
	})
}

func TestSecondPhase(t *testing.T) {
	// second phase starts a new server instance with different configuration in order
	// to cover cases, that could not be covered in the first phase

	ch := make(chan struct{})
	app := New(addr)
	go func(app *App) {
		r := getInbuiltRouter()
		s := settings.Default()
		s.TCP.ReadTimeout = 500 * time.Millisecond
		s.HTTP.OnDisconnect = func(request *http.Request) *http.Response {
			return http.Error(request, status.ErrRequestTimeout)
		}
		_ = app.
			Tune(s).
			NotifyOnStart(func() {
				ch <- struct{}{}
			}).
			NotifyOnStop(func() {
				ch <- struct{}{}
			}).
			Listen(altPort, encryption.Plain).
			AutoHTTPS(httpsPort).
			Serve(r)
	}(app)

	<-ch

	t.Run("idle disconnect", func(t *testing.T) {
		conn, err := net.Dial("tcp4", addr)
		require.NoError(t, err)
		defer conn.Close()

		response, err := io.ReadAll(conn)
		require.NoError(t, err)
		wantResponseLine := "HTTP/1.1 408 Request Timeout\r\n"
		lf := bytes.IndexByte(response, '\n')
		require.NotEqual(t, -1, lf, "http response must contain at least one LF")
		require.Equal(t, wantResponseLine, string(response[:lf+1]))
	})

	doRequest := func(conn net.Conn) error {
		_, _ = conn.Write([]byte("GET / HTTP/1.1\r\n\r\n"))
		buff := make([]byte, 4096)
		_, err := conn.Read(buff)
		return err
	}

	t.Run("graceful shutdown", func(t *testing.T) {
		client := func(ch chan<- error) {
			conn, err := net.Dial("tcp", addr)
			ch <- err
			if err != nil {
				return
			}

			for i := 0; i < 20; i++ {
				if err := doRequest(conn); err != nil {
					ch <- err
					return
				}

				time.Sleep(100 * time.Millisecond)
			}

			ch <- nil
		}

		first := make(chan error)
		go client(first)
		require.NoError(t, <-first)

		// as in main test-case there is no way to combine this test with forced shutdown
		app.GracefulStop()
		time.Sleep(100 * time.Millisecond)
		second := make(chan error)
		go client(second)
		// the socket on graceful shutdown isn't closed, it just stops accepting new connections.
		// So by that, we are able to connect freely
		<-second
		// ...the thing is, that we'll get the read-tcp error on a try to send something
		require.Error(t, <-second)
		require.NoError(t, <-first)
	})
}

func chanRead[T any](ch <-chan T, timeout time.Duration) (value T, ok bool) {
	timer := time.NewTimer(timeout)
	select {
	case value = <-ch:
		return value, true
	case <-timer.C:
		return value, false
	}
}

func sendSimpleRequest(addr, path string) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	request := fmt.Sprintf("GET %s HTTP/1.1\r\n\r\n", path)

	_, err = conn.Write([]byte(request))

	return conn, err
}

func send(addr string, req []byte) ([]byte, error) {
	conn, err := net.Dial("tcp", addr)
	defer conn.Close()
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(req)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(conn)
}

func parseBody(resp *stdhttp.Response) (httptest.Request, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return httptest.Request{}, err
	}

	return httptest.Parse(string(body))
}

func contains(strs []string, substr string) bool {
	for _, str := range strs {
		if str == substr {
			return true
		}
	}

	return false
}
