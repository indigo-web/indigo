package indigo

import (
	"context"
	"errors"
	"github.com/indigo-web/indigo/internal/httptest"
	"github.com/indigo-web/indigo/router/inbuilt/middleware"
	"io"
	"net"
	"testing"
	"time"

	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/router/inbuilt"
	"github.com/indigo-web/indigo/settings"
)

const (
	addr = "localhost:16100"
	URL  = "http://" + addr
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
		return request.Respond().File(request.Params.Value("name"))
	})

	r.Post("/body-reader", func(request *http.Request) *http.Response {
		body, err := io.ReadAll(request.Body)
		if err != nil {
			return http.Error(request, err)
		}

		return http.Bytes(request, body)
	})

	r.Get("/hijack", func(request *http.Request) (_ *http.Response) {
		conn, err := request.Hijack()
		defer func(conn net.Conn) {
			_ = conn.Close()
		}(conn)
		if err != nil {
			return http.Error(request, err)
		}

		length, err := readN(conn, 1)
		if err != nil {
			return
		}

		data, err := readN(conn, int(length[0]))
		if err != nil {
			return
		}

		_, _ = conn.Write(data)

		return
	})

	r.Get("/ctx-value", func(request *http.Request) *http.Response {
		return http.String(request, request.Ctx.Value("easter").(string))
	})

	return r
}

func TestServer(t *testing.T) {
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
		s.TCP.ReadTimeout = 1 * time.Second
		_ = app.
			Tune(s).
			NotifyOnStart(func() {
				ch <- struct{}{}
			}).
			NotifyOnStop(func() {
				ch <- struct{}{}
			}).
			Serve(r)
	}(app)

	<-ch

	//t.Run("/simple-get", func(t *testing.T) {
	//	resp, err := stdhttp.DefaultClient.Get(URL + "/simple-get")
	//	require.NoError(t, err)
	//	defer func() {
	//		_ = resp.Body.Close()
	//	}()
	//	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	//	require.Equal(t, "200 "+stdhttp.StatusText(stdhttp.StatusOK), resp.Status)
	//	require.Equal(t, "HTTP/1.1", resp.Proto)
	//})
	//
	//t.Run("/get-resp-body", func(t *testing.T) {
	//	resp, err := stdhttp.DefaultClient.Get(URL + "/get-resp-body")
	//	require.NoError(t, err)
	//	defer func() {
	//		_ = resp.Body.Close()
	//	}()
	//
	//	//body, err := io.ReadAll(resp.Body)
	//	//require.NoError(t, err)
	//	//require.Equal(t, testRequestBody, string(body))
	//})
	//
	//t.Run("/head", func(t *testing.T) {
	//	resp, err := stdhttp.DefaultClient.Head(URL + "/get-resp-body")
	//	require.NoError(t, err)
	//	defer func() {
	//		_ = resp.Body.Close()
	//	}()
	//
	//	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	//	require.Equal(t, "200 "+stdhttp.StatusText(stdhttp.StatusOK), resp.Status)
	//	require.Equal(t, "HTTP/1.1", resp.Proto)
	//
	//	body, err := io.ReadAll(resp.Body)
	//	require.NoError(t, err)
	//	require.Empty(t, string(body))
	//})
	//
	//t.Run("/get-read-body", func(t *testing.T) {
	//	request := &stdhttp.Request{
	//		Method: stdhttp.MethodGet,
	//		URL: &url.URL{
	//			Scheme: "http",
	//			Host:   addr,
	//			Path:   "/get-read-body",
	//		},
	//		Proto:      "HTTP/1.1",
	//		ProtoMajor: 1,
	//		ProtoMinor: 1,
	//		Header: stdhttp.Header{
	//			"Hello": {"World!"},
	//		},
	//		Host:       addr,
	//		RemoteAddr: addr,
	//	}
	//	resp, err := stdhttp.DefaultClient.Do(request)
	//	require.NoError(t, err)
	//	defer func() {
	//		_ = resp.Body.Close()
	//	}()
	//	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	//})
	//
	//t.Run("/with-query", func(t *testing.T) {
	//	resp, err := stdhttp.DefaultClient.Get(URL + "/with-query?hel+lo=wor+ld")
	//	require.NoError(t, err)
	//	defer func() {
	//		_ = resp.Body.Close()
	//	}()
	//	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	//})
	//
	//t.Run("/read-body", func(t *testing.T) {
	//	body := strings.NewReader(testRequestBody)
	//	resp, err := stdhttp.DefaultClient.Post(URL+"/read-body", "text/html", body)
	//	require.NoError(t, err)
	//	defer func() {
	//		_ = resp.Body.Close()
	//	}()
	//	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	//})
	//
	//t.Run("/read-body-gzipped", func(t *testing.T) {
	//	t.Skip("gzipped does not works correctly yet")
	//	compressed, err := compressGZIP([]byte(testRequestBody))
	//	require.NoError(t, err)
	//	request := fmt.Sprintf(
	//		"POST /read-body-gzipped HTTP/1.1\r\n"+
	//			"Transfer-Encoding: gzip\r\n"+
	//			"Content-Length: %d\r\n"+
	//			"Connection: close\r\n\r\n%s",
	//		len(compressed), string(compressed),
	//	)
	//
	//	conn, err := net.Dial("tcp4", addr)
	//	require.NoError(t, err)
	//	_, err = conn.Write([]byte(request))
	//	require.NoError(t, err)
	//	buff := make([]byte, 1024)
	//	n, err := conn.Read(buff)
	//	code, err := checkResponseStatus(buff[:n])
	//	require.NoError(t, err)
	//	require.Equal(t, status.OK, code)
	//})
	//
	//t.Run("/body-reader", func(t *testing.T) {
	//	body := new(bytes.Buffer)
	//	body.Write([]byte(testRequestBody))
	//	resp, err := stdhttp.DefaultClient.Post(URL+"/body-reader", "text/html", body)
	//	require.NoError(t, err)
	//	defer func() {
	//		_ = resp.Body.Close()
	//	}()
	//	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	//})
	//
	//t.Run("/do-not-read-body", func(t *testing.T) {
	//	body := new(bytes.Buffer)
	//	body.Write([]byte(testRequestBody))
	//	resp, err := stdhttp.DefaultClient.Post(URL+"/read-body", "text/html", body)
	//	require.NoError(t, err)
	//	defer func() {
	//		_ = resp.Body.Close()
	//	}()
	//	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	//})
	//
	//t.Run("/hijack-conn-no-body-read", func(t *testing.T) {
	//	sendSimpleRequest(t, "/hijack-conn-no-body-read", addr)
	//})
	//
	//t.Run("/hijack-conn-with-body-read", func(t *testing.T) {
	//	sendSimpleRequest(t, "/hijack-conn-with-body-read", addr)
	//})
	//
	//t.Run("/with-file", func(t *testing.T) {
	//	resp, err := stdhttp.DefaultClient.Get(URL + "/with-file")
	//	require.NoError(t, err)
	//
	//	data, err := io.ReadAll(resp.Body)
	//	require.NoError(t, err)
	//
	//	actualContent, err := os.ReadFile(testFilename)
	//	require.NoError(t, err)
	//	require.Equal(t, string(actualContent), string(data))
	//})
	//
	//t.Run("/with-file-notfound", func(t *testing.T) {
	//	resp, err := stdhttp.DefaultClient.Get(URL + "/with-file-notfound")
	//	require.NoError(t, err)
	//
	//	data, err := io.ReadAll(resp.Body)
	//	require.NoError(t, err)
	//	require.Equal(t, testFileIfNotFound, string(data))
	//})
	//
	//t.Run("/trace", func(t *testing.T) {
	//	request := &stdhttp.Request{
	//		Method: stdhttp.MethodTrace,
	//		URL: &url.URL{
	//			Scheme: "http",
	//			Host:   addr,
	//			Path:   "/simple-get",
	//		},
	//		Proto:      "HTTP/1.1",
	//		ProtoMajor: 1,
	//		ProtoMinor: 1,
	//		Header: stdhttp.Header{
	//			"Hello": {"World!"},
	//		},
	//		Host:       addr,
	//		RemoteAddr: addr,
	//	}
	//	resp, err := stdhttp.DefaultClient.Do(request)
	//	require.NoError(t, err)
	//
	//	require.Contains(t, resp.Header, "Content-Type")
	//	require.Equal(t, 1, len(resp.Header["Content-Type"]), "too many content-type values")
	//	require.Equal(t, "message/http", resp.Header["Content-Type"][0])
	//
	//	dataBytes, err := io.ReadAll(resp.Body)
	//	data := string(dataBytes)
	//	require.NoError(t, err)
	//
	//	wantRequestLine := "TRACE /simple-get HTTP/1.1\r\n"
	//	require.Greater(t, len(data), len(wantRequestLine))
	//	require.Equal(t, wantRequestLine, data[:len(wantRequestLine)])
	//
	//	headerLines := strings.Split(data[len(wantRequestLine):], "\r\n")
	//	// request is terminated with \r\n\r\n, so 2 last values in headerLines
	//	// are empty strings. Remove them
	//	headerLines = headerLines[:len(headerLines)-2]
	//	wantHeaderLines := []string{
	//		"Hello: World!",
	//		"Host: " + addr,
	//		"User-Agent: Go-http-client/1.1",
	//		"Accept-Encoding: gzip",
	//		"Content-Length: 0",
	//	}
	//
	//	for _, line := range headerLines {
	//		require.True(t, contains(wantHeaderLines, line), "unwanted header line: "+line)
	//	}
	//})
	//
	//t.Run("/method-not-allowed-allow-header", func(t *testing.T) {
	//	request := &stdhttp.Request{
	//		Method: stdhttp.MethodDelete,
	//		URL: &url.URL{
	//			Scheme: "http",
	//			Host:   addr,
	//			Path:   "/simple-get",
	//		},
	//		Proto:      "HTTP/1.1",
	//		ProtoMajor: 1,
	//		ProtoMinor: 1,
	//		Host:       addr,
	//		RemoteAddr: addr,
	//	}
	//	resp, err := stdhttp.DefaultClient.Do(request)
	//	require.NoError(t, err)
	//	require.Equal(t, int(status.MethodNotAllowed), resp.StatusCode)
	//
	//	require.Contains(t, resp.Header, "Allow")
	//	require.Equal(t, "GET", resp.Header["Allow"][0])
	//	require.Equal(t, 1, len(resp.Header["Allow"]))
	//})
	//
	//// this test must ALWAYS be on the bottom as it is the longest-duration test
	//t.Run("/test-idle-disconnect", func(t *testing.T) {
	//	conn, err := net.Dial("tcp4", addr)
	//	require.NoError(t, err)
	//
	//	response, err := io.ReadAll(conn)
	//	require.NoError(t, err)
	//	wantResponseLine := "HTTP/1.1 408 Request Timeout\r\n"
	//	lf := bytes.IndexByte(response, '\n')
	//	require.NotEqual(t, -1, lf, "http response must contain at least one LF")
	//	require.Equal(t, wantResponseLine, string(response[:lf+1]))
	//})
	//
	//t.Run("/ctx-value", func(t *testing.T) {
	//	resp, err := stdhttp.DefaultClient.Get(URL + "/ctx-value")
	//	require.NoError(t, err)
	//	defer func() {
	//		_ = resp.Body.Close()
	//	}()
	//	require.Equal(t, stdhttp.StatusOK, resp.StatusCode)
	//})
	//
	//stdhttp.DefaultClient.CloseIdleConnections()
	//runningServer.Wait(t, 3*time.Second)
}

//func sendSimpleRequest(t *testing.T, path string, addr string) {
//	conn, err := net.Dial("tcp", addr)
//	require.NoError(t, err)
//
//	request := fmt.Sprintf("GET %s HTTP/1.1\r\n\r\n", path)
//
//	_, err = conn.Write([]byte(request))
//	require.NoError(t, err)
//	buff := make([]byte, 1)
//	_, err = conn.Read(buff)
//	require.NoError(t, err)
//	require.Equal(t, buff, []byte{'a'})
//	_, err = conn.Write([]byte(testRequestBody))
//	require.NoError(t, err)
//
//	_ = conn.Close()
//}

func contains(strs []string, substr string) bool {
	for _, str := range strs {
		if str == substr {
			return true
		}
	}

	return false
}
