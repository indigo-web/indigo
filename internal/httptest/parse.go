package httptest

import (
	"fmt"
	"github.com/indigo-web/chunkedbody"
	"github.com/indigo-web/indigo/http/headers"
	"github.com/indigo-web/indigo/internal/datastruct"
	"github.com/indigo-web/utils/uf"
	"strconv"
	"strings"
)

type Request struct {
	Proto   string
	Code    int
	Status  string
	Headers headers.Headers
	Body    string
}

func NewRequest() Request {
	return Request{
		Headers: datastruct.NewKeyValue(),
	}
}

func Parse(raw string) (request Request, err error) {
	var found bool
	request = NewRequest()

	request.Proto, raw, found = strings.Cut(raw, " ")
	if !found || len(raw) == 0 {
		return request, fmt.Errorf("bad request line: lacking code and status")
	}

	var code string
	code, raw, found = strings.Cut(raw, " ")
	request.Code, err = strconv.Atoi(code)
	if err != nil {
		return request, err
	}

	if !found || len(raw) == 0 {
		return request, fmt.Errorf("bad request line: lacking status code")
	}

	request.Status, raw, found = strings.Cut(raw, "\r\n")
	if !found || len(raw) == 0 {
		return request, fmt.Errorf("bad request: only request line is presented")
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

	request.Body, err = processBody(request, raw)

	return request, err
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

func processBody(request Request, data string) (string, error) {
	if request.Headers.Value("connection") == "close" {
		return data, nil
	}

	te := request.Headers.Values("transfer-encoding")
	if len(te) > 0 {
		if len(te) != 1 || te[0] != "chunked" {
			return "", fmt.Errorf("httptest: cannot process encodings: %s", strings.Join(te, ","))
		}

		_, hasTrailer := request.Headers.Get("trailer")

		return processChunkedBody(data, hasTrailer)
	}

	contentLengths := request.Headers.Values("content-length")
	switch len(contentLengths) {
	case 0:
		if len(data) == 0 {
			return "", nil
		}

		return "", fmt.Errorf("bad request: neither Transfer-Encoding or Content-Length are presented")
	case 1:
		length, err := strconv.Atoi(contentLengths[0])
		if err != nil {
			return "", err
		}

		return processPlainBody(data, length)
	default:
		return "", fmt.Errorf(
			"bad request: too many content-lengths: %s", strings.Join(contentLengths, ", "),
		)
	}
}

func processChunkedBody(data string, trailer bool) (string, error) {
	var buff []byte
	parser := chunkedbody.NewParser(chunkedbody.DefaultSettings())

	for len(data) > 0 {
		chunk, extra, err := parser.Parse(uf.S2B(data), trailer)
		if err != nil {
			return "", fmt.Errorf("bad request: bad chunked body: %s", err)
		}

		buff = append(buff, chunk...)
		data = string(extra)
	}

	return string(buff), nil
}

func processPlainBody(data string, length int) (string, error) {
	if len(data) > length {
		return "", fmt.Errorf("got extra body. Please note: no pipelining is supported yet")
	}

	return data, nil
}
