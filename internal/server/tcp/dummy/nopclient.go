package dummy

import (
	"github.com/indigo-web/indigo/internal/server/tcp"
)

func NewNopClient() tcp.Client {
	return NewCircularClient(nil)
}
