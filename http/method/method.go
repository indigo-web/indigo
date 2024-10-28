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

	// Count represents the maximal value an integer representation of a method can have.
	Count = iota - 1
)

// List enlists all known request methods, excluding Unknown.
var List = []Method{GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE, PATCH}

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
		}
	case 5:
		if str == "PATCH" {
			return PATCH
		} else if str == "TRACE" {
			return TRACE
		}
	case 6:
		if str == "DELETE" {
			return DELETE
		}
	case 7:
		if str == "CONNECT" {
			return CONNECT
		} else if str == "OPTIONS" {
			return OPTIONS
		}
	}

	return Unknown
}
