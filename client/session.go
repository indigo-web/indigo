package client

import "github.com/indigo-web/indigo/client/internal/connection"

type Session struct {
	// TODO: add Cookies here
	conns connection.Manager
}
