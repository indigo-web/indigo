package http1

import (
	"fmt"
	"github.com/indigo-web/indigo/client"
	"github.com/indigo-web/indigo/http/method"
	"github.com/indigo-web/indigo/http/proto"
	"net"
)

type Renderer struct {
	conn net.Conn
	buff []byte
}

func NewRenderer(conn net.Conn, buff []byte) *Renderer {
	return &Renderer{
		conn: conn,
		buff: buff,
	}
}

func (r *Renderer) Send(req *client.Request) error {
	if err := r.renderProtocol(req.Proto); err != nil {
		return err
	}
}

func (r *Renderer) renderMethod(m method.Method) {
	
}

func (r *Renderer) renderProtocol(protocol proto.Proto) error {
	protocolBytes := proto.ToBytes(protocol)
	if protocolBytes == nil {
		return fmt.Errorf("BUG: http1 render: unknown protocol: %v", protocol)
	}

	r.buff = append(r.buff, protocolBytes...)
	return nil
}
