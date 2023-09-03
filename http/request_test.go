package http

import (
	"github.com/indigo-web/indigo/internal/server/tcp"
)

type dummyBodyReader struct {
	client tcp.Client
}

func newDummyReader(client tcp.Client) dummyBodyReader {
	return dummyBodyReader{
		client: client,
	}
}

func (d dummyBodyReader) Init(*Request) {}

func (d dummyBodyReader) Read() ([]byte, error) {
	return d.ReadNoDecoding()
}

func (d dummyBodyReader) ReadNoDecoding() ([]byte, error) {
	return d.client.Read()
}
