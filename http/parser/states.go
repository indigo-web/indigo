package parser

type (
	RequestState     uint8
	parsingState     uint8
	chunkedBodyState uint8
)

const (
	Pending RequestState = 1 << iota
	RequestCompleted
	BodyCompleted
	Error
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
