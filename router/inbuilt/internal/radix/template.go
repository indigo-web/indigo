package radix

import (
	"errors"
	"fmt"
)

type templateParserState uint8

const (
	eStatic templateParserState = iota + 1
	eSlash
	eDynamic
	eFinishDynamic
)

var ErrEmptyPath = errors.New("template cannot be empty")

type Segment struct {
	Payload    string
	IsWildcard bool
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

	if tmpl[0] != '/' {
		return template, fmt.Errorf(`"%s": a leading slash is required`, tmpl)
	}

	for i := 1; i < len(tmpl); i++ {
		switch state {
		case eStatic:
			switch tmpl[i] {
			case '/':
				template.segments = append(template.segments, Segment{
					IsWildcard: false,
					Payload:    tmpl[offset:i],
				})
				offset = i + 1
				state = eSlash
			case '{':
				return template, fmt.Errorf(
					`"%s": dynamic part must be a whole path section, without prefixes and suffixes`,
					tmpl,
				)
			}
		case eSlash:
			switch tmpl[i] {
			case '/':
			case '{':
				offset = i + 1
				state = eDynamic
			default:
				state = eStatic
			}
		case eDynamic:
			switch tmpl[i] {
			case '}':
				template.segments = append(template.segments, Segment{
					IsWildcard: true,
					Payload:    tmpl[offset:i],
				})
				state = eFinishDynamic
			case '/', '{':
				return template, fmt.Errorf(
					`"%s": slashes or figure braces are not allowed inside of the template part name`,
					tmpl,
				)
			}
		case eFinishDynamic:
			switch tmpl[i] {
			case '/':
				offset = i + 1
				state = eSlash
			default:
				return template, fmt.Errorf(
					`"%s": dynamic part must be a whole path section, without prefixes and suffixes`,
					tmpl,
				)
			}
		}
	}

	if state == eStatic && offset < len(tmpl)-1 {
		template.segments = append(template.segments, Segment{
			IsWildcard: false,
			Payload:    tmpl[offset:],
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

// IsStatic tells whether the template contains any of wildcards
func (t Template) IsStatic() bool {
	for _, segment := range t.segments {
		if segment.IsWildcard {
			return false
		}
	}

	return true
}
