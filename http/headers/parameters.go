package headers

import (
	"github.com/indigo-web/indigo/internal/strutil"
	"iter"
)

// Params iterates over the parameters. An error is reported as the empty
// key-value pair (key="" and value=""). Such a pair is always the last one.
//
// Note: the passed value MUST be parameters only, otherwise the result depends
// on the header value itself (most probably error will be reported.)
func Params(headers Headers, key string) iter.Seq2[string, string] {
	params := strutil.CutParams(headers.Value(key))

	return strutil.WalkKV(params)
}
