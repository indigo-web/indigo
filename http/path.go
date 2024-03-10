package http

import "github.com/indigo-web/utils/uf"

type Path = string

func Escape(p Path) string {
	var (
		buff   []byte
		offset int
	)

	for i := range p {
		if !isASCIIPrintable(p[i]) {
			if buff == nil {
				buff = allocBuff(len(p))
			}

			buff = append(buff, p[offset:i]...)
			escaped := escapeByte(p[i])
			buff = append(buff, '\\', escaped)
			offset = i + 1
		}
	}

	if len(buff) == 0 {
		return p
	}

	return uf.B2S(append(buff, p[offset:]...))
}

func isASCIIPrintable(c byte) bool {
	return c >= 0x20 && c <= 0x7e
}

var escapeTable = [256]byte{
	0x0:  '0',
	0x1:  '?',
	0x2:  '?',
	0x3:  '?',
	0x4:  '?',
	0x5:  '?',
	0x6:  '?',
	0x7:  'a',
	0x8:  'b',
	0x9:  't',
	0xA:  'n',
	0xB:  'v',
	0xC:  'f',
	0xD:  'r',
	0xE:  '?',
	0xF:  '?',
	0x10: '?',
	0x11: '?',
	0x12: '?',
	0x13: '?',
	0x14: '?',
	0x15: '?',
	0x16: '?',
	0x17: '?',
	0x18: '?',
	0x19: '?',
	0x1A: '?',
	0x1B: '?',
	0x1C: '?',
	0x1D: '?',
	0x1E: '?',
	0x1F: '?',
}

func escapeByte(b byte) byte {
	if b < 0x7f {
		return escapeTable[b]
	}

	return '?'
}

func allocBuff(strsize int) []byte {
	if strsize <= 25 {
		return make([]byte, 0, 40)
	}

	return make([]byte, 0, strsize+strsize/2)
}
