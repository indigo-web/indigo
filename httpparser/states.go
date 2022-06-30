package httpparser

type (
	parsingState     uint8
	chunkedBodyState uint8
)

const (
	messageBegin parsingState = iota + 1
	method
	path
	protocol
	protocolCR
	protocolLF
	headerKey
	headerColon
	headerValue
	headerValueCR
	headerValueLF
	headerValueDoubleCR
	body
	bodyConnectionClose

	dead
)

const (
	chunkLength chunkedBodyState = iota + 1
	chunkLengthCR

	chunkBody
	chunkBodyEnd
	chunkBodyCR

	lastChunk
	lastChunkCR

	transferCompleted
)
