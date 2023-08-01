package http1

type parserState uint8

const (
	eMethod parserState = iota + 1
	ePath
	eHeaderKey
	eContentLength
	eContentLengthCR
	eHeaderValue
	eHeaderValueCRLFCR
)

type chunkedBodyParserState uint8

const (
	eChunkLength1Char chunkedBodyParserState = iota + 1
	eChunkLength
	eChunkLengthCR
	eChunkLengthCRLF
	eChunkBody
	eChunkBodyEnd
	eChunkBodyCR
	eChunkBodyCRLF
	eLastChunkCR
	eFooter
	eFooterCR
	eFooterCRLF
	eFooterCRLFCR
)
