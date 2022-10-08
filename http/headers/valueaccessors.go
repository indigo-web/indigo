package headers

import "strings"

var qualitySubstr = ";q=0."

// ValueOf returns a value until first semicolon is met. Even if the value after semicolon
// is not a parameter, it will anyway be counted as a parameter
func ValueOf(str string) string {
	if index := strings.IndexByte(str, ';'); index != -1 {
		return str[:index]
	}

	return str
}

// QualityOf simply returns a value of quality-parameter as an uint8
func QualityOf(str string) uint8 {
	if index := strings.Index(str, qualitySubstr); index != -1 {
		return str[index+len(qualitySubstr)] - '0'
	}

	return 9
}

// ParamOf looks for a parameter in a value, and if found, returns a parameter value.
// In case parameter is not found, empty string is returned
func ParamOf(str, key, or string) string {
	if index := strings.Index(str, ";"+key+"="); index != -1 {
		valOffset := index + len(key) + 1 + 1

		return str[valOffset:getParamEnd(str[valOffset:])]
	}

	return or
}

func getParamEnd(str string) int {
	for i := range str {
		switch str[i] {
		case ',', ';':
			return i
		}
	}

	return len(str)
}
