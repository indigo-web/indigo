package codec

import (
	"testing"
)

func TestFlate(t *testing.T) {
	testCodec(t, NewDeflate().New())
}
