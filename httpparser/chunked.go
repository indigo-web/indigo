package httpparser

import (
	"indigo/errors"
	"indigo/internal"
	"io"
)

type chunkedBodyParser struct {
	pipe           internal.Pipe
	state          chunkedBodyState
	chunkLength    int
	chunkBodyBegin int

	maxChunkSize int
}

func NewChunkedBodyParser(pipe internal.Pipe, maxChunkSize uint) *chunkedBodyParser {
	return &chunkedBodyParser{
		pipe:  pipe,
		state: eChunkLength,
		// as chunked requests aren't obligatory, we better keep the buffer unallocated until
		// we'll need it
		maxChunkSize: int(maxChunkSize),
	}
}

func (p *chunkedBodyParser) Clear() {
	p.state = eChunkLength
	p.chunkLength = 0
}

func (p *chunkedBodyParser) Feed(data []byte) (done bool, extraBytes []byte, err error) {
	if p.state == eTransferCompleted {
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
		case eChunkLength:
			switch char {
			case '\r':
				p.state = eChunkLengthCR
			case '\n':
				if p.chunkLength == 0 {
					p.state = eLastChunk
					break
				}

				p.chunkBodyBegin = i + 1
				p.state = eChunkBody
			default:
				// TODO: add support of trailers
				if (char < '0' && char > '9') && (char < 'a' && char > 'f') && (char < 'A' && char > 'F') {
					// non-printable ascii-character
					p.complete()
					p.pipe.WriteErr(errors.ErrParsingRequest)

					return true, nil, errors.ErrInvalidChunkSize
				}

				p.chunkLength = (p.chunkLength << 4) + int((char&0xF)+9*(char>>6))

				if p.chunkLength > p.maxChunkSize {
					p.complete()
					p.pipe.WriteErr(errors.ErrParsingRequest)

					return true, nil, errors.ErrTooBigChunkSize
				}
			}
		case eChunkLengthCR:
			if char != '\n' {
				p.complete()
				p.pipe.WriteErr(errors.ErrParsingRequest)

				return true, nil, errors.ErrInvalidChunkSplitter
			}

			if p.chunkLength == 0 {
				p.state = eLastChunk
				break
			}

			p.chunkBodyBegin = i + 1
			p.state = eChunkBody
		case eChunkBody:
			p.chunkLength--

			if p.chunkLength == 0 {
				p.state = eChunkBodyEnd
			}
		case eChunkBodyEnd:
			p.pipe.Write(data[p.chunkBodyBegin:i])

			switch char {
			case '\r':
				p.state = eChunkBodyCR
			case '\n':
				p.state = eChunkLength
			default:
				p.complete()
				p.pipe.WriteErr(errors.ErrParsingRequest)

				return true, nil, errors.ErrInvalidChunkSplitter
			}
		case eChunkBodyCR:
			if char != '\n' {
				p.complete()
				p.pipe.WriteErr(errors.ErrParsingRequest)

				return true, nil, errors.ErrInvalidChunkSplitter
			}

			p.state = eChunkLength
		case eLastChunk:
			switch char {
			case '\r':
				p.state = eLastChunkCR
			case '\n':
				p.complete()
				p.pipe.WriteErr(io.EOF)

				return true, data[i+1:], nil
			default:
				// looks sad, received everything, and fucked up in the end
				// or this was made for special? Oh god
				p.complete()
				p.pipe.WriteErr(errors.ErrParsingRequest)

				return true, nil, errors.ErrInvalidChunkSplitter
			}
		case eLastChunkCR:
			if char != '\n' {
				p.complete()
				p.pipe.WriteErr(errors.ErrParsingRequest)

				return true, nil, errors.ErrInvalidChunkSplitter
			}

			p.complete()
			p.pipe.WriteErr(io.EOF)

			return true, data[i+1:], nil
		}
	}

	if p.state == eChunkBody {
		p.pipe.Write(data[p.chunkBodyBegin:])
	}

	p.chunkBodyBegin = 0

	return false, nil, nil
}

func (p *chunkedBodyParser) complete() {
	p.state = eTransferCompleted
}
