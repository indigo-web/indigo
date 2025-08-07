package parse

import (
	"github.com/indigo-web/indigo/config"
	"github.com/indigo-web/indigo/http"
	"github.com/indigo-web/indigo/internal/codecutil"
	"github.com/indigo-web/indigo/internal/construct"
	"github.com/indigo-web/indigo/internal/protocol/http1"
	"github.com/indigo-web/indigo/transport/dummy"
)

func HTTP11Request(data string) (*http.Request, error) {
	client := dummy.NewMockClient([]byte(data))
	request := construct.Request(config.Default(), client)
	suit := http1.New(config.Default(), nil, client, request, codecutil.NewCache(nil))
	request.Body = http.NewBody(config.Default(), suit)

	for {
		done, extra, err := suit.Parse([]byte(data))
		if err != nil {
			return nil, err
		}

		client.Pushback(extra)

		if done {
			break
		}
	}

	request.Body.Reset(request)
	suit.Reset(request)

	return request, nil
}
