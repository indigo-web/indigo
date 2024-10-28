package coding

import "io"

type decompressor struct {
	Name string
	R    io.Reader
}

type Decompressor struct {
	available []decompressor
}
