package headers

import (
	"strconv"
	"strings"
)

const DefaultEncoding = "iso-8859-1"

var (
	qualitySubstr = ";q=0."
	charsetSubstr = ";charset="
)

// ValueOf returns a value until first semicolon is met. Even if the value after semicolon
// is not a parameter, it will anyway be counted as a parameter
func ValueOf(str string) string {
	if index := strings.IndexByte(str, ';'); index != -1 {
		return str[:index]
	}

	return str
}

// QualityOf simply returns a value of quality-parameter as an uint8. If not presented or
// is not a valid integer, 9 is returned
func QualityOf(str string) int {
	q, err := strconv.Atoi(getParam(str, qualitySubstr, "9"))
	if err != nil {
		return 9
	}

	return q
}

// CharsetOf returns a charset parameter value, returning DefaultEncoding in case not presented
func CharsetOf(str string) string {
	return getParam(str, charsetSubstr, DefaultEncoding)
}

// ParamOf looks for a parameter in a value, and if found, returns a parameter value.
// In case parameter is not found, empty string is returned
func ParamOf(str, key string) string {
	return ParamOfOr(str, key, "")
}

func ParamOfOr(str, key, or string) string {
	return getParam(str, ";"+key+"=", or)
}

func getParam(str, substr, or string) string {
	if index := strings.Index(str, substr); index != -1 {
		valOffset := index + len(substr)

		return str[valOffset : valOffset+getParamEnd(str[valOffset:])]
	}

	return or
}

func getParamEnd(str string) int {
	if index := strings.IndexByte(str, ';'); index != -1 {
		return index
	}

	return len(str)
}
