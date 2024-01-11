package httptest

import (
	"errors"
	"fmt"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/internal/keyvalue"
	"strconv"
	"strings"
)

var errNotPresented = errors.New("no values are presented")

type Request struct {
	Method  method.Method
	Path    string
	Proto   string
	Headers headers.Headers
	Body    string
}

func NewRequest() Request {
	return Request{
		Headers: keyvalue.New(),
	}
}

func Parse(raw string) (request Request, err error) {
	var found bool
	request = NewRequest()

	m, rest, ok := strings.Cut(raw, " ")
	if !ok || len(rest) == 0 {
		return request, fmt.Errorf("bad request line: lacking method")
	}
	request.Method = method.Parse(m)
	raw = rest

	request.Path, raw, found = strings.Cut(raw, " ")
	if !found || len(raw) == 0 {
		return request, fmt.Errorf("bad request line: lacking path")
	}

	request.Proto, raw, found = strings.Cut(raw, "\r\n")
	if !found || len(raw) == 0 {
		return request, fmt.Errorf("bad request: lacking finalizing CRLF")
	}

	for {
		var headerLine string
		headerLine, raw, found = strings.Cut(raw, "\r\n")
		if len(headerLine) == 0 {
			break
		}
		if !found {
			return request, fmt.Errorf("bad header line %s: no breaking CRLF", headerLine)
		}

		key, value, err := parseHeaderLine(headerLine)
		if err != nil {
			return request, err
		}

		request.Headers.Add(key, value)
	}

	length, err := getContentLength(request.Headers.Values("content-length"))
	switch err {
	case nil:
	case errNotPresented:
		if connection := request.Headers.Value("connection"); connection == "close" {
			length = len(raw)
			break
		}

		return request, fmt.Errorf("no content-length is presented")
	default:
		return request, fmt.Errorf("bad content-length: %s", err)
	}

	if length != len(raw) {
		return request, fmt.Errorf(
			"bad body: content-length=%d, actual=%d",
			length, len(raw),
		)
	}

	request.Body = raw

	return request, nil
}

func getContentLength(values []string) (int, error) {
	switch len(values) {
	case 0:
		return 0, errNotPresented
	case 1:
		return strconv.Atoi(values[0])
	default:
		return 0, fmt.Errorf("too many values")
	}
}

func parseHeaderLine(line string) (key, value string, err error) {
	var found bool
	key, value, found = strings.Cut(line, ": ")
	if !found {
		return "", "", fmt.Errorf("bad header %s: no value", line)
	}

	if len(line) == 0 {
		return "", "", fmt.Errorf("bad header %s: empty value", key)
	}

	return key, value, nil
}

func Compare(got Request, want Request) (errs []error) {
	if got.Method != want.Method {
		errs = append(errs, fmt.Errorf("want method %s, got %s", want.Method, got.Method))
	}

	if got.Path != want.Path {
		errs = append(errs, fmt.Errorf("want path %s, got %s", want.Path, got.Path))
	}

	if got.Proto != want.Proto {
		errs = append(errs, fmt.Errorf("want protocol %s, got %s", want.Proto, got.Proto))
	}

	if want.Headers != nil {
		for _, key := range want.Headers.Keys() {
			wantValues := want.Headers.Values(key)
			gotValues := got.Headers.Values(key)
			if !cmpSlice(wantValues, gotValues) {
				errs = append(errs, fmt.Errorf("want %s for %s, got %s", wantValues, key, gotValues))
			}
		}

		if len(want.Headers.Keys()) != len(got.Headers.Keys()) {
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
