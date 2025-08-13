package method

//go:generate stringer -type=Method
type Method uint8

const (
	Unknown Method = iota
	GET
	HEAD
	POST
	PUT
	DELETE
	CONNECT
	OPTIONS
	TRACE
	PATCH
	MKCOL
	MOVE
	COPY
	LOCK
	UNLOCK
	PROPFIND
	PROPPATCH

	// Count represents the maximal value an integer representation of a method can have.
	Count = iota - 1
)

// List enlists all known request methods, excluding Unknown.
var List = []Method{
	GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE, PATCH, MKCOL, MOVE, COPY, LOCK, UNLOCK, PROPFIND, PROPPATCH,
}

func Parse(str string) Method {
	switch len(str) {
	case 3:
		if str == "GET" {
			return GET
		} else if str == "PUT" {
			return PUT
		}
	case 4:
		if str == "POST" {
			return POST
		} else if str == "HEAD" {
			return HEAD
		} else if str == "MOVE" {
			return MOVE
		} else if str == "COPY" {
			return COPY
		} else if str == "LOCK" {
			return LOCK
		}
	case 5:
		if str == "PATCH" {
			return PATCH
		} else if str == "TRACE" {
			return TRACE
		} else if str == "MKCOL" {
			return MKCOL
		}
	case 6:
		if str == "DELETE" {
			return DELETE
		} else if str == "UNLOCK" {
			return UNLOCK
		}
	case 7:
		if str == "CONNECT" {
			return CONNECT
		} else if str == "OPTIONS" {
			return OPTIONS
		}
	case 8:
		if str == "PROPFIND" {
			return PROPFIND
		}
	case 9:
		if str == "PROPPATCH" {
			return PROPPATCH
		}
	}

	return Unknown
}
