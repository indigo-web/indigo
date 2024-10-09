package headers

import (
	"github.com/indigo-web/indigo/internal/strutil"
	"iter"
	"strings"
)

// CutParams behaves exactly as strings.Cut, but strips whitespaces between value
// and the first-encountered parameter in addition.
func CutParams(header string) (value, params string) {
	sep := strings.IndexByte(header, ';')
	if sep == -1 {
		return header, ""
	}

	return header[:sep], strutil.LStripWS(header[sep+1:])
}

// WalkParams iterates over the parameters. An error is reported as the empty
// key-value pair (key="" and value=""). Such a pair is always the last one.
//
// Note: the passed value MUST be parameters only, otherwise the result depends
// on the header value itself (most probably error will be reported.)
func WalkParams(params string) iter.Seq2[string, string] {
	// TODO: must decode keys and values here. The problem is, we cannot do that directly
	// TODO: as this will mutate the buffer directly, which may cause issues and unnecessary
	// TODO: questions when calling request.Body.Bytes() afterwards. Therefore, in order to
	// TODO: avoid holding buffers when we don't need them, we can implement a lazy decoder
	// TODO: which will report if any urlencoded characters were met. So by that, a buffer
	// TODO: will be used only on demand, which seems fairly enough considering that this
	// TODO: is not a (much) likely path.
	return strutil.WalkKV(params, ';')
}
