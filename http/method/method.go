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

	// Count is the last one enum, so contains the greatest integer value of all the
	// methods. So real number of methods is lower by 1
	Count = iota - 1
)

// List contains all the supported HTTP methods. They are sorted by their integer value, however
// Unknown method is not included. So in order to index the List, you must subtract 1 first.
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
