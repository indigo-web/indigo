package indigo

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"indigo/errors"
	methods "indigo/http/method"
	"indigo/http/proto"
	"indigo/router"
	"indigo/types"
	"net"
	"net/http"
	"net/url"
	"testing"
	"time"
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

	testFragment = "Somewhere Here"

	testRequestBody = "Hello, world!"
)

func instantlyDisconnect() {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return
	}

	conn.Close()
}

func getRouter(t *testing.T) router.Router {
	r := router.NewDefaultRouter()

	r.Get("/simple-get", func(request *types.Request) types.Response {
		require.Equal(t, methods.GET, request.Method)
		_, err := request.Query.Get("some non-existing query key")
		require.Error(t, err)
		require.Empty(t, request.Fragment)
		require.Equal(t, proto.HTTP11, request.Proto)

		return types.WithResponse
	})

	r.Get("/get-read-body", func(request *types.Request) types.Response {
		require.Contains(t, request.Headers, testHeaderKey)
		require.Equal(t, testHeaderValue, string(request.Headers[testHeaderKey]))

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

	return r
}

func TestAllCases(t *testing.T) {
	// testing everything in one function only because we do not wanna
	// run server multiple times (each time takes a bit time, in sum it
	// is a lot of time)

	r := getRouter(t)
	app := NewApp(addr)

	shutdown := make(chan bool)

	go func() {
		require.Equal(t, errors.ErrShutdown, app.Serve(r))
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
		resp, err := http.DefaultClient.Get(URL + "/simple-get")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Equal(t, "200 "+http.StatusText(http.StatusOK), resp.Status)
		require.Equal(t, "HTTP/1.1", resp.Proto)
		require.Contains(t, resp.Header, "Server")
		require.Equal(t, 1, len(resp.Header["Server"]))
		require.Equal(t, "indigo", resp.Header["Server"][0])
	})

	t.Run("/get-read-body", func(t *testing.T) {
		request := &http.Request{
			Method: http.MethodGet,
			URL: &url.URL{
				Scheme: "http",
				Host:   addr,
				Path:   "/get-read-body",
			},
			Proto:      "HTTP/1.1",
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: http.Header{
				"Hello": {"World!"},
			},
			Host:       addr,
			RemoteAddr: addr,
		}
		resp, err := http.DefaultClient.Do(request)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("/with-query", func(t *testing.T) {
		resp, err := http.DefaultClient.Get(URL + "/with-query?hel+lo=wor+ld")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("/read-body", func(t *testing.T) {
		body := bytes.Buffer{}
		body.Write([]byte(testRequestBody))
		resp, err := http.DefaultClient.Post(URL+"/read-body", "", &body)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("/do-not-read-body", func(t *testing.T) {
		body := bytes.Buffer{}
		body.Write([]byte(testRequestBody))
		resp, err := http.DefaultClient.Post(URL+"/read-body", "", &body)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	http.DefaultClient.CloseIdleConnections()

	// at this point server is supposed to be closed, but who knows, who knows
}
