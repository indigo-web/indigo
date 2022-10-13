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
	eFragmentDecode1Char
	eFragmentDecode2Char
	eProto
	eH
	eHT
	eHTT
	eHTTP
	eProtoMajor
	eProtoMinor
	eProtoCR
	eProtoCRLF
	eProtoCRLFCR
	eHeaderKey
	eHeaderColon
	eContentLength
	eContentLengthCR
	eContentLengthCRLF
	eContentLengthCRLFCR
	eHeaderValue
	eHeaderValueComma
	eHeaderValueBackslash
	eHeaderValueQuoted
	eHeaderValueQuotedBackslash
	eHeaderValueCR
	eHeaderValueCRLF
	eHeaderValueCRLFCR
	eBody
)

type chunkedBodyParserState uint8

const (
	eChunkLength1Char chunkedBodyParserState = iota + 1
	eChunkLength
	eChunkLengthCR
	eChunkLengthCRLF
	eChunkBody
	eChunkBodyCR
	eChunkBodyCRLF
	eLastChunkCR
	eFooter
	eFooterCR
	eFooterCRLF
	eFooterCRLFCR
)
