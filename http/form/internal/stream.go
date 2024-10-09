package internal

import (
	"github.com/indigo-web/indigo/internal/strutil"
	"strings"
)

// TODO: make Stream a newtype of the string

type stream struct {
	data string
}

func newStream(data string) stream {
	return stream{data}
}

func (s *stream) Find(char byte) int {
	return strings.IndexByte(s.data, char)
}

func (s *stream) FindSubstr(str string) int {
	for {
		begin := s.Find(str[0])
		if begin == -1 {
			return -1
		}

		if s.Compare(begin, str) {
			return begin
		}

		s.Advance(1)
	}
}

func (s *stream) Compare(offset int, str string) bool {
	if len(s.data) < len(str)+offset {
		return false
	}

	return s.data[offset:offset+len(str)] == str
}

func (s *stream) CompareFold(offset int, str string) bool {
	if len(s.data) < len(str)+offset {
		return false
	}

	return strutil.CmpFold(s.data[offset:offset+len(str)], str)
}

func (s *stream) Consume(str string) bool {
	if s.Compare(0, str) {
		s.Advance(len(str))
		return true
	}

	return false
}

func (s *stream) ConsumeFold(str string) bool {
	if s.CompareFold(0, str) {
		s.Advance(len(str))
		return true
	}

	return false
}

func (s *stream) Advance(n int) (leftBehind string) {
	leftBehind, s.data = s.data[:n], s.data[n:]
	return leftBehind
}

func (s *stream) AdvanceExclusively(n int) (leftBehind string) {
	leftBehind = s.Advance(n + 1)
	return leftBehind[:len(leftBehind)-1]
}

func (s *stream) AdvanceLine() (leftBehind string, ok bool) {
	newline := s.Find('\n')
	if newline == -1 {
		return "", false
	}

	leftBehind = s.AdvanceExclusively(newline)
	if leftBehind[len(leftBehind)-1] == '\r' {
		return leftBehind[:len(leftBehind)-1], true
	}

	return leftBehind, true
}

func (s *stream) SkipWhitespaces() {
	s.data = strutil.LStripWS(s.data)
}

func (s *stream) Empty() bool {
	return len(s.data) == 0
}

func (s *stream) Expose() string {
	return s.data
}
