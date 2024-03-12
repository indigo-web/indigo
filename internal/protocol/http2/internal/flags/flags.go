package flags

import "strings"

type Flag byte

const (
	Ack        Flag = 0x01
	EndStream  Flag = 0x01
	EndHeaders Flag = 0x04
	Padded     Flag = 0x08
	Priority   Flag = 0x20
)

func (f Flag) String() string {
	var flags []string

	if f&Ack != 0 {
		flags = append(flags, "ACK")
	}
	if f&EndStream != 0 {
		flags = append(flags, "END_STREAM")
	}
	if f&EndHeaders != 0 {
		flags = append(flags, "END_HEADERS")
	}
	if f&Padded != 0 {
		flags = append(flags, "PADDED")
	}
	if f&Priority != 0 {
		flags = append(flags, "PRIORITY")
	}

	return strings.Join(flags, ",")
}
