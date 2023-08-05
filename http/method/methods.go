package method

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
	Count
)

func (m Method) String() string {
	return ToString(m)
}

type entry struct {
	Method Method
	Origin string
}

func newMethodsMap(methods ...Method) (mmap [256][256]entry) {
	for _, method := range methods {
		str := ToString(method)
		mmap[str[0]][str[1]] = entry{
			Method: method,
			Origin: str,
		}
	}

	return mmap
}

var methodsMap = newMethodsMap(GET, HEAD, POST, PUT, DELETE, CONNECT, OPTIONS, TRACE, PATCH)

func Parse(str string) Method {
	method := methodsMap[str[0]][str[1]]
	if method.Origin != str {
		return Unknown
	}

	return method.Method
}

func ToString(method Method) string {
	switch method {
	case GET:
		return "GET"
	case HEAD:
		return "HEAD"
	case POST:
		return "POST"
	case PUT:
		return "PUT"
	case DELETE:
		return "DELETE"
	case CONNECT:
		return "CONNECT"
	case OPTIONS:
		return "OPTIONS"
	case TRACE:
		return "TRACE"
	case PATCH:
		return "PATCH"
	}

	return "???"
}
