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

func Parse(str string) Method {
	switch len(str) {
	case 3:
		if str == "GET" {
			return GET
		} else if str == "PUT" {
			return PUT
		}

		return Unknown
	case 4:
		if str == "POST" {
			return POST
		} else if str == "HEAD" {
			return HEAD
		}

		return Unknown
	case 5:
		if str == "PATCH" {
			return PATCH
		} else if str == "TRACE" {
			return TRACE
		}

		return Unknown
	case 6:
		if str == "DELETE" {
			return DELETE
		}

		return Unknown
	case 7:
		if str == "CONNECT" {
			return CONNECT
		} else if str == "OPTIONS" {
			return OPTIONS
		}

		return Unknown
	}

	return Unknown
}
