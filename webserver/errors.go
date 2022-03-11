package webserver

import (
	"errors"
)

var (
	ErrHeaderNotFound   = errors.New("expected header not found")
	ErrDuplicatedHeader = errors.New("found a duplicated header")
)
