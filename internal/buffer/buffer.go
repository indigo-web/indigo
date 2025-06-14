package buffer

// Buffer is a giant slice of data you write into it. Serves primarily the purpose of a quasi-arena
// by hosting non-interrelated byte sequences in a single place. Allows writing byte sequences streamingly.
type Buffer struct {
	memory  []byte
	begin   int
	maxSize int
}

func New(initialSize, maxSize int) Buffer {
	return Buffer{
		memory:  make([]byte, 0, initialSize),
		maxSize: maxSize,
	}
}

// Append writes data, checking whether the new amount of elements (bytes) doesn't exceed the
// limit, otherwise discarding the data and returning false.
func (b *Buffer) Append(elements []byte) (ok bool) {
	if len(b.memory)+len(elements) > b.maxSize {
		return false
	}

	b.memory = append(b.memory, elements...)
	return true
}

// AppendByte writes a single byte, checking whether it won't exceed the limit.
func (b *Buffer) AppendByte(c byte) (ok bool) {
	if len(b.memory)+1 > b.maxSize {
		return false
	}

	b.memory = append(b.memory, c)
	return true
}

// SegmentLength returns a number of bytes, taken by current segment, calculated as a difference
// between the beginning of the current segment and the current pointer.
func (b *Buffer) SegmentLength() int {
	return len(b.memory) - b.begin
}

// Trunc truncates the last n bytes from the current segment, guarantying that data of previous
// segments stays intact.
func (b *Buffer) Trunc(n int) {
	if seglen := b.SegmentLength(); n > seglen {
		n = seglen
	}

	b.memory = b.memory[:len(b.memory)-n]
}

// Discard discards current segment, and brings begin mark back by n bytes.
func (b *Buffer) Discard(n int) {
	if n > b.begin {
		n = b.begin
	}

	b.begin -= n
	b.memory = b.memory[:b.begin]
}

// Preview returns current segment without moving the head.
func (b *Buffer) Preview() []byte {
	return b.memory[b.begin:]
}

// Finish completes current segment, returning its value.
func (b *Buffer) Finish() []byte {
	segment := b.memory[b.begin:]
	b.begin = len(b.memory)

	return segment
}

// Clear just resets the pointers, so old values may be overridden by new ones.
func (b *Buffer) Clear() {
	b.begin = 0
	b.memory = b.memory[:0]
}
