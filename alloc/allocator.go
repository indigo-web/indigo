package alloc

// Allocator is simply a container with a big byte-slice inside. It is responsible
// for keeping header values together instead of having a lot of smaller slices.
// This decreases a GC pressure (having less pointers), and potentially makes code
// easier & doing less memory allocations
type Allocator struct {
	memory     []byte
	begin, pos int

	maxSize int
}

func NewAllocator(initialSpace, maxSpace int) Allocator {
	return Allocator{
		memory:  make([]byte, initialSpace),
		maxSize: maxSpace,
	}
}

// Append appends bytes to a buffer. In case space is not enough, trying to allocate a new
// by appending a fitting part into current buffer, and appending the rest, allowing a slice
// to grow by its in-built algorithm
func (a *Allocator) Append(chars []byte) (ok bool) {
	if a.pos+len(chars) > len(a.memory) {
		if len(a.memory)+len(chars) >= a.maxSize {
			return false
		}

		copy(a.memory[a.pos:], chars)
		a.memory = append(a.memory, chars[len(a.memory)-a.pos:]...)

		return true
	}

	copy(a.memory[a.pos:], chars)
	a.pos += len(chars)

	return true
}

func (a *Allocator) Finish() []byte {
	segment := a.memory[a.begin:a.pos]
	a.begin = a.pos

	return segment
}

func (a *Allocator) Clear() {
	a.begin = 0
	a.pos = 0
}
