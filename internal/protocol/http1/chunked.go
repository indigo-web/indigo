package http1

import (
	"bytes"
	"io"
	"strings"

	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/internal/hexconv"
)

type chunkedParserState uint8

const (
	eChunkLength chunkedParserState = iota
	eChunkExt
	eChunkLengthCR
	eChunkBody
	eChunkBodyDone
	eChunkBodyCRLF
	eChunkTrailer
	eChunkTrailerCRLF
	eChunkTrailerFieldLine
)

// maxChunkLengthDigits sets the implicit limit of a single chunk length to 4GiB, which
// is supposedly should be enough.
const maxChunkLengthDigits = 8

type chunkedParser struct {
	state        chunkedParserState
	lengthDigits uint8
	chunkLength  uint64
}

func newChunkedParser() chunkedParser {
	return chunkedParser{state: eChunkLength}
}

// Parse returns a chunk when it's ready, nil otherwise. io.EOF signals that the body
// is complete. The parser resets automatically.
func (c *chunkedParser) Parse(data []byte) (chunk, extra []byte, err error) {
	switch c.state {
	case eChunkLength:
		goto chunkLength
	case eChunkExt:
		goto chunkExt
	case eChunkLengthCR:
		goto chunkLengthCR
	case eChunkBody:
		goto chunkBody
	case eChunkBodyDone:
		goto chunkBodyDone
	case eChunkBodyCRLF:
		goto chunkBodyCRLF
	case eChunkTrailer:
		goto trailer
	case eChunkTrailerCRLF:
		goto chunkTrailerCRLF
	case eChunkTrailerFieldLine:
		goto chunkTrailerFieldLine
	default:
		panic("unreachable code")
	}

chunkLength:
	for i := 0; i < len(data); i++ {
		switch char := data[i]; char {
		case '\r':
			data = data[i+1:]
			goto chunkLengthCR
		case '\n':
			data = data[i:]
			goto chunkLengthCR
		case ';':
			data = data[i+1:]
			goto chunkExt
		default:
			val := hexconv.Halfbyte[char]
			if val == 0xFF {
				return nil, nil, status.ErrBadChunk
			}

			c.chunkLength = (c.chunkLength << 4) | uint64(val)
			if c.lengthDigits++; c.lengthDigits > maxChunkLengthDigits {
				return nil, nil, status.ErrBadChunk
			}
		}
	}

	c.state = eChunkLength
	return nil, nil, nil

chunkExt:
	{
		// currently no chunk extensions are supported, therefore completely ignored.
		boundary := bytes.IndexByte(data, '\n')
		if boundary == -1 {
			c.state = eChunkExt
			return nil, nil, nil
		}

		data = data[boundary+1:]
		if c.chunkLength == 0 {
			goto trailer
		}

		goto chunkBody
	}

chunkLengthCR:
	if len(data) == 0 {
		c.state = eChunkLengthCR
		return nil, nil, nil
	}

	if data[0] != '\n' {
		return nil, nil, status.ErrBadChunk
	}

	data = data[1:]

	if c.chunkLength == 0 {
		goto trailer
	}

	goto chunkBody

chunkBody:
	{
		n := min(c.chunkLength, uint64(len(data)))
		c.chunkLength -= n
		chunk = data[:n]

		if c.chunkLength == 0 {
			c.state = eChunkBodyDone
		} else {
			c.state = eChunkBody
		}

		return chunk, data[n:], nil
	}

chunkBodyDone:
	// omit len(data) == 0 check, as we only jump here from the dispatch, which in turn is executed
	// on new data, which is implied to never be empty.
	c.lengthDigits = 0
	switch data[0] {
	case '\r':
		data = data[1:]
		goto chunkBodyCRLF
	case '\n':
		data = data[1:]
		goto chunkLength
	default:
		return nil, nil, status.ErrBadChunk
	}

chunkBodyCRLF:
	if len(data) == 0 {
		c.state = eChunkBodyCRLF
		return nil, nil, nil
	}

	if data[0] != '\n' {
		return nil, nil, status.ErrBadChunk
	}

	data = data[1:]
	goto chunkLength

trailer:
	if len(data) == 0 {
		c.state = eChunkTrailer
		return nil, nil, nil
	}

	switch data[0] {
	case '\r':
		data = data[1:]
		goto chunkTrailerCRLF
	case '\n':
		c.state = eChunkLength
		return nil, data[1:], io.EOF
	default:
		// we've got some field lines
		goto chunkTrailerFieldLine
	}

chunkTrailerCRLF:
	if len(data) == 0 {
		c.state = eChunkTrailerCRLF
		return nil, nil, nil
	}

	if data[0] != '\n' {
		return nil, nil, status.ErrBadChunk
	}

	c.state = eChunkLength
	return nil, data[1:], io.EOF

chunkTrailerFieldLine:
	{
		boundary := bytes.IndexByte(data, '\n')
		if boundary == -1 {
			c.state = eChunkTrailerFieldLine
			return nil, nil, nil
		}

		data = data[boundary+1:]
		goto trailer
	}
}

var (
	// chunkExtZeroFill is used to fill the gap between chunk length and chunk content. The count
	// 64/4 represents 64 bits - the maximal uint size, and 4 - bits per hex value, therefore
	// resulting in 15 characters (plus semicolon) total.
	chunkExtZeroFill = ";" + strings.Repeat("0", 64/4-1)
	chunkZeroTrailer = []byte("0\r\n\r\n")
)
