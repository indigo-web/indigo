package dummy

import (
	"github.com/fakefloordiv/indigo/internal/server/tcp"
)

func NewNopClient() tcp.Client {
	return NewCircularClient([]byte("\x00"))
}
