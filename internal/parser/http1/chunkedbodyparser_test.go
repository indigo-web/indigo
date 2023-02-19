package http1

import (
	"io"
	"testing"

	"github.com/indigo-web/indigo/http/status"
	"github.com/indigo-web/indigo/settings"
	"github.com/stretchr/testify/require"
)

func feedParserWithPart(
	parser chunkedBodyParser, part []byte, trailer bool,
) (p chunkedBodyParser, result []byte, err error) {
	var bodyPart []byte

	for len(part) > 0 {
		bodyPart, part, err = parser.Parse(part, trailer)
		if err != nil {
			return parser, result, err
		}

		result = append(result, bodyPart...)
	}

	return parser, result, nil
}

func testDifferentPartSizes(t *testing.T, request []byte, wantBody string, trailer bool) {
	for i := 1; i < len(request); i += 1 {
		parser := newChunkedBodyParser(settings.Default().Body)
		parts := splitIntoParts(request, i)
		var (
			content []byte
			piece   []byte
			err     error
		)

		for j, part := range parts {
			parser, piece, err = feedParserWithPart(parser, part, trailer)
			switch err {
			case nil:
				content = append(content, piece...)
			case io.EOF:
				require.Equal(t, wantBody, string(content))
				require.Equal(t, j, len(parts)-1, "Expected all the parts to be parsed")
				return
			default:
				require.NoErrorf(t, err, "Parts size: %d", i)
			}
		}
	}
}

func TestChunkedBodyParser_Parse(t *testing.T) {
	chunked := []byte("d\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n")
	wantedBody := "Hello, world!But what's wrong with you?Finally am here"
	testDifferentPartSizes(t, chunked, wantedBody, false)
}

func TestChunkedBodyParser_Parse_LFOnly(t *testing.T) {
	chunked := []byte("d\nHello, world!\n1a\nBut what's wrong with you?\nf\nFinally am here\n0\n\n")
	wantedBody := "Hello, world!But what's wrong with you?Finally am here"
	testDifferentPartSizes(t, chunked, wantedBody, false)
}

func TestChunkedBodyParser_Parse_FooterHeaders(t *testing.T) {
	chunked := []byte("7\r\nMozilla\r\n9\r\nDeveloper\r\n7\r\nNetwork\r\n0\r\nExpires: date here\r\n\r\n")
	wantedBody := "MozillaDeveloperNetwork"
	testDifferentPartSizes(t, chunked, wantedBody, true)
}

func TestChunkedBodyParser_Parse_Negative(t *testing.T) {
	t.Run("BeginWithCRLF", func(t *testing.T) {
		chunked := []byte("\r\nd\r\nHello, world!\r\n1a\r\nBut what's wrong with you?\r\nf\r\nFinally am here\r\n0\r\n\r\n")

		parser := newChunkedBodyParser(settings.Default().Body)
		piece, extra, err := parser.Parse(chunked, false)
		require.Empty(t, piece)
		require.Empty(t, extra)
		require.EqualError(t, err, status.ErrBadRequest.Error())
	})
}
