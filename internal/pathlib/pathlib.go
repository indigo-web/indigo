package pathlib

import (
	"github.com/flrdv/uf"
)

type Path struct {
	buff                []byte
	buffLen             int
	maxPrefixLen        int
	prefix, replaceWith string
}

func NewPath(prefix, replaceWith string) *Path {
	prefix, replaceWith = withTrailingSep(prefix), withTrailingSep(replaceWith)

	return &Path{
		maxPrefixLen: max(len(prefix), len(replaceWith)),
		prefix:       prefix,
		replaceWith:  replaceWith,
	}
}

// Set copies the passed string into the internal buffer
func (p *Path) Set(path string) {
	offset := p.maxPrefixLen - len(p.prefix)
	if offset+len(path) > len(p.buff) {
		p.buff = make([]byte, offset+len(path))
	}

	p.buffLen = offset + copy(p.buff[offset:], path)
}

func (p *Path) Relative() string {
	offset := p.maxPrefixLen - len(p.replaceWith)
	copy(p.buff[offset:], p.replaceWith)

	return uf.B2S(p.buff[offset:p.buffLen])
}

// usingSeparator scans the path and looks for the first met separator.
// If no separator met, default ('/') will be returned
func usingSeparator(path string) byte {
	const defaultSeparator = '/'

	for i := range path {
		switch char := path[i]; char {
		case '/', '\\':
			return char
		}
	}

	return defaultSeparator
}

func withTrailingSep(path string) string {
	sep := usingSeparator(path)

	if path[len(path)-1] == sep {
		return path
	}

	return path + string(sep)
}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
