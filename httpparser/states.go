package httpparser

type (
	parsingState     uint8
	chunkedBodyState uint8
)

const (
	eMessageBegin parsingState = iota + 1
	eMethod
	ePath
	eProtocol
	eProtocolCR
	eProtocolLF
	eHeaderKey
	eHeaderColon
	eHeaderValue
	eHeaderValueCR
	eHeaderValueLF
	eHeaderValueDoubleCR
	eBody
	eBodyConnectionClose

	eDead
)

const (
	eChunkLength chunkedBodyState = iota + 1
	eChunkLengthCR

	eChunkBody
	eChunkBodyEnd
	eChunkBodyCR

	eLastChunk
	eLastChunkCR

	eTransferCompleted
)
