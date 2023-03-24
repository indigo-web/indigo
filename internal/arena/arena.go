package arena

// Arena is simply a big slice static-sized slice. It behaves just like built-in append()
// function, except encapsulating its internal implementation, so no slice is being returned
// to user - just boolean flag whether newly appended data exceeds size limits.
type Arena struct {
	memory     []byte
	begin, pos int

	maxSize int
}

func NewArena(initialSpace, maxSpace int) Arena {
	return Arena{
		memory:  make([]byte, initialSpace),
		maxSize: maxSpace,
	}
}

// Append appends bytes to a buffer. In case of exceeding the maximal size, false is returned
// and data isn't written
func (a *Arena) Append(chars []byte) (ok bool) {
	if a.pos+len(chars) > len(a.memory) {
		if len(a.memory)+len(chars) >= a.maxSize {
			return false
		}

		copy(a.memory[a.pos:], chars)
		a.memory = append(a.memory, chars[len(a.memory)-a.pos:]...)
		a.pos += len(chars)

		return true
	}

	copy(a.memory[a.pos:], chars)
	a.pos += len(chars)

	return true
}

func (a *Arena) Finish() []byte {
	segment := a.memory[a.begin:a.pos]
	a.begin = a.pos

	return segment
}

func (a *Arena) Clear() {
	a.begin = 0
	a.pos = 0
}
