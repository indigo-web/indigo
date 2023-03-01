package dummy

import (
	"github.com/indigo-web/indigo/v2/internal/server/tcp"
)

func NewNopClient() tcp.Client {
	return NewCircularClient(nil)
}
