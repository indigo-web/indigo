package http1

type parserState uint8

const (
	eMethod parserState = iota + 1
	ePath
	ePathDecode1Char
	ePathDecode2Char
	eQuery
	eQueryDecode1Char
	eQueryDecode2Char
	eFragment
	eProto
	eH
	eHT
	eHTT
	eHTTP
	eProtoMajor
	eProtoDot
	eProtoMinor
	eProtoEnd
	eProtoCR
	eProtoCRLF
	eProtoCRLFCR
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
