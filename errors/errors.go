package errors

import "errors"

type Error error

// user errors, e.g. getting a header that is not presented in the request
var (
	ErrDuplicatedHeader = errors.New("found a duplicated header")
)
