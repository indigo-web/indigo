package client

import (
	"github.com/indigo-web/indigo/client/internal/render"
	"github.com/indigo-web/indigo/internal/server/tcp"
)

type Session struct {
	// TODO: add Cookies here
	conn   tcp.Client
	render *render.Renderer
}
