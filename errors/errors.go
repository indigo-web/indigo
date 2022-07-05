package errors

import "errors"

type Error error

var (
	ErrDuplicatedHeader = errors.New("found a duplicated header")
)
