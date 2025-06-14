package testutil

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/codecutil"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/internal/protocol/http1"
	"github.com/indigo-web/indigo/transport/dummy"
)

type Request struct {
	http.Request
	Body string
}

func ParseRequest(data string) (Request, error) {
	client := dummy.NewCircularClient([]byte(data)).OneTime()
	request := construct.Request(config.Default(), client)
	suit := http1.New(config.Default(), nil, client, request, codecutil.NewCache[http.Decompressor](nil))
	request.Body = http.NewBody(config.Default(), suit)

	for {
		done, extra, err := suit.Parse([]byte(data))
		if err != nil {
			return Request{}, err
		}

		client.Pushback(extra)

		if done {
			break
		}
	}

	request.Body.Reset(request)
	if err := suit.Reset(request); err != nil {
		return Request{}, err
	}

	body, err := request.Body.String()

	return Request{
		Request: *request,
		Body:    body,
	}, err
}
