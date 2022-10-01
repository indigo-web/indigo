package radix

import (
	"errors"
)

type templateParserState uint8

const (
	eStatic templateParserState = iota + 1
	eSlash
	eDynamic
	eFinishDynamic
)

var (
	ErrNeedLeadingSlash = errors.New(
		"a leading slash is compulsory",
	)
	ErrInvalidPartName = errors.New(
		"slashes or figure braces are not allowed inside of the template part name",
	)
	ErrEmptyPath = errors.New(
		"template cannot be empty",
	)
	ErrMustEndWithSlash = errors.New(
		"a following slash is compulsory after the end of dynamic part",
	)
	ErrDynamicMustBeWholeSection = errors.New(
		"dynamic part must be a whole path section, without prefixes and suffixes",
	)
)

type Segment struct {
	IsDynamic bool
	Payload   string
}

// Template is a parsed template. It simply contains static parts, and marker names
// index of each corresponds to index of static part (first static part, then marker)
type Template struct {
	segments []Segment
}

func Parse(tmpl string) (Template, error) {
	var (
		offset   = 1
		template = Template{}
		state    = eStatic
	)

	if len(tmpl) == 0 {
		return template, ErrEmptyPath
	}

	for i, char := range tmpl {
		if i == 0 {
			if char != '/' {
				return template, ErrNeedLeadingSlash
			}

			// skip leading slash
			continue
		}

		switch state {
		case eStatic:
			switch char {
			case '/':
				template.segments = append(template.segments, Segment{
					IsDynamic: false,
					Payload:   tmpl[offset:i],
				})
				offset = i + 1
				state = eSlash
			case '{':
				return template, ErrDynamicMustBeWholeSection
			}
		case eSlash:
			switch char {
			case '/':
			case '{':
				offset = i + 1
				state = eDynamic
			default:
				state = eStatic
			}
		case eDynamic:
			switch char {
			case '}':
				template.segments = append(template.segments, Segment{
					IsDynamic: true,
					Payload:   tmpl[offset:i],
				})
				state = eFinishDynamic
			case '/', '{':
				return template, ErrInvalidPartName
			}
		case eFinishDynamic:
			switch char {
			case '/':
				offset = i + 1
				state = eSlash
			default:
				return template, ErrMustEndWithSlash
			}
		}
	}

	if state == eStatic && offset < len(tmpl)-1 {
		template.segments = append(template.segments, Segment{
			IsDynamic: false,
			Payload:   tmpl[offset:],
		})
	}

	return template, nil
}

func MustParse(tmpl string) Template {
	template, err := Parse(tmpl)
	if err != nil {
		panic(err.Error())
	}

	return template
}
