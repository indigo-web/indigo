package queryparser

type queryParserState uint8

const (
	eKey queryParserState = iota + 1
	eValue
)
