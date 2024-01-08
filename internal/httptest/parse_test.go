package httptest

import (
	"fmt"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestParse(t *testing.T) {
	t.Run("simple 200 ok", func(t *testing.T) {
		data := "HTTP/1.1 200 OK\r\n\r\n"
		request, err := Parse(data)
		require.NoError(t, err)
		for _, err := range compare(request, Request{
			Proto:   "HTTP/1.1",
			Code:    200,
			Status:  "OK",
			Headers: nil,
			Body:    "",
		}) {
			assert.NoError(t, err)
		}
	})

	t.Run("headers", func(t *testing.T) {
		data := "HTTP/1.1 200 OK\r\nHello: World\r\nFoo: bar\r\n\r\n"
		request, err := Parse(data)
		require.NoError(t, err)
		for _, err := range compare(request, Request{
			Proto:  "HTTP/1.1",
			Code:   200,
			Status: "OK",
			Headers: headers.NewFromMap(map[string][]string{
				"Hello": {"World"},
				"Foo":   {"bar"},
			}),
			Body: "",
		}) {
			assert.NoError(t, err)
		}
	})

	t.Run("body and content-length", func(t *testing.T) {
		data := "HTTP/1.1 200 OK\r\nContent-Length: 13\r\n\r\nHello, world!"
		request, err := Parse(data)
		require.NoError(t, err)
		for _, err := range compare(request, Request{
			Proto:  "HTTP/1.1",
			Code:   200,
			Status: "OK",
			Headers: headers.NewFromMap(map[string][]string{
				"Content-Length": {"13"},
			}),
			Body: "Hello, world!",
		}) {
			assert.NoError(t, err)
		}
	})

	t.Run("body with closing connection", func(t *testing.T) {
		data := "HTTP/1.1 200 OK\r\nConnection: close\r\n\r\nHello, world!"
		request, err := Parse(data)
		require.NoError(t, err)
		for _, err := range compare(request, Request{
			Proto:  "HTTP/1.1",
			Code:   200,
			Status: "OK",
			Headers: headers.NewFromMap(map[string][]string{
				"Connection": {"close"},
			}),
			Body: "Hello, world!",
		}) {
			assert.NoError(t, err)
		}
	})
}

func compare(got Request, want Request) (errs []error) {
	if got.Proto != want.Proto {
		errs = append(errs, fmt.Errorf("want protocol %s, got %s", want.Proto, got.Proto))
	}

	if got.Code != want.Code {
		errs = append(errs, fmt.Errorf("want code %d, got %d", want.Code, got.Code))
	}

	if got.Status != want.Status {
		errs = append(errs, fmt.Errorf("want status %s, got %s", want.Status, got.Status))
	}

	if want.Headers != nil {
		for _, key := range want.Headers.Keys() {
			wantValues := want.Headers.Values(key)
			gotValues := got.Headers.Values(key)
			if !cmpSlice(wantValues, gotValues) {
				errs = append(errs, fmt.Errorf("want %s for %s, got %s", wantValues, key, gotValues))
			}
		}

		if len(want.Headers.Unwrap()) != len(got.Headers.Unwrap()) {
			errs = append(errs, fmt.Errorf("got extra headers"))
		}
	}

	if got.Body != want.Body {
		errs = append(errs, fmt.Errorf("want body %s, got %s", want.Body, got.Body))
	}

	return errs
}

func cmpSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
