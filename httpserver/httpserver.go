package httpserver

import (
	"indigo/http"
	"indigo/router"
	"indigo/types"
	"net"
)

/*
HTTP server is a second and core layer. It receives data from tcp server,
delegates it to parser, parser fills parsed data directly into the request
structure, and then Session struct finally delegates request to a Router
(if request has been received completely)
*/

type client struct {
	request types.Request
	parser  http.Parser
	router  router.Router
}

func (c *client) HandleData(conn net.Conn, data []byte) error {
	done, err := c.parser.Parse(&c.request, data)

	if err != nil {
		c.router.OnError(err)
		return err
	}

	if done {
		err = c.router.OnRequest(&c.request, func(b []byte) error {
			_, err = conn.Write(b)

			return err
		})

		if err != nil {
			c.router.OnError(err)
			return err
		}
	}

	return nil
}
