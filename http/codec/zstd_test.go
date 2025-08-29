package codec

import "testing"

func TestZSTD(t *testing.T) {
	testCodec(t, NewZSTD().New())
}
