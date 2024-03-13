package http2

import "github.com/indigo-web/indigo/internal/tcp"

type Suit struct {
	*Parser
	client tcp.Client
}

func New(client tcp.Client) *Suit {
	return &Suit{
		client: client,
	}
}

func (s *Suit) Serve() {

}
