package httpparser

import (
	"indigo/errors"
	"indigo/types"
)

type chunkedBodyParser struct {
	callback       types.BodyWriter
	state          chunkedBodyState
	chunkLength    int
	chunkBodyBegin int

	maxChunkSize int
}

func NewChunkedBodyParser(writeBody types.BodyWriter, maxChunkSize uint) *chunkedBodyParser {
	return &chunkedBodyParser{
		callback: writeBody,
		state:    chunkLength,
		// as chunked requests aren't obligatory, we better keep the buffer unallocated until
		// we'll need it
		maxChunkSize: int(maxChunkSize),
	}
}

func (p *chunkedBodyParser) Clear() {
	p.state = chunkLength
	p.chunkLength = 0
}

func (p *chunkedBodyParser) Feed(data []byte) (done bool, extraBytes []byte, err error) {
	if p.state == transferCompleted {
		/*
			It returns extra-bytes as parser must know, that it's his job now

			But if parser is feeding again, it means only that we really need
			to parse one more chunked body
		*/
		p.Clear()
	}
	if len(data) == 0 {
		return false, nil, nil
	}

	for i, char := range data {
		switch p.state {
		case chunkLength:
			switch char {
			case '\r':
				p.state = chunkLengthCR
			case '\n':
				if p.chunkLength == 0 {
					p.state = lastChunk
					break
				}

				p.chunkBodyBegin = i + 1
				p.state = chunkBody
			default:
				// TODO: add support of trailers
				if (char < '0' && char > '9') && (char < 'a' && char > 'f') && (char < 'A' && char > 'F') {
					// non-printable ascii-character
					p.complete()

					return true, nil, errors.ErrInvalidChunkSize
				}

				p.chunkLength = (p.chunkLength << 4) + int((char&0xF)+9*(char>>6))

				if p.chunkLength > p.maxChunkSize {
					p.complete()

					return true, nil, errors.ErrTooBigChunkSize
				}
			}
		case chunkLengthCR:
			if char != '\n' {
				p.complete()

				return true, nil, errors.ErrInvalidChunkSplitter
			}

			if p.chunkLength == 0 {
				p.state = lastChunk
				break
			}

			p.chunkBodyBegin = i + 1
			p.state = chunkBody
		case chunkBody:
			p.chunkLength--

			if p.chunkLength == 0 {
				p.state = chunkBodyEnd
			}
		case chunkBodyEnd:
			p.callback(data[p.chunkBodyBegin:i])

			switch char {
			case '\r':
				p.state = chunkBodyCR
			case '\n':
				p.state = chunkLength
			default:
				p.complete()

				return true, nil, errors.ErrInvalidChunkSplitter
			}
		case chunkBodyCR:
			if char != '\n' {
				p.complete()

				return true, nil, errors.ErrInvalidChunkSplitter
			}

			p.state = chunkLength
		case lastChunk:
			switch char {
			case '\r':
				p.state = lastChunkCR
			case '\n':
				p.complete()

				return true, data[i+1:], nil
			default:
				// looks sad, received everything, and fucked up in the end
				// or this was made for special? Oh god
				p.complete()

				return true, nil, errors.ErrInvalidChunkSplitter
			}
		case lastChunkCR:
			if char != '\n' {
				p.complete()

				return true, nil, errors.ErrInvalidChunkSplitter
			}

			p.complete()

			return true, data[i+1:], nil
		}
	}

	if p.state == chunkBody {
		p.callback(data[p.chunkBodyBegin:])
	}

	p.chunkBodyBegin = 0

	return false, nil, nil
}

func (p *chunkedBodyParser) complete() {
	p.state = transferCompleted
}
