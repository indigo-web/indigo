package indigo

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/http/codec"
	"github.com/indigo-web/indigo/http/cookie"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/mime"
	"github.com/indigo-web/indigo/http/proto"
	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/httptest/parse"
	"github.com/indigo-web/indigo/internal/httptest/serialize"
	"github.com/indigo-web/indigo/kv"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/router/inbuilt/middleware"
	"github.com/klauspost/compress/gzip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"iter"
	"net"
	stdhttp "net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	addr      = "localhost:16100"
	altAddr   = "localhost:16800"
	httpsAddr = "localhost:16443"
	appURL    = "http://" + addr
)

func getHeaders() http.Headers {
	return kv.New().
		Add("Host", "localhost:16100").
		Add("User-Agent", "Go-http-client/1.1").
		Add("Accept-Encoding", "gzip")
}

func respond(request *http.Request) *http.Response {
	str, err := serialize.Request(request)
	if err != nil {
		return http.Error(request, err)
	}

	return request.Respond().String(str)
}

func getInbuiltRouter() *inbuilt.Router {
	ctx := context.WithValue(context.Background(), "easter", "egg")

	r := inbuilt.New().
		Use(middleware.Recover).
		Use(middleware.CustomContext(ctx)).
		Static("/static", "tests")

	r.Resource("/").
		Get(respond).
		Post(respond)

	r.Get("/file/:name", func(request *http.Request) *http.Response {
		return request.Respond().File("tests/" + request.Vars.Value("name"))
	})

	r.Get("/file/by-path/:path...", func(request *http.Request) *http.Response {
		return http.File(request, request.Vars.Value("path"))
	})

	r.Post("/body-reader", func(request *http.Request) *http.Response {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return http.Error(request, err)
		}

		return http.Bytes(request, body)
	})

	r.Get("/query", func(request *http.Request) *http.Response {
		var buff []byte

		for key, value := range request.Params.Pairs() {
			buff = append(buff, key+":"+value+"."...)
		}

		return http.Bytes(request, buff)
	})

	r.Get("/hijack", func(request *http.Request) *http.Response {
		client, err := request.Hijack()
		if err != nil {
			return nil
		}

		_, _ = client.Write([]byte("j"))
		return nil
	})

	r.Get("/ctx-value", func(request *http.Request) *http.Response {
		return http.String(request, request.Ctx.Value("easter").(string))
	})

	r.Get("/panic", func(request *http.Request) *http.Response {
		panic("ich kann das nicht mehr ertragen")
	}, middleware.Recover)

	r.Get("/json", func(request *http.Request) *http.Response {
		fields := slices.Collect(request.Headers.Values("fields"))

		return http.JSON(request, headersToMap(request.Headers, fields))
	})

	r.Get("/custom-error-with-code", func(request *http.Request) *http.Response {
		return http.Error(request, status.ErrTeapot, status.Teapot)
	})

	r.Get("/cookie", func(request *http.Request) *http.Response {
		jar, err := request.Cookies()
		if err != nil {
			return http.Error(request, err)
		}

		buff := make([]byte, 0, 512)
		for _, c := range jar.Expose() {
			buff = append(buff, fmt.Sprintf("%s=%s\n", c.Key, c.Value)...)
		}

		return http.Bytes(request, buff).
			Cookie(cookie.New("hello", "world")).
			Cookie(cookie.New("men", "in black"))
	})

	r.Get("/form", func(request *http.Request) *http.Response {
		form, err := request.Body.Form()
		if err != nil {
			return http.Error(request, err)
		}

		buff := make([]byte, 0, 512)
		for _, pair := range form {
			buff = append(buff, fmt.Sprintf("%s=%s\n", pair.Name, pair.Value)...)
		}

		return http.Bytes(request, buff)
	})

	return r
}

func headersToMap(hdrs http.Headers, keys []string) map[string]string {
	m := make(map[string]string, len(keys))

	for _, key := range keys {
		m[key] = strings.Join(slices.Collect(hdrs.Values(key)), ", ")
	}

	return m
}

func readFullBody(t *testing.T, resp *stdhttp.Response) string {
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, resp.Body.Close())

	return string(body)
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
		s := config.Default()
		s.NET.ReadTimeout = 1 * time.Second
		require.NoError(t, app.
			Tune(s).
			OnStart(func() {
				ch <- struct{}{}
			}).
			OnStop(func() {
				ch <- struct{}{}
			}).
			Listen(altAddr, TCP()).
			TLS(httpsAddr, LocalCert()).
			Serve(r),
		)
	}(app)

	<-ch
	// Ensure the server is ready to accept connections
	waitForAvailability(t)

	t.Run("root get", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(appURL + "/")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		request, err := parse.HTTP11Request(readFullBody(t, resp))
		require.NoError(t, err)

		require.Equal(t, method.GET, request.Method)
		require.Equal(t, "/", request.Path)
		require.Equal(t, proto.HTTP11, request.Protocol)
		wantHeaders := getHeaders().Add("Content-Length", "0")
		for err = range compareHeaders(request.Headers, wantHeaders) {
			assert.Fail(t, err.Error())
		}
		body, err := request.Body.String()
		require.NoError(t, err)
		require.Empty(t, body)
	})

	t.Run("root get with body", func(t *testing.T) {
		r := strings.NewReader("Hello, world!")
		req, err := stdhttp.NewRequest(stdhttp.MethodGet, appURL+"/", r)
		require.NoError(t, err)
		resp, err := stdhttp.DefaultClient.Do(req)
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		request, err := parse.HTTP11Request(readFullBody(t, resp))
		require.NoError(t, err)

		require.Equal(t, method.GET, request.Method)
		require.Equal(t, "/", request.Path)
		require.Equal(t, proto.HTTP11, request.Protocol)
		wantHeaders := getHeaders().Add("Content-Length", "13")
		for err = range compareHeaders(request.Headers, wantHeaders) {
			assert.Fail(t, err.Error())
		}
		body, err := request.Body.String()
		require.NoError(t, err)
		require.Equal(t, "Hello, world!", body)
	})

	t.Run("root post", func(t *testing.T) {
		r := strings.NewReader("Hello, world!")
		resp, err := stdhttp.DefaultClient.Post(appURL+"/", "text/html", r)
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		request, err := parse.HTTP11Request(readFullBody(t, resp))
		require.NoError(t, err)

		require.Equal(t, method.POST, request.Method)
		require.Equal(t, "/", request.Path)
		require.Equal(t, proto.HTTP11, request.Protocol)
		wantHeaders := getHeaders().
			Add("Content-Type", "text/html").
			Add("Content-Length", "13")
		for err = range compareHeaders(request.Headers, wantHeaders) {
			assert.Fail(t, err.Error())
		}
		body, err := request.Body.String()
		require.NoError(t, err)
		require.Equal(t, "Hello, world!", body)
	})

	t.Run("root head", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Head(appURL + "/")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		require.Empty(t, readFullBody(t, resp))
	})

	t.Run("accept encoding", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Head(appURL + "/")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		require.Equal(t, []string{"identity"}, resp.Header["Accept-Encoding"])
		require.Empty(t, readFullBody(t, resp))
	})

	t.Run("body reader", func(t *testing.T) {
		r := strings.NewReader("Hello, world!")
		resp, err := stdhttp.DefaultClient.Post(appURL+"/body-reader", "text/html", r)
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		body := readFullBody(t, resp)
		require.Equal(t, "Hello, world!", body)
	})

	t.Run("error with custom code", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(appURL + "/custom-error-with-code")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusTeapot, resp.StatusCode)
		require.Empty(t, readFullBody(t, resp))
	})

	t.Run("with query", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(appURL + "/query?hello=world&%20foo=+bar")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		require.Equal(t, "hello:world. foo: bar.", readFullBody(t, resp))
	})

	t.Run("body reader", func(t *testing.T) {
		r := strings.NewReader("Hello, world!")
		resp, err := stdhttp.DefaultClient.Post(appURL+"/", "text/html", r)
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		request, err := parse.HTTP11Request(readFullBody(t, resp))
		require.NoError(t, err)

		require.Equal(t, method.POST, request.Method)
		require.Equal(t, "/", request.Path)
		require.Equal(t, proto.HTTP11, request.Protocol)
		wantHeaders := getHeaders().
			Add("Content-Type", "text/html").
			Add("Content-Length", "13")
		for err = range compareHeaders(request.Headers, wantHeaders) {
			assert.Fail(t, err.Error())
		}
		body, err := request.Body.String()
		require.NoError(t, err)
		require.Equal(t, "Hello, world!", body)
	})

	t.Run("hijacking", func(t *testing.T) {
		conn, err := sendSimpleRequest(addr, "/hijack")
		require.NoError(t, err)
		defer func() {
			_ = conn.Close()
		}()

		data, err := io.ReadAll(conn)
		require.NoError(t, err)
		require.Equal(t, "j", string(data))
	})

	t.Run("request existing file", func(t *testing.T) {
		actualContent, err := os.ReadFile("./tests/index.html")
		require.NoError(t, err)

		resp, err := stdhttp.DefaultClient.Get(appURL + "/file/index.html")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		require.Equal(t, string(actualContent), readFullBody(t, resp))
	})

	t.Run("request non-existing file", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Get(appURL + "/file/doesn't-exist.html")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusNotFound, resp.StatusCode)

		data, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Empty(t, string(data))
	})

	t.Run("request by path", func(t *testing.T) {
		actualContent, err := os.ReadFile("./tests/index.html")
		require.NoError(t, err)

		resp, err := stdhttp.DefaultClient.Get(appURL + "/file/by-path/tests/index.html")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		require.Equal(t, string(actualContent), readFullBody(t, resp))
	})

	testStatic := func(t *testing.T, file, mime string) {
		actualContent, err := os.ReadFile("./tests/" + file)
		require.NoError(t, err)

		resp, err := stdhttp.DefaultClient.Get(appURL + "/static/" + file)
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)

		if len(mime) > 0 {
			require.Equal(t, mime, resp.Header["Content-Type"][0])
		} else {
			require.Empty(t, resp.Header["Content-Type"])
		}

		require.Equal(t, string(actualContent), readFullBody(t, resp))
	}

	t.Run("request static html", func(t *testing.T) {
		testStatic(t, "index.html", "text/html;charset=utf8")
	})

	t.Run("request static css", func(t *testing.T) {
		testStatic(t, "styles.css", "text/css;charset=utf8")
	})

	t.Run("request static non-standard extension", func(t *testing.T) {
		testStatic(t, "pics.vfs", mime.Unset)
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
			require.True(t, slices.Contains(wantHeaderLines, line), "unwanted header line: "+line)
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

	t.Run("idle disconnect", func(t *testing.T) {
		conn, err := net.Dial("tcp", addr)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, conn.Close())
		}()

		response, err := io.ReadAll(conn)
		require.NoError(t, err)
		if len(response) > 0 {
			require.Failf(
				t, "wanted silent connection close, got response:\n%s",
				strconv.Quote(string(response)),
			)
		}
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
		testCtxValue(t, "https://"+httpsAddr)
	})

	t.Run("alternative port", func(t *testing.T) {
		testCtxValue(t, "http://"+altAddr)
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
		defer func() {
			require.NoError(t, conn.Close())
		}()
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
		defer func() {
			require.NoError(t, conn.Close())
		}()
		_, err = conn.Write([]byte(requests))
		require.NoError(t, err)
		require.NoError(t, conn.SetReadDeadline(time.Now().Add(time.Second)))
		data, err := io.ReadAll(conn)
		require.NoError(t, err)
		n := bytes.Count(data, []byte("HTTP/1.0 200 OK\r\n"))
		require.Equal(t, pipelinedRequests, n, "got less successful responses as expected")
	})

	t.Run("cookie", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodGet,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/cookie",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: stdhttp.Header{
				"Cookie": {"hello=world; men=in black", "anything=anywhere"},
			},
			Host:       addr,
			RemoteAddr: addr,
		}
		resp, err := stdhttp.DefaultClient.Do(request)
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		body := readFullBody(t, resp)
		require.Equal(t, "hello=world\nmen=in black\nanything=anywhere\n", body)
		require.Equal(t, []string{"hello=world", "men=in black"}, resp.Header.Values("Set-Cookie"))
	})

	t.Run("form urlencoded", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodGet,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/form",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: stdhttp.Header{
				"Content-Type": {mime.FormUrlencoded},
			},
			Host:       addr,
			RemoteAddr: addr,
			Body:       io.NopCloser(strings.NewReader("hello=world&my+name=Paul&a%2bb=5")),
		}
		resp, err := stdhttp.DefaultClient.Do(request)
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		body := readFullBody(t, resp)
		require.Equal(t, "hello=world\nmy name=Paul\na+b=5\n", body)
	})

	t.Run("form multipart", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodGet,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/form",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: stdhttp.Header{
				"Content-Type": {"multipart/form-data; boundary=someBoundary"},
			},
			Host:       addr,
			RemoteAddr: addr,
			Body: io.NopCloser(strings.NewReader(
				"--someBoundary\r\n" +
					"Content-Disposition: form-data; name=\"hello\"\r\n\r\n" +
					"world" +
					"\r\n--someBoundary\r\n" +
					"Content-Disposition: form-data; name=\"my+name\"\r\n\r\n" +
					"Paul" +
					"\r\n--someBoundary--\r\n",
			)),
		}
		resp, err := stdhttp.DefaultClient.Do(request)
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		body := readFullBody(t, resp)
		require.Equal(t, "hello=world\nmy name=Paul\n", body)
	})

	t.Run("chunked body", func(t *testing.T) {
		request := &stdhttp.Request{
			Method: stdhttp.MethodPost,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/body-reader",
			},
			Proto:            "HTTP/1.1",
			ProtoMajor:       1,
			ProtoMinor:       1,
			TransferEncoding: []string{"chunked"},
			Host:             addr,
			RemoteAddr:       addr,
			Body:             newCircularReader("Mozilla", " ", "Developer Network"),
		}
		resp, err := stdhttp.DefaultClient.Do(request)
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		body := readFullBody(t, resp)
		require.NoError(t, resp.Body.Close())
		require.Equal(t, "Mozilla Developer Network", body)
	})

	t.Run("forced stop", func(t *testing.T) {
		app.Stop()

		timer := time.NewTimer(5 * time.Second)
		select {
		case <-ch:
		case <-timer.C:
			require.Fail(t, "server did not shut down correctly")
		}
	})
}

func TestSecondPhase(t *testing.T) {
	// second phase starts a new server instance with different configuration in order
	// to cover cases, that could not be covered in the first phase

	ch := make(chan struct{})
	app := New(addr)
	go func(app *App) {
		r := getInbuiltRouter()
		s := config.Default()
		s.NET.ReadTimeout = 1 * time.Second
		_ = app.
			Tune(s).
			Codec(codec.NewGZIP()).
			OnStart(func() {
				ch <- struct{}{}
			}).
			OnStop(func() {
				ch <- struct{}{}
			}).
			Serve(r)
	}(app)

	<-ch
	waitForAvailability(t)

	t.Run("accept encoding", func(t *testing.T) {
		resp, err := stdhttp.DefaultClient.Head(appURL + "/")
		require.NoError(t, err)
		require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
		require.Equal(t, []string{"gzip"}, resp.Header["Accept-Encoding"])
		require.Empty(t, readFullBody(t, resp))
	})

	t.Run("gzip compressed request", func(t *testing.T) {
		const data = "Hello, world!"
		buff := bytes.NewBuffer(nil)
		c := gzip.NewWriter(buff)
		_, err := c.Write([]byte(data))
		require.NoError(t, err)
		require.NoError(t, c.Close())

		request := &stdhttp.Request{
			Method: stdhttp.MethodPost,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/body-reader",
			},
			Proto:            "HTTP/1.1",
			ProtoMajor:       1,
			ProtoMinor:       1,
			TransferEncoding: []string{"chunked"},
			Host:             addr,
			RemoteAddr:       addr,
			Body:             io.NopCloser(buff),
		}
		request.Header = make(stdhttp.Header)
		request.Header.Set("Content-Encoding", "gzip")
		resp, err := stdhttp.DefaultClient.Do(request)
		require.NoError(t, err)
		require.Equal(t, data, readFullBody(t, resp))
	})

	t.Run("idle disconnect", func(t *testing.T) {
		conn, err := net.Dial("tcp4", addr)
		require.NoError(t, err)
		defer func() {
			require.NoError(t, conn.Close())
		}()

		connectionClosed := make(chan struct{})
		go func() {
			// the goroutine finishes when the connection is closed
			_, _ = io.ReadAll(conn)
			connectionClosed <- struct{}{}
		}()

		select {
		case <-connectionClosed:
		case <-time.NewTimer(5 * time.Second).C:
			require.Fail(t, "idle connection stays alive")
		}
	})

	doRequest := func(conn net.Conn) error {
		_, _ = conn.Write([]byte("GET / HTTP/1.1\r\n\r\n"))
		buff := make([]byte, 4096)
		_, err := conn.Read(buff)
		return err
	}

	t.Run("shutdown", func(t *testing.T) {
		client := func(ch chan<- error) {
			conn, err := net.Dial("tcp", addr)
			ch <- nil
			if err != nil {
				ch <- err
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
		<-first

		app.Stop()

		second := make(chan error)
		go client(second)
		<-second

		require.Error(t, <-second)
		require.NoError(t, <-first)
	})
}

func waitForAvailability(t *testing.T) {
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.Dial("tcp4", addr)
		if err == nil {
			_ = conn.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("server did not start listening on %s in time: %v", addr, err)
		}
		time.Sleep(50 * time.Millisecond)
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

func compareHeaders(a, b http.Headers) iter.Seq[error] {
	return func(yield func(error) bool) {
		for key := range a.Keys() {
			av, bv := slices.Collect(a.Values(key)), slices.Collect(b.Values(key))
			if slices.Compare(av, bv) != 0 {
				if !yield(fmt.Errorf("%s: mismatching values", key)) {
					return
				}

				break
			}
		}
	}
}

type circularReader struct {
	data []string
	ptr  int
}

func newCircularReader(data ...string) *circularReader {
	return &circularReader{data: data}
}

func (c *circularReader) Read(b []byte) (int, error) {
	if c.ptr >= len(c.data) {
		return 0, io.EOF
	}

	// b is assumed to be big enough
	n := copy(b, c.data[c.ptr])
	c.ptr++

	return n, nil
}

func (c *circularReader) Close() error {
	return nil
}
