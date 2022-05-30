package http

import "indigo/types"

type Parser interface {
	Parse(requestStruct *types.Request, data []byte) (done bool, err error)
}

type portedSnowdropParser struct {
	// TODO: implement a parser that is just a port for
	//       snowdrop (at first)
}

func (p portedSnowdropParser) Parse(rs *types.Request, data []byte) (done bool, err error) {
	panic("not implemented")
}

// TODO: actually, my idea is to wright own embedded parser as it
//       will be a lot faster than snowdrop (that aims at wider
//       community of guys who will use it I can't even imagine how)
