package parser

type Parser interface {
	Parse(data []byte) (headersParsed bool, rest []byte, err error)
}
