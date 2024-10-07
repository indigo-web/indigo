package internal

import (
	"github.com/indigo-web/indigo/internal/strcmp"
	"strings"
)

type Stream struct {
	position int
	data     string
}

func NewStream(data string) Stream {
	return Stream{0, data}
}

func (s *Stream) Find(char byte) int {
	return strings.IndexByte(s.data, char)
}

func (s *Stream) FindSubstr(str string) int {
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

func (s *Stream) Compare(offset int, str string) bool {
	if len(s.data) < len(str)+offset {
		return false
	}

	return s.data[offset:offset+len(str)] == str
}

func (s *Stream) CompareFold(offset int, str string) bool {
	if len(s.data) < len(str)+offset {
		return false
	}

	return strcmp.Fold(s.data[offset:offset+len(str)], str)
}

func (s *Stream) Consume(str string) bool {
	if s.Compare(0, str) {
		s.Advance(len(str))
		return true
	}

	return false
}

func (s *Stream) ConsumeFold(str string) bool {
	if s.CompareFold(0, str) {
		s.Advance(len(str))
		return true
	}

	return false
}

func (s *Stream) Advance(n int) (leftBehind string) {
	leftBehind, s.data, s.position = s.data[:n], s.data[n:], s.position+n
	return leftBehind
}

func (s *Stream) AdvanceExclusively(n int) (leftBehind string) {
	leftBehind = s.Advance(n + 1)
	return leftBehind[:len(leftBehind)-1]
}

func (s *Stream) AdvanceLine() (leftBehind string, ok bool) {
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

func (s *Stream) SkipWhitespaces() {
	for i, c := range s.data {
		switch c {
		case ' ', '\t':
		default:
			s.Advance(i)
			return
		}
	}

	s.Advance(len(s.data))
}

func (s *Stream) Empty() bool {
	return len(s.data) == 0
}

func (s *Stream) Expose() string {
	return s.data
}

func (s *Stream) Pos() int {
	return s.position
}
