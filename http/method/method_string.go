// Code generated by "stringer -type=Method"; DO NOT EDIT.

package method

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Unknown-0]
	_ = x[GET-1]
	_ = x[HEAD-2]
	_ = x[POST-3]
	_ = x[PUT-4]
	_ = x[DELETE-5]
	_ = x[CONNECT-6]
	_ = x[OPTIONS-7]
	_ = x[TRACE-8]
	_ = x[PATCH-9]
}

const _Method_name = "UnknownGETHEADPOSTPUTDELETECONNECTOPTIONSTRACEPATCH"

var _Method_index = [...]uint8{0, 7, 10, 14, 18, 21, 27, 34, 41, 46, 51}

func (i Method) String() string {
	if i >= Method(len(_Method_index)-1) {
		return "Method(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Method_name[_Method_index[i]:_Method_index[i+1]]
}
